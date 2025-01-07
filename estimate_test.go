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

func rotateImage(img *cimg.Image, radians float64) *cimg.Image {
	outWidth, outHeight := img.Width, img.Height
	degrees := radians * rad2Deg
	if (-135 < degrees && degrees < -45) || (45 < degrees && degrees < 135) {
		// swap landscape/portrait
		outWidth, outHeight = img.Height, img.Width
	}

	dst := cimg.NewImage(outWidth, outHeight, img.Format)
	cimg.Rotate(img, dst, radians, nil)
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
	rotated := rotateImage(img, -degrees*deg2Rad)
	rotated.WriteJPEG(filename, cimg.MakeCompressParams(cimg.Sampling444, 95, 0), 0644)
}

func testImage(t *testing.T, filename string, expectedAngle float64) {
	img, err := cimg.ReadFile(filename)
	require.NoError(t, err)
	gray := makeGray(img)
	//saveImage(gray)
	start := time.Now()
	params := NewWhiteLinesParams()
	params.Include90Degrees = true
	score, angle := GetAngleWhiteLines(gray, params)
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
	// Unfortunately these can't be included in the public repo,
	// because they're financial statements of private companies.
	//testImage(t, "testimages/private/red1.jpg", 0.5)
	//testImage(t, "testimages/private/red2.jpg", 1.5)
	//testImage(t, "testimages/private/red3.jpg", 90.5)
	//testImage(t, "testimages/private/diamond_1_Image1.jpg", 0.3)
	//testImage(t, "testimages/private/diamond_2_Image1.png", 1.3)
	//testImage(t, "testimages/private/bpm_1_X1.jpg", 0.5)
	//testImage(t, "testimages/private/bpm_2_X1.jpg", -0.2)
	//testImage(t, "testimages/private/bpm_3_X1.jpg", -0.7)
	//testImage(t, "testimages/private/bpm_4_X1.jpg", -0.6)
	//testImage(t, "testimages/private/bpm_5_X1.jpg", -0.7)
	//testImage(t, "testimages/private/bpm_6_X1.jpg", -0.4)
	//testImage(t, "testimages/private/cadgrafics_1_x2.jpg", 89.8)
	//testImage(t, "testimages/private/cadgrafics_2_x5.jpg", -1.0)
	//testImage(t, "testimages/private/caper_1_Im0.jpg", 0.1)
	//testImage(t, "testimages/private/caper_2_Im1.jpg", 0)
	//testImage(t, "testimages/private/buscon_1_Im1.jpg", 0.3)
	//testImage(t, "testimages/private/buscon_2_Im2.jpg", 0.5)

	testImage(t, "testimages/nvidia-1.jpg", 1.7)
	testImage(t, "testimages/nvidia-2.jpg", 88.8)
	testImage(t, "testimages/xerox-1.jpg", -0.8)

	t.Logf("Average time per page: %v", TotalTime/time.Duration(NumTested))
}
