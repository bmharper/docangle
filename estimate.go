package docangle

import (
	"math"
)

// #include "line.h"
import "C"

const Deg2Rad = math.Pi / 180
const Rad2Deg = 180 / math.Pi

type LineSetup struct {
	dx    int
	dy    int
	stepY int
	// stepX is always 1
	gradient int32 // fixed-point 16.16
}

// angle must be between -45 and +45 degrees
func SetupLine(width int, angle float64) LineSetup {
	angle *= Deg2Rad
	// Calculate deltas, at an arbitrary precision of a few thousand pixels
	dx := 10000 * math.Cos(angle)
	dy := 10000 * math.Sin(angle)

	gradient := int32(0)
	if dx != 0 {
		gradient = int32(dy * (1 << 16) / dx)
	}

	// Determine the direction of the iteration
	stepY := 1
	if dy < 0 {
		stepY = -1
	}
	dx = math.Abs(dx)
	dy = math.Abs(dy)

	ls := LineSetup{
		dx:       int(dx),
		dy:       int(dy),
		stepY:    stepY,
		gradient: gradient,
	}

	return ls
}

// Bresenham iteration with subpixel data for adjacent vertical (above or below) pixel
func IterateLineSubpixel(startX, startY, width int, ls LineSetup, callback func(x, y int, blend int32)) {
	dx, dy, stepY := ls.dx, ls.dy, ls.stepY

	err := dx - dy

	x, y := startX, startY
	endX := startX + width

	blend := int32(0)

	for x < endX {
		callback(x, y, blend)

		// Update error and coordinates
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x += 1
		}
		if e2 < dx {
			err += dx
			y += stepY
			blend -= int32(stepY << 16)
		}
		blend += ls.gradient
	}
}

func abs[T int | int32](x T) T {
	if x < 0 {
		return -x
	}
	return x
}

/*
func IsWhite(img Image, x, y, width int, angle float64, whiteThreshold int, proportionWhitePixels float64) bool {
	ls, _ := SetupLine(width, angle)
	nWhite := 0
	nBlack := 0
	// Subpixel
	IterateLineSubpixel(x, y, width, ls, func(px, py int, blend int32) {
		var va, vb int32
		line := py * img.Width
		va = int32(img.Pixels[line+px])
		if blend < 0 {
			vb = int32(img.Pixels[line-img.Width+px])
		} else {
			vb = int32(img.Pixels[line+img.Width+px])
		}
		wb := abs(blend)
		wa := 65536 - wb
		blended := int((va*wa + vb*wb) >> 16)
		if blended > whiteThreshold {
			nWhite++
		} else {
			nBlack++
		}
	})
	return float64(nWhite)/float64(nWhite+nBlack) > proportionWhitePixels
}
*/

// Inline iteration without callback
// Returns the count of white and black pixels along the sampled line.
func IterateLineSubpixelBaked(img Image, startX, startY, width int, ls LineSetup, whiteThreshold int) (int, int) {
	pixels := img.Pixels
	imgWidth := img.Width

	dx, dy, stepY := ls.dx, ls.dy, ls.stepY
	gradient := ls.gradient

	err := dx - dy
	x, y := startX, startY
	endX := startX + width

	line := y * imgWidth
	nWhite := 0
	nBlack := 0
	wt := int32(whiteThreshold)

	blend := int32(0)

	for x < endX {
		va := int32(pixels[line+x])

		// Compute absolute value of blend and select the vertical pixel:
		wb := blend
		var vb int32
		if wb < 0 {
			wb = -wb
			vb = int32(pixels[line-imgWidth+x])
		} else {
			vb = int32(pixels[line+imgWidth+x])
		}

		diff := vb - va
		blended := va + int32((diff*wb)>>16)
		if blended > wt {
			nWhite++
		} else {
			nBlack++
		}

		// Update error and coordinates
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x += 1
		}
		if e2 < dx {
			err += dx
			y += stepY
			line += stepY * imgWidth
			blend -= int32(stepY << 16)
		}
		blend += gradient
	}

	return nWhite, nBlack
}

func IsWhite(img Image, x, y, width int, angle float64, whiteThreshold int, proportionWhitePixels float64) bool {
	ls := SetupLine(width, angle)
	nWhite, nBlack := IterateLineSubpixelBaked(img, x, y, width, ls, whiteThreshold)
	return float64(nWhite)/float64(nWhite+nBlack) > proportionWhitePixels
}

func LineScore(img Image, x, y, width int, angle float64, whiteThreshold int) float64 {
	ls := SetupLine(width, angle)
	var nWhite, nBlack, nTransitions C.int
	args := C.IterateLineSubpixelBakedC_Args{
		imgWidth:       C.int(img.Width),
		startX:         C.int(x),
		startY:         C.int(y),
		width:          C.int(width),
		dx:             C.int(ls.dx),
		dy:             C.int(ls.dy),
		stepY:          C.int(ls.stepY),
		gradient:       C.int32_t(ls.gradient),
		whiteThreshold: C.int(whiteThreshold),
	}
	C.IterateLineSubpixelBakedC(&args, (*C.uint8_t)(&img.Pixels[0]), &nWhite, &nBlack, &nTransitions)
	// The less transitions the better.
	// Note that the threshold in countTransitions() is dependent on the scoring function.
	// 1:    No transitions
	// 0.5:  1 transition
	// 0.33: 2 transitions
	return 1 / (1 + float64(nTransitions))
	//if nTransitions < 5 {
	//	return 1
	//} else {
	//	return 0
	//}

	//if float64(nWhite)/float64(nWhite+nBlack) > proportionWhitePixels {
	//	return 1
	//} else {
	//	return 0
	//}
}

func GetAngleWhiteLines(img *Image) (score, degrees float64) {
	delta := 2.0 // degrees
	angles := []float64{}
	for angle := -delta; angle <= delta; angle += 0.1 {
		angles = append(angles, angle)
	}
	hScore, hAngle := getAngleWhiteLinesInner(img, angles)

	// try 90 degrees rotated
	rotated := img.Rotate90()
	vScore, vAngle := getAngleWhiteLinesInner(rotated, angles)

	if hScore > vScore {
		return hScore, hAngle
	}
	return vScore, -90 - vAngle
}

// Count the number of times that score[i] flips between 0 and 1
func countTransitions(score []float64) int {
	on := false
	transitions := 0
	for _, s := range score {
		// this threshold is dependent on our scoring function
		// If you plot a histogram of a positive case, you'll see extremely bimodal distribution, with
		// a ton of values at 1.0, and tons of values below 0.2.
		//fmt.Printf("%.3f,", s)
		high := s > 0.3
		if high != on {
			transitions++
			on = high
		}
	}
	return transitions
}

// run horizontal lines across the image, at all test angles, and pick the angle
// where we get the most uninterrupted lines (pure white)
func getAngleWhiteLinesInner(img *Image, angles []float64) (score, angle float64) {

	padX := img.Width / 10  // arbitrary padding, but we stay away from the edges to avoid binder holes from affecting us
	padY := img.Height / 10 // Padding to prevent us overflowing the image, which we will do, because our lines are rotated

	scoreAtAngle := make([]float64, len(angles))
	x := padX

	// 0..255 threshold where we consider a pixel white
	// Note that this algorithm tends to work even if this threshold is very low.
	pixelIsWhiteThreshold := 200

	// Proportion of pixels that must be white for us to consider the line white
	// (used in prior algo)
	//lineIsWhiteThreshold := 0.98

	// The maximum number of white lines we expect to see in the image.
	totalLineCount := img.Height - padY*2

	// Score for each angle and y position
	scoreAtAngleAndY := make([][]float64, len(angles))
	for i := range angles {
		scoreAtAngleAndY[i] = make([]float64, totalLineCount)
	}

	for y := padY; y < img.Height-padY; y++ {
		for i := range angles {
			score := LineScore(*img, x, y, img.Width-padX*2, angles[i], pixelIsWhiteThreshold)
			scoreAtAngle[i] += score
			scoreAtAngleAndY[i][y-padY] = score
		}
	}
	maxScore := 0.0
	for _, score := range scoreAtAngle {
		maxScore = max(maxScore, score)
	}
	bestScore := 0.0
	bestAngle := 0.0 // degrees
	//bestTransitions := 0
	for i, score := range scoreAtAngle {
		transitions := countTransitions(scoreAtAngleAndY[i])
		degrees := angles[i]
		//fmt.Printf("angle %.1f degrees: %v\n", degrees, lines)
		// Ensure there are at least X transitions, so that we don't pick a page with a narrow column of text down the middle,
		// and two wide margins. We need transitions between white line and content.
		if score > bestScore && transitions >= 10 {
			bestAngle = degrees
			bestScore = score
			//bestTransitions = transitions
		}
	}
	//fmt.Printf("Transitions: %v\n", bestTransitions)
	return float64(bestScore) / float64(totalLineCount), bestAngle
}
