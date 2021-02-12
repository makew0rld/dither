package dither

// SpecialDither is used to represent dithering algorithms that require custom
// code, because they cannot be represented by a PixelMapper or error diffusion
// matrix.
//
// There are currently no SpecialDither options, but they will be added in the
// future.
type SpecialDither int
