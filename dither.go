package dither

import (
	"image"
	"image/color"
	"image/draw"
	"math"
	"runtime"
)

// Ditherer dithers images according to the settings in the struct.
// It can be safely reused for many images, and used concurrently.
//
// Some members of the struct are public. Those members can be changed
// in-between dithering images, if you would like to dither again.
// If you change those public methods while an image is being dithered, the
// output image will have problems, so only change in-between dithering.
//
// You can only set one of Matrix, Mapper, or Special. Trying to dither when
// none or more than one of those are set will cause the function to panic.
type Ditherer struct {

	// Matrix is the ErrorDiffusionMatrix for dithering.
	Matrix ErrorDiffusionMatrix

	// Mapper is the ColorMapper function for dithering.
	Mapper PixelMapper

	// Special is the special dithering algorithm that's being used. The default
	// value of 0 indicates that no special dithering algorithm is being used.
	Special SpecialDither

	// SingleThreaded controls whether the dithering happens sequentially or using
	// runtime.GOMAXPROCS(0) workers, which defaults to the number of CPUs.
	//
	// Note that error diffusion dithering (using Matrix) is sequential by nature
	// and so this field has no effect.
	//
	// Setting this to true is only useful in rare cases, like when numbers are
	// used sequentially in a PixelMapper, and the output must be deterministic.
	// Because otherwise the numbers will be retrieved in a different order each
	// time, as the goroutines call on the PixelMapper.
	SingleThreaded bool

	// Serpentine controls whether the error diffusion matrix is applied in a
	// serpentine manner, meaning that it goes right-to-left every other line.
	// This greatly reduces line-type artifacts. If a Mapper is being used this
	// field will have no effect.
	Serpentine bool

	// palette holds the colors the dithered image is allowed to use, in the
	// sRGB color space.
	palette []color.Color

	// linearPalette holds all the palette colors, but in linear RGB space.
	linearPalette [][3]uint8
}

// NewDitherer creates a new Ditherer that uses a copy of the provided palette.
// If the palette is empty or nil then nil will be returned.
func NewDitherer(palette []color.Color) *Ditherer {
	if len(palette) == 0 {
		return nil
	}

	d := &Ditherer{}

	// Palette is copied so the user can't modify it externally later
	d.palette = make([]color.Color, len(palette))
	copy(d.palette, palette)

	// Create linear RGB version of the palette
	d.linearPalette = make([][3]uint8, len(d.palette))
	for i := range d.linearPalette {
		r, g, b := toLinearRGB(d.palette[i])
		d.linearPalette[i] = [3]uint8{r, g, b}
	}

	return d
}

// invalid returns true when the current struct fields of the Ditherer make it
// impossible to dither.
func (d *Ditherer) invalid() bool {
	// This basically XORs three bools that represent whether each value is
	// unset or not. The if statement evaluates to true if one is set, but
	// false if none or more than one are set. But then it's flipped with !()
	// on the outside.
	if !((d.Mapper != nil) != ((d.Matrix != nil) != (d.Special != 0))) {
		return true
	}
	if d.Special != 0 {
		// No special dithering supported right now
		return true
	}
	return false
}

// GetPalette returns a copy of the current palette being used by the Ditherer.
func (d *Ditherer) GetPalette() []color.Color {
	// Palette is copied so the user can't modify it externally later
	p := make([]color.Color, len(d.palette))
	copy(p, d.palette)
	return p
}

func sqDiff(v1 uint8, v2 uint8) uint16 {
	// This optimization is copied from Go stdlib, see
	// https://github.com/golang/go/blob/go1.15.7/src/image/color/color.go#L314

	d := uint16(v1) - uint16(v2)
	return (d * d) >> 2
}

// closestColor returns the index of the color in the palette that's closest to
// the provided one, using Euclidean distance in linear RGB space. The provided
// RGB values must be linear RGB.
func (d *Ditherer) closestColor(r, g, b uint8) int {
	// Go through each color and find the closest one
	color, best := 0, uint16(math.MaxUint16)
	for i, c := range d.linearPalette {
		// Euclidean distance, but the square root part is removed
		dist := sqDiff(r, c[0]) + sqDiff(g, c[1]) + sqDiff(b, c[2])

		if dist < best {
			if dist == 0 {
				return i
			}
			color, best = i, dist
		}
	}
	return color
}

// Dither dithers the provided image.
//
// It will always try to change the provided image and return nil, but if that
// is not possible it will return the dithered image as a copy.
//
// In comparison to DitherCopy, this can greatly reduce memory usage, and is quicker
// because it usually won't copy the image at the beginning. It should be preferred
// if you don't need to keep the original image.
//
// Cases where a copy will be returned include:
//
// If the input image is *image.Paletted and the image's palette is different than
// the Ditherer's; if the image can't be casted to draw.Image.
//
// The returned image type (when not nil) is always *image.RGBA.
func (d *Ditherer) Dither(src image.Image) image.Image {
	if d.invalid() {
		panic("dither: invalid Ditherer")
	}

	var img draw.Image
	var ret image.Image = nil

	if pi, ok := img.(*image.Paletted); ok {
		if !samePalette(d.palette, pi.Palette) {
			// Can't use this because it will change image colors
			// Instead make a copy, and return that later
			img = copyOfImage(src)
			ret = img
		}
	} else if img, ok = src.(draw.Image); !ok {
		// Can't be changed
		// Instead make a copy and dither and return that
		img = copyOfImage(src)
		ret = img
	}

	if d.Mapper != nil {
		workers := 1
		if !d.SingleThreaded {
			workers = runtime.GOMAXPROCS(0)
		}
		parallel(workers, img.(draw.Image), img, func(x, y int, c color.Color) color.Color {
			r, g, b := toLinearRGB(c)
			return d.palette[d.closestColor(d.Mapper(x, y, r, g, b))]
		})
		return ret
	}

	// Matrix needs to be applied instead

	b := img.Bounds()
	curPx := d.Matrix.CurrentPixel()

	// Store linear values here instead of converting back and forth and storing
	// sRGB values inside the image.
	// Pointers are used to differentiate between a zero value and an unset value
	lins := make([][][3]*uint8, b.Dy())
	for i := 0; i < len(lins); i++ {
		lins[i] = make([][3]*uint8, b.Dx())
	}

	// Setters and getters for that linear storage
	linearSet := func(x, y int, r, g, b uint8) {
		lins[y][x] = [3]*uint8{&r, &g, &b}
	}
	linearAt := func(x, y int) (uint8, uint8, uint8) {
		c := lins[y][x]
		if c[0] == nil {
			// This pixel hasn't been linearized yet
			r, g, b := toLinearRGB(img.At(x, y))
			linearSet(x, y, r, g, b)
			return r, g, b
		}
		return *c[0], *c[1], *c[2]
	}

	for y := b.Min.Y; y < b.Max.Y; y++ {
		for x := b.Min.X; x < b.Max.X; x++ {

			oldX := x
			if d.Serpentine && y%2 == 0 {
				// Reverse direction
				x = b.Max.X - 1 - x
			}

			// Quantize current pixel
			oldR, oldG, oldB := linearAt(x, y)
			newColorIdx := d.closestColor(oldR, oldG, oldB)
			img.Set(x, y, d.palette[newColorIdx])

			new := d.linearPalette[newColorIdx]
			// Quant errors in each channel
			er, eg, eb := int16(oldR)-int16(new[0]), int16(oldG)-int16(new[1]), int16(oldB)-int16(new[2])

			// Diffuse error in two dimensions
			for yy := range d.Matrix {
				for xx := range d.Matrix[yy] {
					// Get the coords of the pixel the error is being applied to
					deltaX, deltaY := d.Matrix.Offset(xx, yy, curPx)
					if d.Serpentine && y%2 == 0 {
						// Reflect the matrix horizontally because we're going right-to-left
						// Otherwise the matrix would change pixels that have already been set
						deltaX *= -1
					}
					pxX := x + deltaX
					pxY := y + deltaY

					if !(image.Point{pxX, pxY}.In(b)) {
						// This is outside the image, so don't bother doing any further calculations
						continue
					}

					r, g, b := linearAt(pxX, pxY)
					linearSet(pxX, pxY,
						// +0.5 is important to prevent pure white from being quantized to black due
						// to the addition of error being rounded. This can be seen if you remove
						// the +0.5 and test Floyd-Steinberg dithering on the gradient.png image.
						clamp(float32(r)+float32(er)*d.Matrix[yy][xx]+0.5),
						clamp(float32(g)+float32(eg)*d.Matrix[yy][xx]+0.5),
						clamp(float32(b)+float32(eb)*d.Matrix[yy][xx]+0.5),
					)
				}
			}

			// Reset the x value to not mess up the for loop
			// The x value is only changed when (d.Serpentine && y%2 == 0)
			// But it's reset every time to avoid another if statement
			x = oldX
		}
	}
	return ret
}

// GetColorModel returns the Ditherer's palette as a color.Model that finds the
// closest color using Euclidean distance in sRGB space.
func (d *Ditherer) GetColorModel() color.Model {
	return color.Palette(d.palette)
}

// DitherConfig is like Dither, but returns an image.Config as well.
func (d *Ditherer) DitherConfig(src draw.Image) (image.Image, image.Config) {
	return d.Dither(src), image.Config{
		ColorModel: d.GetColorModel(),
		Width:      src.Bounds().Dx(),
		Height:     src.Bounds().Dy(),
	}
}

// DitherCopy dithers a copy of the src image and returns it. The src image remains
// unchanged. If you don't need to keep the original image, use Dither.
func (d *Ditherer) DitherCopy(src image.Image) *image.RGBA {
	if d.invalid() {
		panic("dither: invalid Ditherer")
	}

	dst := copyOfImage(src)
	d.Dither(dst) // Will always return nil since dst is *image.RGBA
	return dst
}

// DitherCopyConfig is like DitherCopy, but returns an image.Config as well.
func (d *Ditherer) DitherCopyConfig(src image.Image) (*image.RGBA, image.Config) {
	return d.DitherCopy(src), image.Config{
		ColorModel: d.GetColorModel(),
		Width:      src.Bounds().Dx(),
		Height:     src.Bounds().Dy(),
	}
}

// DitherPaletted dithers a copy of the src image and returns it as an
// *image.Paletted. The src image remains unchanged. If you don't need an
// *image.Paletted, using Dither or DitherCopy should be preferred.
//
// If the Ditherer's palette has over 256 colors then the function will panic,
// because *image.Paletted does not allow for that.
func (d *Ditherer) DitherPaletted(src image.Image) *image.Paletted {
	if len(d.palette) > 256 {
		panic("dither: DitherPaletted: palette has over 256 colors which *image.Paletted doesn't support")
	}

	rgba := d.DitherCopy(src)
	p := image.NewPaletted(rgba.Bounds(), d.palette)
	copyImage(p, rgba)
	return p
}

// DitherPalettedConfig is like DitherPaletted, but returns an image.Config as well.
func (d *Ditherer) DitherPalettedConfig(src image.Image) (*image.Paletted, image.Config) {
	return d.DitherPaletted(src), image.Config{
		ColorModel: d.GetColorModel(),
		Width:      src.Bounds().Dx(),
		Height:     src.Bounds().Dy(),
	}
}

// clamp clamps i to the interval [0, 255].
func clamp(i float32) uint8 {
	if i < 0 {
		return 0
	}
	if i > 255 {
		return 255
	}
	return uint8(i)
}

// copyImage copies src's pixels into dst.
// They must be the same size.
func copyImage(dst draw.Image, src image.Image) {
	draw.Draw(dst, src.Bounds(), src, src.Bounds().Min, draw.Src)
}

func copyOfImage(img image.Image) *image.RGBA {
	dst := image.NewRGBA(img.Bounds())
	copyImage(dst, img)
	return dst
}

// samePalette returns true if both palettes contain the same colors,
// regardless of order.
func samePalette(p1 []color.Color, p2 []color.Color) bool {
	if len(p1) != len(p2) {
		return false
	}

	// Modified from: https://stackoverflow.com/a/36000696/7361270

	diff := make(map[color.Color]int, len(p1))
	for _, x := range p1 {
		// 0 value for int is 0, so just increment a counter for the string
		diff[x]++
	}
	for _, y := range p2 {
		// If _y is not in diff bail out early
		if _, ok := diff[y]; !ok {
			return false
		}
		diff[y] -= 1
		if diff[y] == 0 {
			delete(diff, y)
		}
	}
	return len(diff) == 0
}
