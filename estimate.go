package docangle

import (
	"math"
)

// #include "line.h"
import "C"

const deg2Rad = math.Pi / 180
const rad2Deg = 180 / math.Pi

type lineSetup struct {
	dx    int
	dy    int
	stepY int
	// stepX is always 1
	gradient int32 // fixed-point 16.16
}

// angle must be between -45 and +45 degrees
func setupLine(angle float64) lineSetup {
	angle *= deg2Rad
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

	ls := lineSetup{
		dx:       int(dx),
		dy:       int(dy),
		stepY:    stepY,
		gradient: gradient,
	}

	return ls
}

func lineScore(img Image, x, y, width int, angle float64, whiteThreshold int) float64 {
	ls := setupLine(angle)
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

	// The maximum number of white lines we expect to see in the image.
	totalLineCount := img.Height - padY*2

	// Score for each angle and y position
	scoreAtAngleAndY := make([][]float64, len(angles))
	for i := range angles {
		scoreAtAngleAndY[i] = make([]float64, totalLineCount)
	}

	for y := padY; y < img.Height-padY; y++ {
		for i := range angles {
			score := lineScore(*img, x, y, img.Width-padX*2, angles[i], pixelIsWhiteThreshold)
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

// Parameters to the GetAngleWhiteLines function
type WhiteLinesParams struct {
	MinDeltaDegrees  float64 // Minimum angle in degrees to test for
	MaxDeltaDegrees  float64 // Maximum angle in degrees to test for
	StepDegrees      float64 // Increment of each sample in degrees
	Include90Degrees bool    // Try rotating the document by 90 degrees and then sampling through MinDeltaDegrees..MaxDeltaDegrees. This doubles the computation time.
	MaxResolution    int     // If image width or height exceeds this, then shrink image to this size before processing. Zero to disable.
}

// Create a new WhiteLinesParams with defaults
func NewWhiteLinesParams() *WhiteLinesParams {
	return &WhiteLinesParams{
		MinDeltaDegrees:  -2.5,
		MaxDeltaDegrees:  2.5,
		StepDegrees:      0.1,
		Include90Degrees: true,
		MaxResolution:    1000,
	}
}

// Given an 8-bit grayscale document image, return the estimated degrees of rotation.
// A positive value means the document has been rotated clockwise.
// This is a brute force algorithm which tries every degree increment one by one,
// and picks the angle that yields the best score.
func GetAngleWhiteLines(img *Image, params *WhiteLinesParams) (score, degrees float64) {
	if params == nil {
		params = NewWhiteLinesParams()
	}
	angles := []float64{}
	for angle := params.MinDeltaDegrees; angle <= params.MaxDeltaDegrees; angle += params.StepDegrees {
		angles = append(angles, angle)
	}

	if params.MaxResolution != 0 {
		img = img.shrinkImageIfLargerThan(params.MaxResolution)
	}

	// Test image as-is
	hScore, hAngle := getAngleWhiteLinesInner(img, angles)

	if params.Include90Degrees {
		// First rotate by 90 degrees, and then test
		rotated := img.Rotate90()
		vScore, vAngle := getAngleWhiteLinesInner(rotated, angles)
		if vScore > hScore {
			return vScore, -90 - vAngle
		}
	}
	return hScore, hAngle
}
