package dither

// ErrorDiffusionMatrix holds the matrix for the error-diffusion type of dithering.
// An example of this would be Floyd-Steinberg or Atkinson.
//
// Zero values can be used to represent pixels that have already been processed.
// The current pixel is assumed to be the right-most zero value in the top row.
type ErrorDiffusionMatrix [][]float32

// CurrentPixel returns the index the current pixel.
// The current pixel is assumed to be the right-most zero value in the top row.
// In all matrixes that I have seen, the current pixel is always in the middle,
// but this function exists just in case.
//
// Therefore with an ErrorDiffusionMatrix named edm, the current pixel is at:
//     edm[0][edm.CurrentPixel()]
//
// Usually you'll want to cache this value.
func (e ErrorDiffusionMatrix) CurrentPixel() int {
	for i, v := range e[0] {
		if v != 0 {
			return i - 1
		}
	}
	// The whole first line is zeros, which doesn't make sense
	// Just default to returning the middle of the row.
	return len(e[0]) / 2
}

// Offset will take the index of where you are in the matrix and return the
// offset from the current pixel. You have to pass the curPx value yourself
// to allow for caching, but it can be retrieved by calling CurrentPixel().
func (e ErrorDiffusionMatrix) Offset(x, y, curPx int) (int, int) {
	return x - curPx, y
}

// ErrorDiffusionStrength modifies an existing error diffusion matrix so that it will
// be applied with the specified strength.
//
// strength is usually a value from 0 to 1.0, where 1.0 means 100% strength, and will
// not modify the matrix at all. It is inversely proportional to contrast - reducing the
// strength increases the contrast. It can be useful at values like 0.8 for reducing
// noise in the dithered image.
//
// See the documentation for Bayer for more details.
func ErrorDiffusionStrength(edm ErrorDiffusionMatrix, strength float32) ErrorDiffusionMatrix {
	if strength == 1 {
		return edm
	}

	dy := len(edm)
	dx := len(edm[0])
	edm2 := make(ErrorDiffusionMatrix, dy)
	for y := 0; y < dy; y++ {
		edm2[y] = make([]float32, dx)
		for x := 0; x < dx; x++ {
			edm2[y][x] = edm[y][x] * strength
		}
	}
	return edm2
}

var Simple2D = ErrorDiffusionMatrix{
	{0, 0.5},
	{0.5, 0},
}

var FloydSteinberg = ErrorDiffusionMatrix{
	{0, 0, 7.0 / 16},
	{3.0 / 16, 5.0 / 16, 1.0 / 16},
}

var FalseFloydSteinberg = ErrorDiffusionMatrix{
	{0, 3.0 / 8},
	{3.0 / 8, 2.0 / 8},
}

var JarvisJudiceNinke = ErrorDiffusionMatrix{
	{0, 0, 0, 7.0 / 48, 5.0 / 48},
	{3.0 / 48, 5.0 / 48, 7.0 / 48, 5.0 / 48, 3.0 / 48},
	{1.0 / 48, 3.0 / 48, 5.0 / 48, 3.0 / 48, 1.0 / 48},
}

var Atkinson = ErrorDiffusionMatrix{
	{0, 0, 1.0 / 8, 1.0 / 8},
	{1.0 / 8, 1.0 / 8, 1.0 / 8, 0},
	{0, 1.0 / 8, 0, 0},
}

var Stucki = ErrorDiffusionMatrix{
	{0, 0, 0, 8.0 / 42, 4.0 / 42},
	{2.0 / 42, 4.0 / 42, 8.0 / 42, 4.0 / 42, 2.0 / 42},
	{1.0 / 42, 2.0 / 42, 4.0 / 42, 2.0 / 42, 1.0 / 42},
}

var Burkes = ErrorDiffusionMatrix{
	{0, 0, 0, 8.0 / 32, 4.0 / 32},
	{2.0 / 32, 4.0 / 32, 8.0 / 32, 4.0 / 32, 2.0 / 32},
}

var Sierra = ErrorDiffusionMatrix{
	{0, 0, 0, 5.0 / 32, 3.0 / 32},
	{2.0 / 32, 4.0 / 32, 5.0 / 32, 4.0 / 32, 2.0 / 32},
	{0, 2.0 / 32, 3.0 / 32, 2.0 / 32, 0},
}

// Sierra3 is another name for the original Sierra matrix.
var Sierra3 = Sierra

var TwoRowSierra = ErrorDiffusionMatrix{
	{0, 0, 0, 4.0 / 16, 3.0 / 16},
	{1.0 / 16, 2.0 / 16, 3.0 / 16, 2.0 / 16, 1.0 / 16},
}

// Sierra2 is another name for TwoRowSierra
var Sierra2 = TwoRowSierra

var SierraLite = ErrorDiffusionMatrix{
	{0, 0, 2.0 / 4},
	{1.0 / 4, 1.0 / 4, 0},
}

// Sierra2_4A (usually written as Sierra2-4A) is another name for SierraLite.
var Sierra2_4A = SierraLite

// StevenPigeon is an error diffusion matrix developed by Steven Pigeon.
// Source: https://hbfs.wordpress.com/2013/12/31/dithering/
var StevenPigeon = ErrorDiffusionMatrix{
	{0, 0, 0, 2.0 / 14, 1.0 / 14},
	{0, 2.0 / 14, 2.0 / 14, 2.0 / 14, 0},
	{1.0 / 14, 0, 1.0 / 14, 0, 1.0 / 14},
}
