package docangle

// #include "rotate.h"
import "C"
import "github.com/bmharper/cimg/v2"

func RotateImage(img *cimg.Image, radians float64) *cimg.Image {
	outWidth, outHeight := img.Width, img.Height
	degrees := radians * Rad2Deg
	if (-135 < degrees && degrees < -45) || (45 < degrees && degrees < 135) {
		// swap landscape/portrait
		outWidth, outHeight = img.Height, img.Width
	}

	dst := cimg.NewImage(outWidth, outHeight, img.Format)
	C.rotate_image_bilinear((*C.uint8_t)(&img.Pixels[0]), (*C.uint8_t)(&dst.Pixels[0]),
		C.int(img.Width), C.int(img.Height),
		C.int(dst.Width), C.int(dst.Height),
		C.double(radians))
	return dst
}
