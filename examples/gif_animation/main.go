package main

// This more complicated example shows how to dither a GIF animation.

import (
	"fmt"
	"image"
	"image/color"
	"image/gif"
	_ "image/png" // For frame decoding
	"os"

	"github.com/makeworld-the-better-one/dither/v2"
)

const numFrames = 20

func main() {
	// Create the kind of Dither we want

	palette := []color.Color{
		color.Black,
		color.White,
		color.RGBA{255, 0, 0, 255}, // Red
		color.RGBA{0, 255, 0, 255}, // Green
		color.RGBA{0, 0, 255, 255}, // Blue
	}
	d := dither.NewDitherer(palette)
	d.Matrix = dither.FloydSteinberg // Why not?

	// Decode first frame and get image.Config for use in gif.GIF.
	// gif.GIF requires *image.Paletted is used, so DitherPaletted
	// is called instead of Dither.

	f, err := os.Open("../input/ball_001.png")
	if err != nil {
		panic(err)
	}
	img, _, err := image.Decode(f)
	if err != nil {
		panic(err)
	}
	f.Close()
	firstFrame, config := d.DitherPalettedConfig(img)

	frames := make([]*image.Paletted, numFrames)
	frames[0] = firstFrame

	// Decode other frames
	for i := 1; i < numFrames; i++ {
		f, err := os.Open(fmt.Sprintf("../input/ball_0%02d.png", i))
		if err != nil {
			panic(err)
		}
		img, _, err := image.Decode(f)
		if err != nil {
			panic(err)
		}
		f.Close()

		frames[i] = d.DitherPaletted(img)
	}

	// Frame delay - same for each frame
	delays := make([]int, numFrames)
	for i := range delays {
		delays[i] = 7
	}

	// Setup GIF and encode
	g := gif.GIF{
		Image: frames,
		Delay: delays,

		// By specifying a Config, we can set a global color table for the GIF.
		// This is more efficient then each frame having its own color table, which
		// is the default when there's no config.
		Config: config,
	}

	f2, err := os.Create("../output/gif_animation.gif")
	if err != nil {
		panic(err)
	}

	err = gif.EncodeAll(f2, &g)
	if err != nil {
		panic(err)
	}
}
