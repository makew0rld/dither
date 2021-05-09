package dither

import (
	"image/color"
	"math"
)

// linearize1 linearizes an R, G, or B channel value from an sRGB color.
// Must be in the range [0, 1].
func linearize1(v float64) float64 {
	if v <= 0.04045 {
		return v / 12.92
	}
	return math.Pow((v+0.055)/1.055, 2.4)
}

func linearize65535(i uint16) uint16 {
	v := float64(i) / 65535.0
	return uint16(math.RoundToEven(linearize1(v) * 65535.0))
}

func linearize255to65535(i uint8) uint16 {
	v := float64(i) / 255.0
	return uint16(math.RoundToEven(linearize1(v) * 65535.0))
}

// toLinearRGB converts a non-linear sRGB color to a linear RGB color space.
// RGB values are taken directly and alpha value is ignored, so this will not
// handle non-opaque colors properly.
func toLinearRGB(c color.Color) (uint16, uint16, uint16) {
	// Optimize for different color types
	switch v := c.(type) {
	case color.Gray:
		g := linearize255to65535(v.Y)
		return g, g, g
	case color.Gray16:
		g := linearize65535(v.Y)
		return g, g, g
	case color.NRGBA:
		return linearize255to65535(v.R), linearize255to65535(v.G), linearize255to65535(v.B)
	case color.NRGBA64:
		return linearize65535(v.R), linearize65535(v.G), linearize65535(v.B)
	case color.RGBA:
		return linearize255to65535(v.R), linearize255to65535(v.G), linearize255to65535(v.B)
	case color.RGBA64:
		return linearize65535(v.R), linearize65535(v.G), linearize65535(v.B)
	}

	r, g, b, _ := c.RGBA()
	return linearize65535(uint16(r)), linearize65535(uint16(g)), linearize65535(uint16(b))
}
