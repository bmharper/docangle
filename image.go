package docangle

import (
	"math"

	"github.com/bmharper/cimg/v2"
)

// 8-bit grayscale image with stride = Width
type Image struct {
	Width  int
	Height int
	Pixels []byte
}

// Rotate image 90 degrees clockwise
func (s *Image) Rotate90() *Image {
	dst := make([]byte, s.Width*s.Height)
	for y := 0; y < s.Height; y++ {
		for x := 0; x < s.Width; x++ {
			dst[x*s.Height+y] = s.Pixels[y*s.Width+x]
		}
	}
	return &Image{
		Width:  s.Height,
		Height: s.Width,
		Pixels: dst,
	}
}

func (s *Image) shrinkImageIfLargerThan(maxSize int) *Image {
	scaleX := float64(maxSize) / float64(s.Width)
	scaleY := float64(maxSize) / float64(s.Height)
	if scaleX < 1 || scaleY < 1 {
		scale := min(scaleX, scaleY)
		wrapped := cimg.WrapImage(s.Width, s.Height, cimg.PixelFormatGRAY, s.Pixels)
		resized := cimg.ResizeNew(wrapped, int(math.Round(float64(s.Width)*scale)), int(math.Round(float64(s.Height)*scale)), nil)
		if resized.Stride != resized.Width {
			panic("unexpected stride")
		}
		return &Image{
			Width:  resized.Width,
			Height: resized.Height,
			Pixels: resized.Pixels,
		}
	}
	return s
}
