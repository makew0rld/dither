package dither

import (
	"math/rand"
)

// PixelMapper is a function that takes the coordinate and color of a pixel,
// and returns a new color. That new color does not need to be part of any
// palette.
//
// This is used for thresholding, random dithering, patterning, and
// ordered dithering - basically any dithering that can be applied to each pixel
// individually.
//
// The provided RGB values are in the linear RGB space, and the returned values
// must be as well. All dithering operations should be happening in this space
// anyway, so this is done as a convenience. The RGB values are in the range
// [0, 65535], and must be returned in the same range.
//
// It must be thread-safe, as it will be called concurrently.
type PixelMapper func(x, y int, r, g, b uint16) (uint16, uint16, uint16)

// RandomNoiseGrayscale returns a PixelMapper that adds random noise to the
// color before returning. This is the simplest form of dithering.
//
// Non-grayscale colors will be converted to grayscale before the noise is added.
//
// You must call rand.Seed before calling using the PixelMapper, otherwise the
// output will be the same each time. A simple way to initialize rand.Seed is:
//
//     rand.Seed(time.Now().UnixNano())
//
// The noise added to each channel will be randomly chosen from within the range
// of min (inclusive) and max (exclusive). To simplify things, you can consider
// valid color values to range from 0 to 1. This means if you wanted the noise to
// shift the color through 50% of the color space at most, the min and max would be
// -0.5 and 0.5.
//
// Statistically, -0.5 and 0.5 are the best values for random dithering, as they
// evenly dither colors. Using values closer to zero (like -0.2 and 0.2) will
// effectively reduce the contrast of the image, and values further from zero
// (like -0.7 and 0.7) will increase the contrast.
//
// Making the min and max different values, like using -0.2 and 0.5 will make
// the image brighter or darker. In that example, the image will become brighter,
// as the randomness is more likely to land on the positive side and increase the
// color value.
//
// If the noise puts the channel value too high or too low it will be clamped,
// not wrapped. Basically, don't worry about the values of your min and max
// distorting the image in an unexpected way.
func RandomNoiseGrayscale(min, max float32) PixelMapper {
	return PixelMapper(func(x, y int, r, g, b uint16) (uint16, uint16, uint16) {
		// These values were taken from Wikipedia:
		// https://en.wikipedia.org/wiki/Grayscale#Colorimetric_(perceptual_luminance-preserving)_conversion_to_grayscale
		// 0.2126, 0.7152, 0.0722
		// Then multiplied by 65535, to scale them for 16-bit color.
		// Note that 13933 + 46871 + 4732 = 65536
		//
		// Basically, this takes linear RGB and gives a linear gray.
		gray := (13933*uint32(r) + 46871*uint32(g) + 4732*uint32(b) + 1<<15) >> 16

		new := RoundClamp(float32(gray) + 65535.0*(rand.Float32()*(max-min)+min))
		return new, new, new
	})
}

// RandomNoiseRGB is like RandomNoiseGrayscale but it adds randomness in the
// R, G, and B channels. It should not be used when you want a grayscale output
// image, ie when your palette is grayscale.
//
// Most of the time you will want all the mins to be the same, and all the maxes
// to be the same.
//
// See RandomNoiseGrayscale for more details about values and how this function
// works.
func RandomNoiseRGB(minR, maxR, minG, maxG, minB, maxB float32) PixelMapper {
	return PixelMapper(func(x, y int, r, g, b uint16) (uint16, uint16, uint16) {
		return RoundClamp(float32(r) + 65535.0*(rand.Float32()*(maxR-minR)+minR)),
			RoundClamp(float32(g) + 65535.0*(rand.Float32()*(maxG-minG)+minG)),
			RoundClamp(float32(b) + 65535.0*(rand.Float32()*(maxB-minB)+minB))
	})
}

func log2(v uint) uint {
	// Sources:
	// https://graphics.stanford.edu/~seander/bithacks.html#IntegerLogObvious
	// https://stackoverflow.com/a/18139978/7361270

	var r uint
	v = v >> 1
	for v != 0 {
		r++
		v = v >> 1
	}
	return r
}

// bayerMatrix returns a Bayer matrix with the given dimensions. The returned
// matrix is not divided, you will need to divide it by x*y.
//
// The x and y dimensions must be powers of two.
func bayerMatrix(xdim, ydim uint) [][]uint {
	// Bit math algorithm is used to calculate each cell of matrix individually.
	// This allows for easy generation of non-square matrices, as long as side
	// lengths are powers of two.
	//
	// Source for this bit math algorithm:
	// https://bisqwit.iki.fi/story/howto/dither/jy/#Appendix%202ThresholdMatrix
	//
	// The second code example on that part of the page is what this was based off
	// of, the one that works for rectangular matrices.
	//
	// The code was re-implemented exactly and tested to make sure the results
	// are the same. No algorithmic changes were made. The only code change was
	// to create a 2D slice to store and return results.

	M := log2(xdim)
	L := log2(ydim)

	matrix := make([][]uint, ydim)

	for y := uint(0); y < ydim; y++ {
		matrix[y] = make([]uint, xdim)
		for x := uint(0); x < xdim; x++ {

			var v, offset uint
			xmask := M
			ymask := L

			if M == 0 || (M > L && L != 0) {
				xc := x ^ ((y << M) >> L)
				yc := y
				for bit := uint(0); bit < M+L; {
					ymask--
					v |= ((yc >> ymask) & 1) << bit
					bit++
					for offset += M; offset >= L; offset -= L {
						xmask--
						v |= ((xc >> xmask) & 1) << bit
						bit++
					}
				}
			} else {
				xc := x
				yc := y ^ ((x << L) >> M)
				for bit := uint(0); bit < M+L; {
					xmask--
					v |= ((xc >> xmask) & 1) << bit
					bit++
					for offset += L; offset >= M; offset -= M {
						ymask--
						v |= ((yc >> ymask) & 1) << bit
						bit++
					}
				}
			}

			matrix[y][x] = v
		}
	}
	return matrix
}

// convThresholdToAddition takes a value from a matrix usually used for thresholding,
// and returns a value that can be added to a color instead of thresholded.
//
// scale is the number that's multiplied at the end, usually you want this to be
// 65535 to scale to match the color value range. value is the cell of the matrix.
// max is the divisor of the cell value, usually this is the product of the matrix
// dimensions.
func convThresholdToAddition(scale float32, value uint, max uint) float32 {
	// See:
	// https://en.wikipedia.org/wiki/Ordered_dithering
	// https://en.wikipedia.org/wiki/Talk:Ordered_dithering#Sources

	// 0.50000006 is next possible float32 value after 0.5. This is to correct
	// a rounding error that occurs when the number is exactly 0.5, which results
	// in pure black being dithered when it should be left alone.
	return scale * (float32(value+1.0)/float32(max) - 0.50000006)
}

// Bayer returns a PixelMapper that applies a Bayer matrix with the specified size.
// Please read this entire documentation, and see my recommendations at the end,
// especially if you're dithering color images.
//
// First off, cache the result of this function. It's not trivial to generate,
// and it can be re-used or used concurrently with no issues.
//
// The provided dimensions of the bayer matrix can only be powers of 2, but they do
// not need to be the same. If they are not powers of two this function will panic.
//
// There are currently two exceptions to this, which come from hand-derived Bayer
// matrices by Joel Yliluoma: 5x3, 3x5, 3x3. As he notes, "they can have a visibly
// uneven look, and thus are rarely worth using".
//
// Source:
//     https://bisqwit.iki.fi/story/howto/dither/jy/#Appendix%202ThresholdMatrix
//
// strength should be in the range [-1, 1]. It is multiplied with 65535
// (the max color value), which is then multiplied with the matrix.
//
// You can use this to change the amount the matrix is applied to the image, the
// "strength" of the dithering matrix. Usually just keeping it at 1.0 is fine.
//
// The closer to zero stength is, the smaller the range of colors that will be
// dithered. Colors outside that range will just be quantized, and not have a Bayer matrix
// applied. To dither the entire color range, set it to 1.0.
//
// Why would you want to shrink the dither range? Well Bayer matrixes are fundamentally
// biased to making the image brighter, increasing the value in each channel. This means
// that there might be darker parts that would be better off just quantized to the darkest
// color in your palette, instead of made lighter and dithered. By shrinking the dither
// range, you dither the colors that are more in the "middle", and let the darker and
// lighter ones just get quantized.
//
// You might also want to reduce the strength to reduce noise in the image, as dithering
// doesn't produce smooth colored areas. Usually a value around 0.8 is good for this.
//
// You can also make strength negative. If you know already that your image is dark, and so
// you don't want it to be made bright, then this is a better approach then shrinking the
// dither range. A negative strength flips the bias of the Bayer matrix, making it biased
// towards making images darker. To dither the entire color range but inverted, set strength
// to -1.0.
//
// The closer to zero you get, the more similar the effect of the negative and positive
// strength become. This is because they are shrinking the dither range towards the same spot.
//
// At Bayer sizes above 4x4, the brightness bias mostly disappears, and the difference
// between strength being -1.0 vs 1.0 is not really noticeable. Decreasing it below 1.0 or
// or above -1.0 will still shrink the dithering range, but instead of fixing some bias,
// it will just increase the contrast of the image.
//
// Greater than 1 or less than -1  doesn't really make sense, so stay away from that range.
// It expands the range of the dithering outside the possible color range, so there won't be
// enough dithering patterns in the output image. The further from zero, the larger the range.
//
// Going away from zero is similar to reducing contrast. If you go too far from zero, the
// whole image becomes gray.
//
// RECOMMENDATIONS
//
// For grayscale output, I would recommend 1.0 for lighter images, or -1.0 for darker images.
// If you cannot know beforehand, you may want to decrease that value, to reduce the risk of
// making dark images really bright. Try staying between 0.5 and 1.0.
//
// If you're using a Bayer size larger than 4x4, just using 1.0 for strength should be fine
// for most kinds of grayscale images.
//
// Color images are different. The Bayer matrix's bias to brightness applies to each RGB
// channel, and so the color of the image can become quite distorted at 1.0 strength.
// Several sites I have seen recommend 0.64 strength (written as 256/4), and from my own
// testing this is often a good value for color images. Do not default to 1.0 for Bayer
// dithering of color images.
//
// Of course, experiment for yourself. And let me know if I'm wrong!
func Bayer(x, y uint, strength float32) PixelMapper {
	var matrix [][]uint

	if x == 0 || y == 0 {
		panic("dither: Bayer: neither x or y can be zero")
	}
	if x == 3 && y == 3 {
		matrix = [][]uint{
			{0, 5, 2},
			{3, 8, 7},
			{6, 1, 4},
		}
	} else if x == 5 && y == 3 {
		matrix = [][]uint{
			{0, 12, 7, 3, 9},
			{14, 8, 1, 5, 11},
			{6, 4, 10, 13, 2},
		}
	} else if x == 3 && y == 5 {
		matrix = [][]uint{
			{0, 14, 16},
			{12, 8, 4},
			{7, 1, 10},
			{3, 5, 13},
			{9, 11, 2},
		}
	} else if (x&(x-1)) == 0 && (y&(y-1)) == 0 {
		// Both are powers of two
		matrix = bayerMatrix(x, y)
	} else {
		// Neither are powers of two
		panic("dither: Bayer: dimensions aren't both a power of two")
	}

	// Create precalculated matrix
	scale := 65535.0 * strength
	max := x * y

	precalc := make([][]float32, y)
	for i := uint(0); i < y; i++ {
		precalc[i] = make([]float32, x)
		for j := uint(0); j < x; j++ {
			precalc[i][j] = convThresholdToAddition(scale, matrix[i][j], max)
		}
	}

	return PixelMapper(func(xx, yy int, r, g, b uint16) (uint16, uint16, uint16) {
		return RoundClamp(float32(r) + precalc[yy%int(y)][xx%int(x)]),
			RoundClamp(float32(g) + precalc[yy%int(y)][xx%int(x)]),
			RoundClamp(float32(b) + precalc[yy%int(y)][xx%int(x)])
	})
}

// PixelMapperFromMatrix takes an OrderedDitherMatrix, and will return
// a PixelMapper. This is a simple way to make use of the clustered-dot matrices
// in this library, or to try out some matrix you found online.
//
// Because a PixelMapper is returned, this can make the matrix usable in more
// situations than originally designed, like with color images and multi-color
// palettes.
//
// See Bayer for a detailed explanation of strength. You can use this to change the
// amount the matrix is applied to the image, and to reduce noise. Usually you'll
// just want to set it to 1.0.
func PixelMapperFromMatrix(odm OrderedDitherMatrix, strength float32) PixelMapper {
	ydim := len(odm.Matrix)
	xdim := len(odm.Matrix[0])
	scale := 65535.0 * strength

	// Create precalculated matrix
	precalc := make([][]float32, ydim)
	for i := 0; i < ydim; i++ {
		precalc[i] = make([]float32, xdim)
		for j := 0; j < xdim; j++ {
			precalc[i][j] = convThresholdToAddition(scale, odm.Matrix[i][j], odm.Max)
		}
	}

	return PixelMapper(func(xx, yy int, r, g, b uint16) (uint16, uint16, uint16) {
		return RoundClamp(float32(r) + precalc[yy%ydim][xx%xdim]),
			RoundClamp(float32(g) + precalc[yy%ydim][xx%xdim]),
			RoundClamp(float32(b) + precalc[yy%ydim][xx%xdim])
	})
}
