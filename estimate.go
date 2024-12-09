package docangle

import (
	"fmt"
	"math"
	"slices"
)

const Deg2Rad = math.Pi / 180

//const LineErrShift = 16
//const LineErrMul = (1 << LineErrShift)

type LineSetup struct {
	dx    int
	dy    int
	stepY int
	// stepX is always 1
	gradient int32 // fixed-point 16.16
}

// angle must be between -45 and +45 degrees
func SetupLine(width int, angle float64) (LineSetup, int) {
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

	// Determine the max distance we'll walk in the Y direction
	length := float64(width) / math.Cos(angle)
	yDelta := int(math.Ceil(length * math.Sin(angle)))
	return ls, yDelta
}

// Bresenham iteration
func IterateLine(startX, startY, width int, ls LineSetup, callback func(x, y int)) {
	dx, dy, stepY := ls.dx, ls.dy, ls.stepY

	err := dx - dy

	x, y := startX, startY
	endX := startX + width

	for x < endX {
		callback(x, y)

		// Update error and coordinates
		e2 := 2 * err
		if e2 > -dy {
			err -= dy
			x += 1
		}
		if e2 < dx {
			err += dx
			y += stepY
		}
	}
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

// Returns the max delta between two rotated scanlines
func AnalyzeBlock(img Image, x, y, width, height int, angle float64) int {
	ls, _ := SetupLine(width, angle)
	prev1 := 0
	prev2 := 0
	prev3 := 0
	maxDelta := 0
	for dy := 0; dy < height; dy++ {
		total := 0
		// Whole pixel
		//IterateLine(x, y+dy, width, ls, func(px, py int) {
		//	total += int(img.Pixels[py*img.Width+px])
		//})
		// Subpixel
		IterateLineSubpixel(x, y+dy, width, ls, func(px, py int, blend int32) {
			var va, vb int32
			va = int32(img.Pixels[py*img.Width+px])
			if blend < 0 {
				vb = int32(img.Pixels[(py-1)*img.Width+px])
			} else {
				vb = int32(img.Pixels[(py+1)*img.Width+px])
			}
			wb := abs(blend)
			wa := 65536 - wb
			total += int((va*wa + vb*wb) >> 16)
		})
		if dy >= 3 {
			delta := abs(total - prev3)
			//delta := abs(total - prev1)
			if delta > maxDelta {
				maxDelta = delta
			}
		}
		prev3 = prev2
		prev2 = prev1
		prev1 = total
	}
	return maxDelta
}

func GetAngle(img *Image) float64 {
	//a, _ := SetupLine(64, -3)
	//b, _ := SetupLine(64, 3)
	//fmt.Printf("%v, %v\n", a, b)

	padX := 70 // arbitrary padding, but we stay away from the edges to avoid binder holes from affecting us
	padY := 30 // Padding to prevent us overflowing the image, which we will do, because our lines are rotated

	maxIteration := 2
	angleEstimate := 0.0 // degrees

	const angleUnit = 10.0

	for iteration := 0; iteration < maxIteration; iteration++ {
		blockWidth := 64
		blockHeight := 20
		minAngle := -50 // units are base 0.1 degrees. So 30 = 3 degrees.
		maxAngle := 50
		stepAngle := 10
		if iteration == 1 {
			// search from -1.5 to +1.5 degrees of our closest estimate
			minAngle = int(angleEstimate*10) - 15
			maxAngle = int(angleEstimate*10) + 15
			stepAngle = 1
		}
		// brute force
		//minAngle = int(-5 * 10)
		//maxAngle = int(5 * 10)
		//stepAngle = 1

		numAngleSteps := (maxAngle-minAngle)/stepAngle + 1
		blocksAtAngle := make([][]int, numAngleSteps)
		scoreAtAngle := make([]float64, numAngleSteps)
		x2 := img.Width - blockWidth - padX
		y2 := img.Height - blockHeight - padY
		//fmt.Printf("Num blocks: %v\n", (x2-padX)*(y2-padY)/(blockWidth*blockHeight))
		for x := padX; x < x2; x += blockWidth {
			for y := padY; y < y2; y += blockHeight {
				for angle := minAngle; angle <= maxAngle; angle += stepAngle {
					angleIdx := (angle - minAngle) / stepAngle
					maxDelta := AnalyzeBlock(*img, x, y, blockWidth, blockHeight, float64(angle)/angleUnit)
					blocksAtAngle[angleIdx] = append(blocksAtAngle[angleIdx], maxDelta)
				}
			}
		}
		for i, deltas := range blocksAtAngle {
			slices.Sort(deltas)
			score := 0
			for i := len(deltas) - 50; i < len(deltas); i++ {
				score += deltas[i]
			}
			scoreAtAngle[i] = float64(score)
		}
		bestScore := 0.0
		for i, score := range scoreAtAngle {
			degrees := float64(minAngle+i*stepAngle) / angleUnit
			fmt.Printf("angle %.1f degrees: %v\n", degrees, score)
			if score > bestScore {
				angleEstimate = degrees
				bestScore = score
			}
		}
	}
	return angleEstimate
}

func IsWhite(img Image, x, y, width int, angle float64, whiteThreshold int, proportionWhitePixels float64) bool {
	ls, _ := SetupLine(width, angle)
	nWhite := 0
	nBlack := 0
	// Subpixel
	IterateLineSubpixel(x, y, width, ls, func(px, py int, blend int32) {
		var va, vb int32
		va = int32(img.Pixels[py*img.Width+px])
		if blend < 0 {
			vb = int32(img.Pixels[(py-1)*img.Width+px])
		} else {
			vb = int32(img.Pixels[(py+1)*img.Width+px])
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

func GetAngle2(img *Image) float64 {
	delta := 2.5
	angles := []float64{}
	for angle := -delta; angle <= delta; angle += 0.1 {
		angles = append(angles, angle)
	}
	hScore, hAngle := GetAngleInner2(img, angles)

	// try 90 degrees rotated
	rotated := img.Rotate90()
	vScore, vAngle := GetAngleInner2(rotated, angles)

	if hScore > vScore {
		return hAngle
	}
	return -90 - vAngle
}

// run horizontal lines across the image, at all test angles, and pick the angle
// where we get the most uninterrupted lines (pure white)
func GetAngleInner2(img *Image, angles []float64) (score, angle float64) {

	padX := img.Width / 12  // arbitrary padding, but we stay away from the edges to avoid binder holes from affecting us
	padY := img.Height / 10 // Padding to prevent us overflowing the image, which we will do, because our lines are rotated

	linesAtAngle := make([]int, len(angles))
	x := padX

	// 0..255 threshold where we consider a pixel white
	// Note that this algorithm tends to work even if this threshold is very low.
	pixelIsWhiteThreshold := 200

	// Proportion of pixels that must be white for us to consider the line white
	lineIsWhiteThreshold := 0.995

	// Keep lowering our "is this line white" threshold until we get a reasonable number of white lines.
	//for iter := 0; iter < 10; iter++ {

	// The maximum number of white lines we expect to see in the image.
	totalLineCount := img.Height - padY*2

	//targetLines := int(0.05 * float64(totalLineCount))
	for y := padY; y < img.Height-padY; y++ {
		for i := range angles {
			isWhite := IsWhite(*img, x, y, img.Width-padX*2, angles[i], pixelIsWhiteThreshold, lineIsWhiteThreshold)
			if isWhite {
				linesAtAngle[i] = linesAtAngle[i] + 1
			}
		}
	}
	maxWhiteLines := 0
	for _, lines := range linesAtAngle {
		maxWhiteLines = max(maxWhiteLines, lines)
	}
	//if maxWhiteLines > targetLines {
	//	break
	//}
	//}
	bestLines := 0
	bestAngle := 0.0 // degrees
	for i, lines := range linesAtAngle {
		degrees := angles[i]
		//fmt.Printf("angle %.1f degrees: %v\n", degrees, lines)
		if lines > bestLines {
			bestAngle = degrees
			bestLines = lines
		}
	}
	return float64(bestLines) / float64(totalLineCount), bestAngle
}
