# Doc Angle

This is a tiny bit of code to compute the angle of a scanned document. It's a
simple brute force approach that tries many different candidate angles by
drawing lines through the image and finding which angle produces the cleanest
lines. A 'clean' line is one which does not oscillate between dark and light
pixels often. The more oscillations between dark and light, the worse the line.
We make no attempt to determine an automatic threshold for dark and light, but
just hardcode the value 200 (out of 255) as the threshold.

Because of our brute force approach, the function is slow. So it's only really
usable if you know a-priori that the document is close to straight. The major
cloud OCR vendors all seem to deal quite well with documents that are
signifiantly skew (eg 3 degrees), but AWS Textract in particular struggles with
documents that have only 1.5 degrees or less, of rotation. In low rotation
scenarios, the OCR system will sometimes fail to treat the rotation, and this
completely breaks the comprehension of tables in the documents, because cells
get assigned to the wrong row.

This code is intended to detect those low rotation scenarios so that one can
correct for them. By default, we only scan from -2.5 to 2.5 degrees, in 0.1
degree increments. This takes about 50ms per image on a 2024 era x86 CPU.

The function has a MaxResolution parameter, which is invoked if either the width
or the height of the image is greater than this value. In that case, we first
downsample the image to MaxResolution before processing.

## Usage

> go get github.com/bmharper/docangle

```go
// Input is a grayscale 8-bit image

// Simple case using default
_, degrees := docangle.GetAngleWhiteLines(img, nil)

// Override defaults
params := docangle.NewWhiteLinesParams()
params.MaxResolution = 800
_, degrees = docangle.GetAngleWhiteLines(img, &params)
```
