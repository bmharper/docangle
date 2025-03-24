package main

import (
	"fmt"
	"math"
	"os"

	"github.com/bmharper/cimg/v2"
	"github.com/bmharper/docangle"
)

func check(err error) {
	if err != nil {
		panic(err)
	}
}

func main() {
	inputFilename := os.Args[1]
	outputFilename := os.Args[2]
	org, err := cimg.ReadFile(inputFilename)
	check(err)

	gray := org.ToGray()
	orgImg := docangle.Image{
		Width:  gray.Width,
		Height: gray.Height,
		Pixels: gray.Pixels,
	}

	params := docangle.NewWhiteLinesParams()
	params.Include90Degrees = true
	_, angle := docangle.GetAngleWhiteLines(&orgImg, params)

	var dst *cimg.Image
	finalAngle := -angle
	if math.Abs(angle) > 80 {
		dst = cimg.NewImage(org.Height, org.Width, org.Format)
	} else {
		dst = cimg.NewImage(org.Width, org.Height, org.Format)
	}
	cimg.Rotate(org, dst, finalAngle*math.Pi/180, nil)
	dst.WriteJPEG(outputFilename, cimg.MakeCompressParams(cimg.Sampling444, 95, 0), 0644)

	fmt.Printf("%.1f\n", finalAngle)
}
