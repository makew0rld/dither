package dither

import (
	"image"
	"image/color"
	"image/draw"
	"runtime"
	"sync"
)

// parallel parallelizes per-pixel image modifications, dividing the
// image up horizontally depending on the number of workers.
//
// Setting numWorkers to 0 or below will result in runtime.GOMAXPROCS(0) workers being used.
func parallel(workers int, dst draw.Image, src image.Image, f func(x, y int, c color.Color) color.Color) {
	if workers <= 0 {
		workers = runtime.GOMAXPROCS(0)
	}

	b := src.Bounds()
	height := b.Dy()

	worker := func(minY, maxY int) {
		for y := minY; y < maxY; y++ {
			for x := b.Min.X; x < b.Max.X; x++ {
				dst.Set(x, y, f(x, y, src.At(x, y)))
			}
		}
	}

	if workers == 1 || height == 1 {
		// Fast path for just using one worker
		worker(b.Min.Y, b.Max.Y)
		return
	}

	partSize := height / workers

	if partSize == 0 {
		// workers > height
		workers = height
		partSize = 1
	}

	var wg sync.WaitGroup

	// Launch workers
	for i := 0; i < workers; i++ {
		var min, max int // Beginning and end of this part (Y axis values)

		if i+1 == workers {
			// Last part
			// Fix off-by-one error, catch last line
			min = partSize*i + b.Min.Y
			max = b.Max.Y
		} else {
			min = partSize*i + b.Min.Y
			max = partSize*(i+1) + b.Min.Y
		}

		wg.Add(1)
		go func() {
			worker(min, max)
			wg.Done()
		}()
	}

	wg.Wait()
}
