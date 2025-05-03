package main

import (
	"bufio"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/draw"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"log"
	"math"
	"os"
	"path/filepath"
	"strings"

	"github.com/disintegration/imaging"
	dither "github.com/makeworld-the-better-one/dither/v2"
	"github.com/pterm/pterm"
	"golang.org/x/image/bmp"
)

// Colors for waveshare's photo painter B
var einkColors = []color.RGBA{
	{0, 0, 0, 255},       // Black
	{255, 255, 255, 255}, // White
	{0, 255, 0, 255},     // Green
	{0, 0, 255, 255},     // Blue
	{255, 0, 0, 255},     // Red
	{255, 255, 0, 255},   // Yellow
}

const (
	TARGET_SHORT_SIDE = 480
	TARGET_LONG_SIDE  = 800
)

func main() {
	var inDir, outDir string

	flag.StringVar(&inDir, "in-dir", "", "Input directory containing images")
	flag.StringVar(&outDir, "out-dir", "", "Output directory for converted images")
	flag.Parse()

	// flags
	reader := bufio.NewReader(os.Stdin)
	if inDir == "" {
		fmt.Print("Enter input directory: ")
		inDir, _ = reader.ReadString('\n')
		inDir = strings.TrimSpace(inDir)
	}

	if outDir == "" {
		fmt.Print("Enter output directory: ")
		outDir, _ = reader.ReadString('\n')
		outDir = strings.TrimSpace(outDir)
	}

	// create dir if needed
	if err := os.MkdirAll(outDir, 0755); err != nil {
		log.Fatalf("Failed to create output directory: %v", err)
	}

	processImages(inDir, outDir)
}

func processImages(inDir, outDir string) {
	entries, err := os.ReadDir(inDir)
	if err != nil {
		log.Fatalf("Error reading input directory: %v", err)
	}

	palette := make([]color.Color, len(einkColors))
	for i, c := range einkColors {
		palette[i] = c
	}

	// Create ditherer
	d := dither.NewDitherer(palette)
	d.Matrix = dither.FloydSteinberg
	d.Serpentine = true

	p, _ := pterm.DefaultProgressbar.WithTotal(len(entries)).WithTitle("Converting images").Start()
	for _, entry := range entries {
		if entry.IsDir() {
			continue
		}

		filename := entry.Name()
		ext := strings.ToLower(filepath.Ext(filename))
		if ext != ".jpg" && ext != ".jpeg" && ext != ".png" {
			continue // Skip non-image files
		}
		p.UpdateTitle("Converting " + filename)

		// Load image
		imgPath := filepath.Join(inDir, filename)
		img, err := loadImage(imgPath)
		if err != nil {
			log.Printf("Error loading image %s: %v", imgPath, err)
			continue
		}

		resized := resizeAndCropImage(img)

		// dither
		dithered := d.DitherCopy(resized)

		// output
		outputFilename := strings.TrimSuffix(filename, ext) + ".eink6c.bmp"
		outPath := filepath.Join(outDir, outputFilename)

		if err := saveBMP(dithered, outPath); err != nil {
			log.Printf("Error saving image %s: %v", outPath, err)
			continue
		}

		// terminal output
		pterm.Printf("Converted %s\n", filename)
		p.Increment()
	}

	pterm.Success.Printf("Converted %d photos from:\n  %s\nto:\n  %s\n\n", len(entries), inDir, outDir)
}

func resizeAndCropImage(img image.Image) *image.NRGBA {
	bounds := img.Bounds()
	orgWidth := bounds.Dx()
	orgHeight := bounds.Dy()

	var targetWidth, targetHeight int
	if orgWidth > orgHeight {
		targetWidth = TARGET_LONG_SIDE
		targetHeight = TARGET_SHORT_SIDE
	} else {
		targetWidth = TARGET_SHORT_SIDE
		targetHeight = TARGET_LONG_SIDE
	}

	// Calculate scaling ratio to preserve aspect ratio
	widthRatio := float64(orgWidth) / float64(targetWidth)
	heightRatio := float64(orgHeight) / float64(targetHeight)
	ratio := math.Min(widthRatio, heightRatio)

	// Scale image
	newWidth := int(float64(orgWidth) / ratio)
	newHeight := int(float64(orgHeight) / ratio)
	scaledImg := imaging.Resize(img, newWidth, newHeight, imaging.Lanczos)

	// Center crop to target dimensions
	left := (newWidth - targetWidth) / 2
	top := (newHeight - targetHeight) / 2
	resized := imaging.Crop(scaledImg, image.Rect(left, top, left+targetWidth, top+targetHeight))
	return resized
}

func loadImage(path string) (image.Image, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	img, _, err := image.Decode(f)
	return img, err
}

func saveBMP(img image.Image, path string) error {
	f, err := os.Create(path)
	if err != nil {
		return err
	}
	defer f.Close()

	bounds := img.Bounds()
	rgba := image.NewRGBA(bounds)
	draw.Draw(rgba, bounds, img, bounds.Min, draw.Src)

	return bmp.Encode(f, rgba)
}
