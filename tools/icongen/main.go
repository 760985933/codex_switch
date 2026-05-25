package main

import (
	"flag"
	"fmt"
	"image"
	"image/png"
	"os"
	"path/filepath"

	"github.com/srwiley/oksvg"
	"github.com/srwiley/rasterx"
)

func main() {
	inPath := flag.String("in", "", "Input SVG path")
	outPath := flag.String("out", "", "Output PNG path")
	size := flag.Int("size", 1024, "Output square image size (px)")
	flag.Parse()

	if *inPath == "" || *outPath == "" {
		_, _ = fmt.Fprintln(os.Stderr, "Usage: icongen -in <svg> -out <png> [-size 1024]")
		os.Exit(2)
	}
	if *size <= 0 {
		_, _ = fmt.Fprintln(os.Stderr, "size must be > 0")
		os.Exit(2)
	}

	inFile, err := os.Open(*inPath)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer inFile.Close()

	icon, err := oksvg.ReadIconStream(inFile)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	w := float64(*size)
	h := float64(*size)
	icon.SetTarget(0, 0, w, h)

	img := image.NewRGBA(image.Rect(0, 0, *size, *size))
	scanner := rasterx.NewScannerGV(*size, *size, img, img.Bounds())
	raster := rasterx.NewDasher(*size, *size, scanner)
	icon.Draw(raster, 1.0)

	if mkErr := os.MkdirAll(filepath.Dir(*outPath), 0o755); mkErr != nil {
		_, _ = fmt.Fprintln(os.Stderr, mkErr)
		os.Exit(1)
	}

	outFile, err := os.Create(*outPath)
	if err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer outFile.Close()

	if err := png.Encode(outFile, img); err != nil {
		_, _ = fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
