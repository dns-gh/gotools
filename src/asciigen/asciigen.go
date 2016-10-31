package main

import (
	"bufio"
	"flag"
	"fmt"
	"math"
	"os"
	s "strings"
)

// TODO
// add generation of nodata values of different types : bands, noise, holes
// add filepath management/option/flag

func main() {

	//relativeFilePathPtr := flag.String("rpath", "", "relative file path to the test dir")
	xstepPtr := flag.Float64("xstep", math.Pi/2, "x step of the grid")
	ystepPtr := flag.Float64("ystep", math.Pi/2, "y step of the grid")
	freqPtr := flag.Float64("freq", 1.0, "sinusoidal frequency at the begining of the generation")
	freqEndPtr := flag.Float64("lastfreq", 1.0, "sinusoidal frequency at the end of the generation")
	xllPtr := flag.Float64("xll", 0, "xllcorner value")
	yllPtr := flag.Float64("yll", 0, "yllcorner value")
	cellsizePtr := flag.Float64("cell", 100, "cellsize")
	nodataPtr := flag.Int64("nodata", -30000, "nodata value")
	verbosePtr := flag.Bool("v", false, "verbose mode")
	filenamePtr := flag.String("filename", "sinusGrid.asc", "name of the data file (.asc)")
	nodatatypePtr := flag.String("nodatatype", "none", "type of the nodata region")
	nodatamoduloPtr := flag.Int64("modulo", 2, "nodata band modulo")
	//randNoisePtr := flag.Float64("noise", 0.0, "randomn noise from 0 to 1 (completly noised)")

	flag.Parse()
	fmt.Println("Generating ASCII grid...")

	tabSizeX := int(2.0 * math.Pi / *xstepPtr)
	tabSizeY := int(2.0 * math.Pi / *ystepPtr)
	dataTab := make([][]float64, 1, 1000)

	valueTab := make([]float64, 0, 1000)
	for i := 0; i < tabSizeX; i++ {
		valueTab = append(valueTab, *xstepPtr*float64(i))
	}

	// linear interpolation between the first frequence and the last
	linearPoly := make([]float64, 0, 1000)
	for i := 0; i < tabSizeX; i++ {
		linearPoly = append(linearPoly, float64(i)/float64(tabSizeX-1))
	}

	// fill the grid
	for i := 0; i < tabSizeX; i++ {
		dataTab[i] = make([]float64, 0)

		for j := 0; j < tabSizeY; j++ {
			if *nodatatypePtr != string("none") && j%int(*nodatamoduloPtr) == 0 {
				dataTab[i] = append(dataTab[i], float64(*nodataPtr))
			} else {
				dataTab[i] = append(dataTab[i], 500*math.Sin(valueTab[i]*(linearPoly[i]**freqEndPtr+(1-linearPoly[i])**freqPtr))+500)
			}
		}
		dataTab = append(dataTab, make([]float64, 0))
	}

	if *verbosePtr {
		fmt.Println(" - x step : ", *xstepPtr)
		fmt.Println(" - y step : ", *ystepPtr)
		fmt.Println(" - freq : ", *freqPtr)
		fmt.Println(" - freqEnd : ", *freqEndPtr)
		fmt.Println(" - ncols : ", tabSizeX)
		fmt.Println(" - nrows : ", tabSizeY)
		fmt.Println(" - xllcorner ", *xllPtr)
		fmt.Println(" - yllcorner ", *yllPtr)
		fmt.Println(" - CELLSIZE ", *cellsizePtr)
		fmt.Println(" - NODATA_VALUE ", *nodataPtr)
		fmt.Println(" - nodata type ", *nodatatypePtr)
		fmt.Println(" - save filepath", *filenamePtr)
	}

	// create file to save results in ascii format
	fmt.Println("Creating file... ")
	filePath := s.Join([]string{"./", *filenamePtr}, "")
	f, err := os.Create(filePath)
	check(err)
	defer f.Close()
	fmt.Println("File created at ", filePath)

	w := bufio.NewWriter(f)

	fmt.Fprintf(w, "ncols %d\n", tabSizeY)
	fmt.Fprintf(w, "nrows %d\n", tabSizeX)
	fmt.Fprintf(w, "xllcorner %6.2f\n", *xllPtr)
	fmt.Fprintf(w, "yllcorner %6.2f\n", *yllPtr)
	fmt.Fprintf(w, "CELLSIZE %6.2f\n", *cellsizePtr)
	fmt.Fprintf(w, "NODATA_VALUE %d\n", *nodataPtr)

	for i := 0; i < tabSizeX; i++ {
		for j := 0; j < tabSizeY; j++ {
			fmt.Fprintf(w, "%6.2f ", dataTab[i][j])
		}
		fmt.Fprintf(w, "\n")
	}

	w.Flush()

	fmt.Println("ASCII grid generated")
}

func check(e error) {
	if e != nil {
		panic(e)
	}
}
