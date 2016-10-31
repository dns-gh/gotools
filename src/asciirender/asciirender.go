package main

import (
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/ioutil"
	"math"
	"os"
	"strconv"
	"strings"
)

type asciiData struct {
	ncols        int
	nrows        int
	xllcorner    float64
	yllcorner    float64
	CELLSIZE     float64
	NODATA_VALUE int64
	asciiTab     [][]float64
}

// http://en.wikipedia.org/wiki/HSL_and_HSV
func hslToRgb(h, s, l float64) (r, g, b float64) {
	C := (1 - math.Abs(2*l-1)) * s
	hprime := h / 60
	X := C * (1 - math.Abs(math.Mod(hprime, 2.0)-1))
	var r1, g1, b1 float64
	switch {
	case 0 <= hprime && hprime < 1:
		r1, g1, b1 = C, X, 0
	case 1 <= hprime && hprime < 2:
		r1, g1, b1 = X, C, 0
	case 2 <= hprime && hprime < 3:
		r1, g1, b1 = 0, C, X
	case 3 <= hprime && hprime < 4:
		r1, g1, b1 = 0, X, C
	case 4 <= hprime && hprime < 5:
		r1, g1, b1 = X, 0, C
	case 5 <= hprime && hprime < 6:
		r1, g1, b1 = C, 0, X
	}
	m := l - 0.5*C
	r, g, b = r1+m, g1+m, b1+m
	return
}

func getRGB(v float64) (r, g, b float64) {
	h := 360 * v
	s := 1.0
	l := 0.5 * v
	r, g, b = hslToRgb(h, s, l)
	return
}

// Get an rgba color from v in [0,1]
func getColor(v float64) color.RGBA {
	r, g, b := getRGB(v)
	return color.RGBA{uint8(255 * r), uint8(255 * g), uint8(255 * b), 255}
}

func splitMD(s string) []string {
	return strings.FieldsFunc(s, func(r rune) bool {
		switch r {
		case ' ', '\r', '\n':
			return true
		}
		return false
	})
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}

func readLines(path string) []string {

	buff, err := ioutil.ReadFile(path)
	check(err)
	s := string(buff)
	return strings.Split(s, "\n")
}

func main() {
	// Read ASCII data
	// TODO
	var lines []string
	lines = readLines("./testdata/asciidata/test.asc")

	numLines := 0
	for _ = range lines {
		numLines += 1
	}
	fmt.Println("numlines : ", numLines)
	for idx, line := range lines {
		if idx < 6 {
			fmt.Println(line)
		}
	}

	var rdata asciiData
	var splitedRow []string
	//var tempValue int64
	splitedRow = splitMD(lines[0])
	tempValue, err := strconv.ParseInt(splitedRow[1], 10, 0)
	rdata.ncols = int(tempValue)
	splitedRow = splitMD(lines[1])
	tempValue, err = strconv.ParseInt(splitedRow[1], 10, 0)
	rdata.nrows = int(tempValue)

	splitedRow = splitMD(lines[2])
	rdata.xllcorner, err = strconv.ParseFloat(splitedRow[1], 16)
	splitedRow = splitMD(lines[3])
	rdata.yllcorner, _ = strconv.ParseFloat(splitedRow[1], 16)
	splitedRow = splitMD(lines[4])
	rdata.CELLSIZE, _ = strconv.ParseFloat(splitedRow[1], 16)
	splitedRow = splitMD(lines[5])
	rdata.NODATA_VALUE, _ = strconv.ParseInt(splitedRow[1], 10, 0)

	fmt.Println(rdata.ncols)
	fmt.Println(rdata.nrows)
	fmt.Println(rdata.xllcorner)
	fmt.Println(rdata.yllcorner)
	fmt.Println(rdata.CELLSIZE)
	fmt.Println(rdata.NODATA_VALUE)

	splitData := splitMD(lines[6])
	for i := 7; i < numLines; i++ {
		splitedLine := splitMD(lines[i])
		for j := 0; j < len(splitedLine); j++ {
			splitData = append(splitData, splitedLine[j])
		}
	}

	asciiTab := make([][]float64, 1, 10000)
	for i := 0; i < rdata.nrows; i++ {
		asciiTab[i] = make([]float64, 0)
		for j := 0; j < rdata.ncols; j++ {
			var nextFloat float64
			nextFloat, _ = strconv.ParseFloat(splitData[i*rdata.ncols+j], 16)
			asciiTab[i] = append(asciiTab[i], nextFloat)
		}
		asciiTab = append(asciiTab, make([]float64, 0))
	}

	for i := 0; i < rdata.nrows; i++ {
		if i > 10 {
			break
		}
		for j := 0; j < rdata.ncols; j++ {
			if j > 10 {
				break
			}
			fmt.Print(asciiTab[i][j], " ")
		}
		fmt.Println("")
	}

	// Get min and max val of the valueTab
	minVal := asciiTab[0][0]
	maxVal := asciiTab[0][0]
	for i := 0; i < rdata.nrows; i++ {
		for j := 0; j < rdata.ncols; j++ {
			if asciiTab[i][j] < minVal {
				minVal = asciiTab[i][j]
			}
			if asciiTab[i][j] > maxVal {
				maxVal = asciiTab[i][j]
			}
		}
	}

	// Create a PNG file from data
	imageFile, err := os.Create("test.png")
	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}

	rgbaData := image.NewNRGBA(image.Rectangle{Min: image.Point{0, 0}, Max: image.Point{rdata.ncols, rdata.nrows}})
	for i := 0; i < rdata.ncols; i++ {
		for j := 0; j < rdata.nrows; j++ {
			r, g, b := getRGB((asciiTab[j][i] - minVal) / (maxVal - minVal))
			rgbaData.SetNRGBA(i, j, color.NRGBA{uint8(255 * r), uint8(255 * g), uint8(255 * b), 255})
		}
	}

	if err = png.Encode(imageFile, rgbaData); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
