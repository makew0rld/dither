package dither

// This file contains code that implements some image/draw interfaces, namely
// draw.Drawer and draw.Quantizer.

import (
	"image"
	"image/color"
	"image/draw"
)

// subImager is a draw.Image that also implements SubImage. All stdlib image types
// that are already draw.Image implement this.
type subImager interface {
	draw.Image
	SubImage(r image.Rectangle) image.Image
}

func sameColor(c1 color.Color, c2 color.Color) bool {
	r1, g1, b1, a1 := c1.RGBA()
	r2, g2, b2, a2 := c2.RGBA()
	return r1 == r2 && g1 == g2 && b1 == b2 && a1 == a2
}

// subset returns true if p1 is a subset of p2, regardless of the order
// of elements.
func subset(p1 []color.Color, p2 []color.Color) bool {
	for i := range p1 {
		for j := range p2 {
			if sameColor(p1[i], p2[j]) {
				// Move on to the next color of the provided palette
				break
			}
			if j == len(p2)-1 && !sameColor(p1[i], p2[j]) {
				// The last color of the second palette, and no match with
				// the first palette's color has been found.
				// This means that p1 is not a subset of p2
				return false
			}
		}
	}
	return true
}

// Draw implements draw.Drawer. This means you can use a Ditherer
// in many places, such as for encoding GIFs.
//
// Draw ignores whether dst has a palette or not, and just uses the internal Ditherer
// palette. If the dst image passed has a palette (i.e. is of the type *image.Paletted),
// and the palette is the not the same as the Ditherer's palette, it will panic.
func (d *Ditherer) Draw(dst draw.Image, r image.Rectangle, src image.Image, sp image.Point) {
	if d.invalid() {
		panic("dither: invalid Ditherer")
	}

	dst2 := dst
	paletted := false
	if p, ok := dst.(*image.Paletted); ok {
		if !samePalette(d.palette, p.Palette) {
			panic("dither: Draw: dst was an *image.Paletted that doesn't have the same palette")
		}
		// src needs to copied onto dst, and then dst is dithered
		// But dst is paletted and so the copy will change colors
		// So instead an RGBA copy of dst is made, and then values are copied back
		// into the paletted image after dithering, at the bottom of the function.
		dst2 = copyOfImage(dst)
		paletted = true
	}
	// No longer use dst, only dst2

	dst3, ok := dst2.(subImager)
	if !ok {
		panic("dither: Draw: dst Image passed does not have SubImage method")
	}
	// No longer use dst2, only dst3 - they are the same object but it's easier
	// to stick to one

	// Like Go stdlib does with their Drawer:
	// https://github.com/golang/go/blob/go1.15.7/src/image/draw/draw.go#L62
	//
	// This is done here, even though draw.Draw will take care of it. That's
	// because the rectangle I have needs to be clipped because it's used later
	// to only dither the correct area.
	clip(dst3, &r, src, &sp, nil, nil)
	if r.Empty() {
		return
	}

	// Copy src onto dst, using the provided boundaries (see draw.Drawer for more)
	draw.Draw(dst3, r, src, sp, draw.Src)

	// Then dither only the newly-copied area
	d.Dither(dst3.SubImage(r).(draw.Image))

	if paletted {
		// The dithered values in the RGBA image need to copied back into the
		// original paletted image. See above.
		copyImage(dst, dst2)
	}
}

// clip clips r against each image's bounds (after translating into the
// destination image's coordinate space) and shifts the points sp and mp by
// the same amount as the change in r.Min.
//
// Copied from Go stdlib, see
//     https://github.com/golang/go/blob/go1.15.7/src/image/draw/draw.go#L73
func clip(dst draw.Image, r *image.Rectangle, src image.Image, sp *image.Point, mask image.Image, mp *image.Point) {
	orig := r.Min
	*r = r.Intersect(dst.Bounds())
	*r = r.Intersect(src.Bounds().Add(orig.Sub(*sp)))
	if mask != nil {
		*r = r.Intersect(mask.Bounds().Add(orig.Sub(*mp)))
	}
	dx := r.Min.X - orig.X
	dy := r.Min.Y - orig.Y
	if dx == 0 && dy == 0 {
		return
	}
	sp.X += dx
	sp.Y += dy
	if mp != nil {
		mp.X += dx
		mp.Y += dy
	}
}

// Quantize implements draw.Quantizer. It ignores the provided image
// and just returns the Ditherer's palette each time. This is useful for places that
// only allow you to set the palette through a draw.Quantizer, like the image/gif
// package.
//
// This function will panic if the Ditherer's palette has more colors than the
// caller wants, which the caller indicates by cap(p).
//
// It will also panic if there's already colors in the color.Palette provided
// to the func and not all of those colors are included in the Ditherer's palette.
// This is because the caller is indicating that certain colors must be in the
// palette, but the user who created the Ditherer does not want those colors.
func (d *Ditherer) Quantize(p color.Palette, m image.Image) color.Palette {
	if cap(p) < len(d.palette) {
		// The Ditherer palette has more colors than allowed
		panic("dither: Quantize: Ditherer palette has too many colors for this Quantize call")
	}
	if len(p) > len(d.palette) {
		// There's already colors in the palette, more than the Ditherer's
		// Note this assumes there aren't duplicate colors in the palette
		panic("dither: Quantize: provided palette has colors the Ditherer palette doesn't")
	}
	if len(p) > 0 && !subset(p, d.palette) {
		// There's already colors in the palette, but they aren't all included
		// in the Ditherer's palette
		panic("dither: Quantize: provided palette has colors the Ditherer palette doesn't")
	}
	return d.palette
}
