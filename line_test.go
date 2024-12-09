package docangle

import "testing"

func TestLineIteration(t *testing.T) {
	ls := SetupLine(10, 1)
	IterateLineSubpixel(0, 0, 40, ls, func(x, y int, blend int32) {
		bf := float32(blend) / 65536
		t.Logf("x: %v, y: %v, blend: %.2f", x, y, bf)
	})
}
