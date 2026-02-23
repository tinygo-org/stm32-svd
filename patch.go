package main

import (
	"bufio"
	"bytes"
	"fmt"
	"iter"
	"os"
	"slices"
	"strings"
)

// This file adjusts some of the patch files in stm32-rs/devices/**
// It doesn't apply modifications twice, so it can be re-run safely.

var patchFiles16Bit = []string{
	// patches/16bit*, to keep peripheral registers of
	// IWDG, WWDG, SPI1 or I2C accessible as 32-bit regs
	"stm32-rs/devices/patches/16bit.yaml",
	"stm32-rs/devices/patches/16bit_with_mask.yaml",

	// USART registers, which appear to have size 32 bits; see RM0008, p. 818
	// https://www.st.com/resource/en/reference_manual/rm0008-stm32f101xx-stm32f102xx-stm	32f103xx-stm32f105xx-and-stm32f107xx-advanced-armbased-32bit-mcus-stmicroelectronics.pdf
	"stm32-rs/devices/patches/usart/v1.yaml",

	// added by @9887ffa on 2025-12-16 "TIM v1 16-bit regs"
	"stm32-rs/devices/patches/tim/v1/16bit.yaml",
	"stm32-rs/devices/patches/tim/v1/tim1_16bit.yaml",
	"stm32-rs/devices/patches/tim/v1/tim2_16bit.yaml",

	// added by @3708bbb on 2025-02-24 "v2 start"
	"stm32-rs/devices/patches/tim/v2/arr_16bit.yaml",
	"stm32-rs/devices/patches/tim/v2/tim1.yaml",
	"stm32-rs/devices/patches/tim/v2/tim10.yaml",
	"stm32-rs/devices/patches/tim/v2/tim13.yaml",
	"stm32-rs/devices/patches/tim/v2/tim15.yaml",
	"stm32-rs/devices/patches/tim/v2/tim16.yaml",
	"stm32-rs/devices/patches/tim/v2/tim2.yaml",
	"stm32-rs/devices/patches/tim/v2/tim3.yaml",
	"stm32-rs/devices/patches/tim/v2/tim3_with_mask.yaml",
	"stm32-rs/devices/patches/tim/v2/tim6.yaml",
	"stm32-rs/devices/patches/tim/v2/tim9.yaml",
}

func main() {
	for _, filename := range patchFiles16Bit {
		deact16BitConstr(filename)
	}

	// remove 16-bit constraint of USART.GTPR
	// it is a 32 bit register, with upper 16 bit reserved; see RM0090, p. 1020
	// https://www.st.com/resource/en/reference_manual/dm00031020-stm32f405-415-stm32f407-417-stm32f427-437-and-stm32f429-439-advanced-arm-based-32-bit-mcus-stmicroelectronics.pdf#page=1020
	//
	deact16BitConstr("stm32-rs/devices/patches/usart/f4_add_UART_GTPR.yaml")

	// stm32g071: IWDG, SPI1: remove 16-bit constraints
	//
	deact16BitConstr("stm32-rs/devices/stm32g071.yaml", "IWDG", "SPI1")

	enforceSize32("stm32-rs/devices/stm32l4r5.yaml", "TIM[18]")

	os.Exit(exitCode)
}

// In stm32-rs, some original SVDs already come with
// timer registers set to 16 bits, which won't be changed by
// the stm32-rs patches.
// EnforceSize32 adds modifications to the specified sections
// that set the size to 32 bits.
func enforceSize32(filename, section string) {
	for y := range yamlLines(filename, section) {
		if y.text == "_include:" {
			y.insertBefore = `_modify:
  "?*":
    size: 32`
		}
	}
}

// deact16BitConstr deactivates a 16-bit constraints from registers;
// sections can be used to narrow operation to specific sections of the
// YAML file, like "SPI1".
func deact16BitConstr(filename string, sections ...string) {
	for y := range yamlLines(filename, sections...) {
		if y.level < 2 {
			continue
		}
		if y.text == "size: 16" {
			y.deactivate = true
		}
	}
}

// The section below this line contains the parser implementation.

type yamlLine struct {
	level int
	text  string

	// edit actions
	deactivate     bool
	insertBefore   string
	alreadyApplied bool

	skipSection bool
}

func yamlLines(filename string, sections ...string) iter.Seq[*yamlLine] {
	f, err := os.OpenFile(filename, os.O_RDWR, 0644)
	if err != nil {
		setError(filename, err)
		return nil
	}
	s := bufio.NewScanner(f)

	return func(yield func(y *yamlLine) bool) {
		defer f.Close()
		var y yamlLine
		b := new(bytes.Buffer)
		numChanges := 0
		numChangesWant := max(len(sections), 1)
		skipToEnd := false
		numSeen := 0
		for s.Scan() {
			line := s.Text()
			l := 0
			for _, c := range line {
				if c != ' ' {
					break
				}
				l++
			}
			y.text = line[l:]
			remaining, match := matchesSection(y.text, sections)
			if match {
				sections = remaining
			}
			switch {
			case skipToEnd:

			case y.text == "":

			case l == 0 && !match:
				y.skipSection = true

			case y.skipSection && l > 0:

			case y.text == "# tinygo inserted":
				numSeen++
				y.skipSection = true
				break
			case strings.HasPrefix(y.text, "tinygo_disabled_"):
				numSeen++
				y.skipSection = true
				break
			default:
				y.skipSection = false
				y.level = l
				y.deactivate = false

				if !yield(&y) {
					skipToEnd = true
					break
				}
				indent := line[:y.level]
				if y.insertBefore != "" {
					fmt.Fprintln(b, indent+"# tinygo inserted")
					for _, s := range strings.Split(y.insertBefore, "\n") {
						fmt.Fprintln(b, indent+s)
					}
					y.insertBefore = ""
					numChanges++
				}
				if y.deactivate {
					line = indent + "tinygo_disabled_" + line[y.level:]
					numChanges++
				}
			}
			fmt.Fprintln(b, line)
		}
		if numChanges < numChangesWant {
			if numSeen < numChangesWant {
				if len(sections) > 0 {
					setError(filename, fmt.Errorf("sections not found: %q", strings.Join(sections, ", ")))
				} else {
					setError(filename, fmt.Errorf("could not apply changes"))
				}
			}
			return
		}
		f.Truncate(0)
		f.Seek(0, 0)
		_, err := b.WriteTo(f)
		if err != nil {
			setError(filename, err)
			return
		}
		fmt.Println(filename+":", numChanges)
	}
}

func matchesSection(name string, sections []string) ([]string, bool) {
	if name == "" {
		return sections, false
	}
	if sections == nil {
		return nil, true
	}
	if len(sections) == 0 {
		return sections, false
	}
	name, hasColon := strings.CutSuffix(name, ":")
	if !hasColon {
		return sections, false
	}
	i := slices.Index(sections, name)
	if i == -1 {
		return sections, false
	}
	return slices.Delete(sections, i, i+1), true
}

var exitCode = 0

func setError(filename string, err error) {
	fmt.Fprintf(os.Stderr, "%s: %v\n", filename, err)
	exitCode = 1
}
