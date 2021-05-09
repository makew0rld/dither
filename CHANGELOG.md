# Changelog
All notable changes to this project will be documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.0.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).


## [2.2.0] - 2021-05-09
### Added
- Support for images with transparency (#8)


## [2.1.1] - 2021-04-30
### Changed
- Update Bayer strength recommendations for color images


## [2.1.0] - 2021-04-29
### Added
- JSON tags for `OrdereredDitherMatrix`

### Changed
- `Dither` never returns `nil`, making code simpler

### Fixed
- Bug where paletted images would never be detected as needing to be copied in `Dither`
- Palette is actually fully copied when needed, before the colors were shared with passed slice
  - `NewDitherer`
  - `Ditherer.GetColorPalette`
  - `DitherPaletted` and `DitherPalettedConfig`
  - `GetColorModel`


## [2.0.0] - 2021-02-13
### Added
- Added `ErrorDiffusionStrength` to set the strength of error diffusion dithering (#4)
- `RoundClamp` function for making your own `PixelMappers` that round correctly

### Changed
- All linear RGB values are represented using `uint16` instead of `uint8` now, because 8-bits is not enough to accurately hold a linearized value. This is a breaking change, hence the new major version.

### Fixed
- Rounding is no longer biased, because ties are rounded to the nearest even number


## [1.0.0] - 2021-02-11
Initial release.
