package main

import (
	"encoding/xml"
	"flag"
	"fmt"
	"image"
	"image/png"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

type VectorPath struct {
	PathData  string `xml:"pathData,attr"`
	FillColor string `xml:"fillColor,attr"`
	FillType  string `xml:"fillType,attr"`
}

type Group struct {
	Paths []VectorPath `xml:"path"`
}

type AndroidVectorDrawable struct {
	XMLName xml.Name     `xml:"vector"`
	Width   float64      `xml:"viewportWidth,attr"`
	Height  float64      `xml:"viewportHeight,attr"`
	Paths   []VectorPath `xml:"path"`
	Groups  []Group      `xml:"group"`
}

var (
	inputFile  = flag.String("input", "", "Input android vector xml file")
	outputFile = flag.String("output", "", "Output png file")
)

// Convert Android color format to SVG color format
func convertColor(androidColor string) string {
	if len(androidColor) == 9 {
		// Remove the alpha channel
		return "#" + androidColor[3:]
	}
	return androidColor
}

func changeExtension(filename, newExt string) string {
	ext := filepath.Ext(filename)
	return strings.TrimSuffix(filename, ext) + newExt
}

func main() {

	flag.Parse()
	args := flag.Args()

	if *inputFile == "" && len(args) >= 1 {
		*inputFile = args[0]
	}
	if *outputFile == "" && len(args) >= 2 {
		*outputFile = args[1]
	}
	if *outputFile == "" {
		*outputFile = changeExtension(*inputFile, ".png")
	}

	if *inputFile == "" {
		fmt.Println("Missing input file name (android xml vector drawable).")
		os.Exit(11)
		return
	}

	// Load Android Vector Drawable XML
	xmlFile, err := os.Open(*inputFile)
	if err != nil {
		fmt.Println("Error opening file:", err)
		os.Exit(22)
	}
	defer xmlFile.Close()

	byteValue, _ := ioutil.ReadAll(xmlFile)

	var drawable AndroidVectorDrawable
	xml.Unmarshal(byteValue, &drawable)

	// Create SVG
	svg := fmt.Sprintf(`<svg viewBox="0 0 %.1f %.1f" xmlns="http://www.w3.org/2000/svg">`, drawable.Width, drawable.Height)

	for _, path := range drawable.Paths {
		fillColor := convertColor(path.FillColor)
		svg += fmt.Sprintf(`<path d="%s" fill="%s"/>`, path.PathData, fillColor)
	}

	for _, group := range drawable.Groups {
		svg += "<g>"
		for _, path := range group.Paths {
			fillColor := convertColor(path.FillColor)
			svg += fmt.Sprintf(`<path d="%s" fill="%s"/>`, path.PathData, fillColor)
		}
		svg += "</g>"
	}

	svg += "</svg>"

	// Save SVG
	var svgFilename = changeExtension(*outputFile, ".svg")
	ioutil.WriteFile(svgFilename, []byte(svg), 0644)

	svg2png(svgFilename, *outputFile)
}

func svg2png(svgFile, outFile string) {
	w, h := 512, 512

	in, err := os.Open(svgFile)
	if err != nil {
		panic(err)
	}
	defer in.Close()

	icon, _ := oksvg.ReadIconStream(in)
	icon.SetTarget(0, 0, float64(w), float64(h))
	rgba := image.NewRGBA(image.Rect(0, 0, w, h))
	icon.Draw(rasterx.NewDasher(w, h, rasterx.NewScannerGV(w, h, rgba, rgba.Bounds())), 1)

	out, err := os.Create(outFile)
	if err != nil {
		panic(err)
	}
	defer out.Close()

	err = png.Encode(out, rgba)
	if err != nil {
		panic(err)
	}
}
