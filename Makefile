
DEVICES := $(sort $(wildcard stm32-rs/devices/*.yaml))
SVD_IN := $(patsubst stm32-rs/devices/%.yaml,stm32-rs/svd/%.svd,$(DEVICES))
SVD_OUT := $(patsubst stm32-rs/devices/%.yaml,svd/%.svd,$(DEVICES))

files: $(SVD_OUT)

# Extract all source .svd files from the vendor supplied archive.
$(SVD_IN):
	cd stm32-rs/svd && ./extract.sh

# Patch the SVD file.
stm32-rs/svd/%.svd.patched: stm32-rs/devices/%.yaml stm32-rs/svd/%.svd
	svd patch $<

# Copy the resulting SVD file to the destination location.
svd/%.svd: stm32-rs/svd/%.svd.patched
	cp $< $@
