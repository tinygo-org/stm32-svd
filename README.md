# Patched SVD files for STM32 MCUs

This repository contains SVD files for most STM32 chips, generated from the [stm32-rs](https://github.com/stm32-rs/stm32-rs), part of the [Rust Embedded](https://github.com/rust-embedded) community. These files are originally from [STMicroelectronics](https://www.st.com/) but have many patches to fix bugs and fill in gaps in the original SVD files.

These files are not intended to be modified. Instead, they are intended to be used as any SVD file, for example to generate access to registers for various languages.

## Contributing

Please do not contribute changes directly to the SVD files in this repository. Instead, controbute patches upstream in the [stm32-rs](https://github.com/stm32-rs/stm32-rs) repository.

## Updating

From time to time this repository will need to be updated, to incorporate changes from upstream. You can do so as follows:

 1. Make sure the stm32-rs submodule is pulled, using `git submodule update --init`.
 2. Download the latest patches by going to the stm32-rs subdirectory and running `git pull`.
 3. Run `make`.

## License

The SVD files in this repository have two sources: the original ST SVD files and the stm32-rs patch files.

  * The original ST SVD files are licensed under a license from ST, see [`ST_SLA.pdf`](ST_SLA.pdf) for details.
  * The patch files are dual licensed under either:
    - Apache License, Version 2.0 ([LICENSE-APACHE](LICENSE-APACHE) or http://www.apache.org/licenses/LICENSE-2.0)
    - MIT license ([LICENSE-MIT](LICENSE-MIT) or http://opensource.org/licenses/MIT)

See [stm32-rs#License](https://github.com/stm32-rs/stm32-rs#License) for details.
