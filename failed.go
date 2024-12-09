package docangle

import (
	"fmt"
	"slices"
)

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

// Returns the max delta between two rotated scanlines
func analyzeBlock(img Image, x, y, width, height int, angle float64) int {
	ls := SetupLine(width, angle)
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

func GetAngleBlocks(img *Image) float64 {
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
					maxDelta := analyzeBlock(*img, x, y, blockWidth, blockHeight, float64(angle)/angleUnit)
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
