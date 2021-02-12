package dither

import (
	"image/color"
	"math"
)

// linearize1 linearizes an R, G, or B channel value from an sRGB color.
// Must be in the range [0, 1].
func linearize1(v float32) float32 {
	if v <= 0.04045 {
		return v / 12.92
	}
	return float32(math.Pow((float64(v)+0.055)/1.055, 2.4))
}

func linearize65535to255(i uint16) uint8 {
	v := float32(i) / 65535.0
	return uint8(linearize1(v)*255.0 + 0.5)
}

func linearize255(i uint8) uint8 {
	v := float32(i) / 255.0
	return uint8(linearize1(v)*255.0 + 0.5)
}

// toLinearRGB converts a non-linear sRGB color to a linear RGB color space.
func toLinearRGB(c color.Color) (uint8, uint8, uint8) {
	// Optimize for different color types
	switch v := c.(type) {
	case color.Gray:
		g := linearize255(v.Y)
		return g, g, g
	case color.Gray16:
		g := linearize65535to255(v.Y)
		return g, g, g
	case color.NRGBA:
		return linearize255(v.R), linearize255(v.G), linearize255(v.B)
	case color.NRGBA64:
		return linearize65535to255(v.R), linearize65535to255(v.G), linearize65535to255(v.B)
	case color.RGBA:
		return linearize255(v.R), linearize255(v.G), linearize255(v.B)
	case color.RGBA64:
		return linearize65535to255(v.R), linearize65535to255(v.G), linearize65535to255(v.B)
	}

	r, g, b, _ := c.RGBA()
	return linearize65535to255(uint16(r)), linearize65535to255(uint16(g)), linearize65535to255(uint16(b))
}
