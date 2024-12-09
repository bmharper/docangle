package docangle

// 8-bit grayscale image
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
