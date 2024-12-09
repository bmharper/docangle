package docangle

import (
	"path/filepath"
	"testing"
	"time"

	"github.com/bmharper/cimg/v2"
	"github.com/stretchr/testify/require"
)

var NumTested int
var TotalTime time.Duration

func makeGray(src *cimg.Image) *Image {
	dst := &Image{
		Width:  src.Width,
		Height: src.Height,
		Pixels: make([]byte, src.Width*src.Height),
	}
	rgb := src.ToRGB()
	for y := 0; y < rgb.Height; y++ {
		srcLine := rgb.Pixels[rgb.Stride*y : rgb.Stride*(y+1)]
		dstLine := dst.Pixels[y*rgb.Width : (y+1)*rgb.Width]
		i := 0
		for x := 0; x < rgb.Width; x++ {
			r := uint32(srcLine[i])
			g := uint32(srcLine[i+1])
			b := uint32(srcLine[i+2])
			gray := byte((306*r + 601*g + 117*b) >> 10)
			dstLine[x] = gray
			i += 3
		}
	}
	return dst
}

func saveImage(img *Image) {
	rgb := cimg.NewImage(img.Width, img.Height, cimg.PixelFormatRGB)
	for y := 0; y < img.Height; y++ {
		srcLine := img.Pixels[img.Width*y : img.Width*(y+1)]
		dstLine := rgb.Pixels[rgb.Stride*y : rgb.Stride*(y+1)]
		i := 0
		for x := 0; x < img.Width; x++ {
			dstLine[i] = srcLine[x]
			dstLine[i+1] = srcLine[x]
			dstLine[i+2] = srcLine[x]
			i += 3
		}
	}
	rgb.WriteJPEG("gray.jpg", cimg.MakeCompressParams(cimg.Sampling420, 95, 0), 0644)
}

func unrotateAndSave(img *cimg.Image, degrees float64, filename string) {
	if img.NChan() != 3 {
		img = img.ToRGB()
	}
	rotated := RotateImage(img, -degrees*Deg2Rad)
	rotated.WriteJPEG(filename, cimg.MakeCompressParams(cimg.Sampling444, 95, 0), 0644)
}

func testImage(t *testing.T, filename string, expectedAngle float64) {
	img, err := cimg.ReadFile(filename)
	require.NoError(t, err)
	gray := makeGray(img)
	//saveImage(gray)
	start := time.Now()
	score, angle := GetAngleWhiteLines(gray)
	NumTested++
	duration := time.Since(start)
	TotalTime += duration
	t.Logf("%30v angle: %5.1f, %5.1f expected (time %4dms). score %.2f", filepath.Base(filename), angle, expectedAngle, duration.Milliseconds(), score)
	if expectedAngle != 999 {
		require.InDelta(t, expectedAngle, angle, 0.2)
	}
	unrotatedFilename := filepath.Join("unrotated", filepath.Base(filename))
	unrotateAndSave(img, angle, unrotatedFilename)
}

func TestImages(t *testing.T) {
	testImage(t, "testimages/red1.jpg", 0.5)
	testImage(t, "testimages/red2.jpg", 1.5)
	testImage(t, "testimages/red3.jpg", -89.4)
	testImage(t, "testimages/diamond_1_Image1.jpg", 0.3)
	testImage(t, "testimages/diamond_2_Image1.png", 1.3)
	testImage(t, "testimages/bpm_1_X1.jpg", 999)
	testImage(t, "testimages/bpm_2_X1.jpg", 999)
	testImage(t, "testimages/bpm_3_X1.jpg", 999)
	testImage(t, "testimages/bpm_4_X1.jpg", 999)
	testImage(t, "testimages/bpm_5_X1.jpg", 999)
	testImage(t, "testimages/bpm_6_X1.jpg", 999)
	testImage(t, "testimages/cadgrafics_1_x2.jpg", 999)
	testImage(t, "testimages/cadgrafics_2_x5.jpg", 999)
	testImage(t, "testimages/caper_1_Im0.jpg", 999)
	testImage(t, "testimages/caper_2_Im1.jpg", 999)
	testImage(t, "testimages/buscon_1_Im1.jpg", 0.3)
	testImage(t, "testimages/buscon_2_Im2.jpg", 0.5)
	t.Logf("Average time per document: %v", TotalTime/time.Duration(NumTested))
}
