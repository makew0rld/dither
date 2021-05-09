package dither

import (
	"image"
	"image/color"
	_ "image/jpeg"
	"image/png"
	"os"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

var (
	blackWhite    = []color.Color{color.Black, color.White}
	redGreenBlack = []color.Color{
		color.RGBA{255, 0, 0, 255},
		color.RGBA{0, 255, 0, 255},
		color.Black,
	}
	redGreenYellowBlack = []color.Color{
		color.RGBA{255, 0, 0, 255},
		color.RGBA{0, 255, 0, 255},
		color.RGBA{255, 255, 0, 255},
		color.Black,
	}
)

const (
	gradient = "images/input/gradient.png"
	peppers  = "images/input/peppers.png"
	dice     = "images/input/dice.png"
)

func ditherAndCompareImage(input string, expected string, d *Ditherer, t *testing.T) {
	expected = "images/output/" + expected

	f, err := os.Open(input)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		t.Fatal(err)
	}

	img = d.Dither(img)

	f2, err := os.Open(expected)
	if err != nil {
		t.Fatal(err)
	}
	defer f2.Close()

	img2, _, err := image.Decode(f2)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(img, img2) {
		t.Error("expected image and dithered image are not equal")
	}
}

func createDitheredImage(input, output string, d *Ditherer, t *testing.T) {
	output = "images/output/" + output

	f, err := os.Open(input)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		t.Fatal(err)
	}

	img = d.Dither(img)

	f2, err := os.Create(output)
	if err != nil {
		t.Fatal(err)
	}
	defer f2.Close()

	png.Encode(f2, img)
	t.Log("Output: " + output)
}

func TestRandomNoiseGrayscale(t *testing.T) {
	d := NewDitherer(blackWhite)
	d.Mapper = RandomNoiseGrayscale(-0.5, 0.5)
	d.SingleThreaded = true
	ditherAndCompareImage(gradient, "random_noise_grayscale.png", d, t)
}

func TestRandomNoiseRGB(t *testing.T) {
	noise := RandomNoiseRGB(-0.5, 0.5, -0.5, 0.5, -0.5, 0.5)

	d := NewDitherer(redGreenBlack)
	d.Mapper = noise
	d.SingleThreaded = true
	ditherAndCompareImage(peppers, "random_noise_rgb_red-green-black.png", d, t)

	d = NewDitherer(redGreenYellowBlack)
	d.Mapper = noise
	d.SingleThreaded = true
	ditherAndCompareImage(peppers, "random_noise_rgb_red-green-yellow-black.png", d, t)
}

func TestBayerMatrix(t *testing.T) {
	// Source for test cases is the same place as the original algorithm code
	// https://bisqwit.iki.fi/story/howto/dither/jy/#Appendix%202ThresholdMatrix

	t2x2 := [][]uint{
		{0, 3},
		{2, 1},
	}
	t4x4 := [][]uint{
		{0, 12, 3, 15},
		{8, 4, 11, 7},
		{2, 14, 1, 13},
		{10, 6, 9, 5},
	}
	t4x2 := [][]uint{
		{0, 4, 2, 6},
		{3, 7, 1, 5},
	}
	t2x4 := [][]uint{
		{0, 3},
		{4, 7},
		{2, 1},
		{6, 5},
	}

	assert.Equal(t, t2x2, bayerMatrix(2, 2))
	assert.Equal(t, t4x4, bayerMatrix(4, 4))
	assert.Equal(t, t4x2, bayerMatrix(4, 2))
	assert.Equal(t, t2x4, bayerMatrix(2, 4))
}

func TestBayerGrayscale(t *testing.T) {
	strength := float32(1.0)
	d := NewDitherer(blackWhite)

	d.Mapper = Bayer(2, 2, strength)
	ditherAndCompareImage(gradient, "bayer_2x2_gradient.png", d, t)

	d.Mapper = Bayer(4, 4, strength)
	ditherAndCompareImage(gradient, "bayer_4x4_gradient.png", d, t)

	d.Mapper = Bayer(8, 8, strength)
	ditherAndCompareImage(gradient, "bayer_8x8_gradient.png", d, t)

	d.Mapper = Bayer(16, 16, strength)
	ditherAndCompareImage(gradient, "bayer_16x16_gradient.png", d, t)

	d.Mapper = Bayer(16, 8, strength)
	ditherAndCompareImage(gradient, "bayer_16x8_gradient.png", d, t)
}

func TestBayerColor(t *testing.T) {
	bayer16 := Bayer(16, 16, 1.0)

	d := NewDitherer(redGreenBlack)
	d.Mapper = bayer16
	ditherAndCompareImage(peppers, "bayer_16x16_red-green-black.png", d, t)

	d = NewDitherer(redGreenYellowBlack)
	d.Mapper = bayer16
	ditherAndCompareImage(peppers, "bayer_16x16_red-green-yellow-black.png", d, t)
}

func TestErrorDiffusionMatrix(t *testing.T) {
	assert.Equal(t, 0, Simple2D.CurrentPixel())
	assert.Equal(t, 2, JarvisJudiceNinke.CurrentPixel())
}

func TestErrorDiffusionGrayscale(t *testing.T) {
	d := NewDitherer(blackWhite)

	d.Matrix = Simple2D
	ditherAndCompareImage(gradient, "edm_simple2d.png", d, t)

	d.Matrix = FloydSteinberg
	ditherAndCompareImage(gradient, "edm_floyd-steinberg.png", d, t)

	d.Matrix = JarvisJudiceNinke
	ditherAndCompareImage(gradient, "edm_jarvis-judice-ninke.png", d, t)

	d.Matrix = Atkinson
	ditherAndCompareImage(gradient, "edm_atkinson.png", d, t)
}

func TestSerpentine(t *testing.T) {
	d := NewDitherer(blackWhite)
	d.Serpentine = true

	d.Matrix = Simple2D
	ditherAndCompareImage(gradient, "edm_simple2d_serpentine.png", d, t)

	d.Matrix = FloydSteinberg
	ditherAndCompareImage(gradient, "edm_floyd-steinberg_serpentine.png", d, t)
}

func TestErrorDiffusionStrength(t *testing.T) {
	d := NewDitherer(blackWhite)
	d.Matrix = ErrorDiffusionStrength(FloydSteinberg, 0.5)
	ditherAndCompareImage(gradient, "edm_floyd-steinberg_strength_02.png", d, t)
}

func TestErrorDiffusionColor(t *testing.T) {
	d := NewDitherer(redGreenBlack)

	d.Matrix = Simple2D
	ditherAndCompareImage(peppers, "edm_peppers_simpled2d_red-green-black.png", d, t)

	d.Matrix = FloydSteinberg
	ditherAndCompareImage(peppers, "edm_peppers_floyd-steinberg_red-green-black.png", d, t)

	d.Matrix = JarvisJudiceNinke
	ditherAndCompareImage(peppers, "edm_peppers_jarvis-judice-ninke_red-green-black.png", d, t)

	d.Matrix = Atkinson
	ditherAndCompareImage(peppers, "edm_peppers_atkinson_red-green-black.png", d, t)

	d = NewDitherer(redGreenYellowBlack)

	d.Matrix = Simple2D
	ditherAndCompareImage(peppers, "edm_peppers_simpled2d_red-green-yellow-black.png", d, t)

	d.Matrix = FloydSteinberg
	ditherAndCompareImage(peppers, "edm_peppers_floyd-steinberg_red-green-yellow-black.png", d, t)

	d.Matrix = JarvisJudiceNinke
	ditherAndCompareImage(peppers, "edm_peppers_jarvis-judice-ninke_red-green-yellow-black.png", d, t)

	d.Matrix = Atkinson
	ditherAndCompareImage(peppers, "edm_peppers_atkinson_red-green-yellow-black.png", d, t)
}

func TestSubset(t *testing.T) {
	assert.Equal(t, true, subset([]color.Color{color.Black}, blackWhite))
	assert.Equal(t, false, subset(blackWhite, []color.Color{color.Black}))
	assert.Equal(t, true, subset(redGreenBlack, redGreenYellowBlack))
}

func TestSamePalette(t *testing.T) {
	assert.Equal(t, true, samePalette(blackWhite, blackWhite))
	assert.Equal(t, true, samePalette([]color.Color{color.White, color.Black}, blackWhite))
	assert.Equal(t, false, samePalette(blackWhite, redGreenBlack))
}

func sameImage(img1 image.Image, img2 image.Image) bool {
	if !img1.Bounds().Eq(img2.Bounds()) {
		return false
	}
	b := img1.Bounds()
	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {
			if !sameColor(img1.At(x, y), img2.At(x, y)) {
				return false
			}
		}
	}
	return true
}

func TestDitherPaletted(t *testing.T) {
	// Test that the paletted image returned matches the image that would be
	// returned by Dither.

	f, err := os.Open(peppers)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()

	src, _, err := image.Decode(f)
	if err != nil {
		t.Fatal(err)
	}

	d := NewDitherer(redGreenYellowBlack)
	d.Matrix = Simple2D // Whatever

	rgba := d.DitherCopy(src)
	pi := d.DitherPaletted(src)

	if !sameImage(rgba, pi) {
		t.Error("DitherPaletted output pixels are not the same as Dither")
	}
}

func TestPixelMapperFromMatrix(t *testing.T) {
	d := NewDitherer(blackWhite)

	d.Mapper = PixelMapperFromMatrix(ClusteredDot4x4, 1.0)
	ditherAndCompareImage(gradient, "ClusteredDot4x4.png", d, t)
	d.Mapper = PixelMapperFromMatrix(ClusteredDotDiagonal8x8, 1.0)
	ditherAndCompareImage(gradient, "ClusteredDotDiagonal8x8.png", d, t)
	d.Mapper = PixelMapperFromMatrix(Vertical5x3, 1.0)
	ditherAndCompareImage(gradient, "Vertical5x3.png", d, t)
	d.Mapper = PixelMapperFromMatrix(Horizontal3x5, 1.0)
	ditherAndCompareImage(gradient, "Horizontal3x5.png", d, t)
	d.Mapper = PixelMapperFromMatrix(ClusteredDotDiagonal6x6, 1.0)
	ditherAndCompareImage(gradient, "ClusteredDotDiagonal6x6.png", d, t)
	d.Mapper = PixelMapperFromMatrix(ClusteredDotDiagonal8x8_2, 1.0)
	ditherAndCompareImage(gradient, "ClusteredDotDiagonal8x8_2.png", d, t)
	d.Mapper = PixelMapperFromMatrix(ClusteredDotDiagonal16x16, 1.0)
	ditherAndCompareImage(gradient, "ClusteredDotDiagonal16x16_gradient.png", d, t)
	d.Mapper = PixelMapperFromMatrix(ClusteredDot6x6, 1.0)
	ditherAndCompareImage(gradient, "ClusteredDot6x6.png", d, t)
	d.Mapper = PixelMapperFromMatrix(ClusteredDotSpiral5x5, 1.0)
	ditherAndCompareImage(gradient, "ClusteredDotSpiral5x5.png", d, t)
	d.Mapper = PixelMapperFromMatrix(ClusteredDotHorizontalLine, 1.0)
	ditherAndCompareImage(gradient, "ClusteredDotHorizontalLine.png", d, t)
	d.Mapper = PixelMapperFromMatrix(ClusteredDotVerticalLine, 1.0)
	ditherAndCompareImage(gradient, "ClusteredDotVerticalLine.png", d, t)
	d.Mapper = PixelMapperFromMatrix(ClusteredDot8x8, 1.0)
	ditherAndCompareImage(gradient, "ClusteredDot8x8.png", d, t)
	d.Mapper = PixelMapperFromMatrix(ClusteredDot6x6_2, 1.0)
	ditherAndCompareImage(gradient, "ClusteredDot6x6_2.png", d, t)
	d.Mapper = PixelMapperFromMatrix(ClusteredDot6x6_3, 1.0)
	ditherAndCompareImage(gradient, "ClusteredDot6x6_3.png", d, t)
	d.Mapper = PixelMapperFromMatrix(ClusteredDotDiagonal8x8_3, 1.0)
	ditherAndCompareImage(gradient, "ClusteredDotDiagonal8x8_3.png", d, t)
}

func TestAlpha(t *testing.T) {
	d := NewDitherer([]color.Color{
		color.Black,
		color.White,
		color.RGBA{255, 0, 0, 255},
		color.RGBA{0, 255, 0, 255},
		color.RGBA{0, 0, 255, 255},
	})
	d.Mapper = Bayer(4, 4, 1)

	ditherAndCompareImage(dice, "alpha_bayer.png", d, t)

	d.Mapper = nil
	d.Matrix = FloydSteinberg

	ditherAndCompareImage(dice, "alpha_floyd-steinberg.png", d, t)
}

// func TestDrawer(t *testing.T) {
// 	palette := []color.Color{
// 		color.Gray{Y: 255},
// 		color.Gray{Y: 160},
// 		color.Gray{Y: 70},
// 		color.Gray{Y: 35},
// 		color.Gray{Y: 0},
// 	}
// 	d := NewDitherer(palette)
// 	d.Matrix = FloydSteinberg

// 	// Modified from Go stdlib:
// 	// https://github.com/golang/go/blob/go1.15.7/src/image/draw/example_test.go
// 	const width = 130
// 	const height = 50

// 	im := image.NewGray(image.Rectangle{Max: image.Point{X: width, Y: height}})
// 	for x := 0; x < width; x++ {
// 		for y := 0; y < height; y++ {
// 			dist := math.Sqrt(math.Pow(float64(x-width/2), 2)/3+math.Pow(float64(y-height/2), 2)) / (height / 1.5) * 255
// 			var gray uint8
// 			if dist > 255 {
// 				gray = 255
// 			} else {
// 				gray = uint8(dist)
// 			}
// 			im.SetGray(x, y, color.Gray{Y: 255 - gray})
// 		}
// 	}
// 	pi := image.NewPaletted(im.Bounds(), palette)

// 	d.Draw(pi, im.Bounds(), im, image.ZP)
// 	shade := []string{" ", "░", "▒", "▓", "█"}
// 	for i, p := range pi.Pix {
// 		fmt.Print(shade[p])
// 		if (i+1)%width == 0 {
// 			fmt.Print("\n")
// 		}
// 	}
// }
