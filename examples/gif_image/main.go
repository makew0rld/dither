package main

// This example showcases how the interfaces Ditherer implements can be easily
// used to encode a full color image as a GIF, which can only use 256 colors.

import (
	"image"
	"image/color"
	"image/gif"
	_ "image/png" // Imported for decoding of the input image
	"os"

	"github.com/makeworld-the-better-one/dither/v2"
)

func main() {
	// Create the kind of Dither we want

	palette := []color.Color{
		color.Black,
		color.White,
		color.RGBA{255, 255, 0, 255}, // Yellow
	}
	d := dither.NewDitherer(palette)
	d.Mapper = dither.Bayer(8, 8, 1.0) // Why not?

	// GIF settings - all of these are required!
	opts := gif.Options{
		NumColors: len(palette),
		Quantizer: d, // dither.Ditherer fulfills both these interfaces!
		Drawer:    d, // How useful!
	}

	// Open an image and save it as a dithered GIF

	f, err := os.Open("../../images/input/peppers.png")
	if err != nil {
		panic(err)
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}

	f2, err := os.Create("../output/gif_image.gif")
	if err != nil {
		panic(err)
	}

	// The GIF encoder calls on the Ditherer itself, because it's the Drawer in
	// the GIF options.
	err = gif.Encode(f2, img, &opts)
	if err != nil {
		panic(err)
	}
}
