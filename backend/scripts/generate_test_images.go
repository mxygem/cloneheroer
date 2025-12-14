package main

import (
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"

	"github.com/fogleman/gg"
)

const (
	width  = 1920
	height = 1080
)

func main() {
	// Get the testdata/images directory
	scriptDir, err := os.Getwd()
	if err != nil {
		fmt.Fprintf(os.Stderr, "Error getting working directory: %v\n", err)
		os.Exit(1)
	}

	// Navigate to project root (assuming we're in backend/scripts)
	projectRoot := filepath.Join(scriptDir, "..", "..")
	outputDir := filepath.Join(projectRoot, "testdata", "images")

	// Create directory if it doesn't exist
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		fmt.Fprintf(os.Stderr, "Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	fmt.Printf("Generating test images in: %s\n", outputDir)
	fmt.Println("------------------------------------------------------------")

	testImages := map[string]func() image.Image{
		"test-top-left.png":    createTestTopLeft,
		"test-center.png":      createTestCenter,
		"test-players.png":     createTestPlayers,
		"test-numbers.png":     createTestNumbers,
		"test-difficulty.png":  createTestDifficulty,
		"test-percentages.png": createTestPercentages,
		"test-empty.png":       createTestEmpty,
		"test-mixed.png":       createTestMixed,
		"test-stars.png":       createTestStars,
		"test-misses.png":      createTestMisses,
		"test-complete.png":    createTestComplete,
		"test-large.png":       createTestLarge,
	}

	for filename, createFunc := range testImages {
		img := createFunc()
		filepath := filepath.Join(outputDir, filename)
		if err := savePNG(filepath, img); err != nil {
			fmt.Printf("✗ Failed to create %s: %v\n", filename, err)
		} else {
			bounds := img.Bounds()
			fmt.Printf("✓ Created: %s (%dx%d)\n", filename, bounds.Dx(), bounds.Dy())
		}
	}

	// Create JPEG version
	jpegImg := createTestMixed()
	jpegPath := filepath.Join(outputDir, "test-image.jpg")
	if err := saveJPEG(jpegPath, jpegImg, 95); err != nil {
		fmt.Printf("✗ Failed to create test-image.jpg: %v\n", err)
	} else {
		fmt.Printf("✓ Created: test-image.jpg\n")
	}

	fmt.Println("------------------------------------------------------------")
	fmt.Printf("Generated %d test images\n", len(testImages)+1)
}

func createTestTopLeft() image.Image {
	dc := gg.NewContext(width, height)
	dc.SetRGB(0, 0, 0) // Black background
	dc.Clear()

	// Top-left region: 0-30% width, 0-20% height
	drawText(dc, 10, 40, "The Beatles", 40)
	drawText(dc, 10, 90, "Hey Jude", 40)
	drawText(dc, 10, 140, "Charter: Custom", 40)

	return dc.Image()
}

func createTestCenter() image.Image {
	dc := gg.NewContext(width, height)
	dc.SetRGB(0, 0, 0) // Black background
	dc.Clear()

	// Center region: 30-70% width, 0-25% height
	centerX := float64(width) / 2
	drawText(dc, centerX-200, 60, "1,234,567", 60)
	drawText(dc, centerX-100, 120, "Stars: 5", 40)

	return dc.Image()
}

func createTestPlayers() image.Image {
	dc := gg.NewContext(width, height)
	dc.SetRGB(0, 0, 0) // Black background
	dc.Clear()

	// Player area: 25% to 90% of height
	top := float64(height) * 25 / 100
	left := 50.0

	players := []string{
		"Player1",
		"Expert",
		"987,654",
		"98.5%",
		"Misses: 2",
		"Combo: 150",
		"",
		"Player2",
		"Hard",
		"876,543",
		"95.2%",
		"Misses: 5",
		"Combo: 120",
	}

	y := top + 20
	for _, line := range players {
		if line != "" {
			drawText(dc, left, y, line, 40)
		}
		y += 50
	}

	return dc.Image()
}

func createTestNumbers() image.Image {
	dc := gg.NewContext(width, height)
	dc.SetRGB(0, 0, 0) // Black background
	dc.Clear()

	numbers := []string{
		"1,234,567",
		"1 234 567",
		"1234567",
		"999,999,999",
	}

	y := 50.0
	for _, num := range numbers {
		drawText(dc, 50, y, num, 40)
		y += 60
	}

	return dc.Image()
}

func createTestDifficulty() image.Image {
	dc := gg.NewContext(width, height)
	dc.SetRGB(0, 0, 0) // Black background
	dc.Clear()

	difficulties := []string{
		"Easy",
		"EASY",
		"easy",
		"Medium",
		"Hard",
		"Expert",
		"Expert Mode",
	}

	y := 50.0
	for _, diff := range difficulties {
		drawText(dc, 50, y, diff, 40)
		y += 60
	}

	return dc.Image()
}

func createTestPercentages() image.Image {
	dc := gg.NewContext(width, height)
	dc.SetRGB(0, 0, 0) // Black background
	dc.Clear()

	percentages := []string{
		"98.5%",
		"100%",
		"87.23%",
		"Accuracy: 95.5%",
	}

	y := 50.0
	for _, pct := range percentages {
		drawText(dc, 50, y, pct, 40)
		y += 60
	}

	return dc.Image()
}

func createTestEmpty() image.Image {
	dc := gg.NewContext(width, height)
	dc.SetRGB(0, 0, 0) // Black background
	dc.Clear()
	return dc.Image()
}

func createTestMixed() image.Image {
	dc := gg.NewContext(width, height)
	dc.SetRGB(0, 0, 0) // Black background
	dc.Clear()

	content := []string{
		"Artist Name",
		"Song Title",
		"Score: 1,234,567",
		"Stars: 5",
		"Player1 - Expert - 98.5%",
		"Misses: 2",
		"Combo: 150",
	}

	y := 50.0
	for _, line := range content {
		drawText(dc, 50, y, line, 40)
		y += 60
	}

	return dc.Image()
}

func createTestStars() image.Image {
	dc := gg.NewContext(width, height)
	dc.SetRGB(0, 0, 0) // Black background
	dc.Clear()

	// Center region: 30-70% width, 0-25% height
	centerX := float64(width) / 2
	drawText(dc, centerX-100, 50, "Stars: 5", 40)
	drawText(dc, centerX-50, 110, "5", 40)
	drawText(dc, centerX-150, 170, "* * * * *", 40)

	return dc.Image()
}

func createTestMisses() image.Image {
	dc := gg.NewContext(width, height)
	dc.SetRGB(0, 0, 0) // Black background
	dc.Clear()

	// Main area: 25% to 90% of height
	top := float64(height) * 25 / 100
	y := top + 20

	misses := []string{
		"Misses: 3",
		"3 misses",
		"Miss Count: 5",
	}

	for _, line := range misses {
		drawText(dc, 50, y, line, 40)
		y += 60
	}

	return dc.Image()
}

func createTestComplete() image.Image {
	dc := gg.NewContext(width, height)
	dc.SetRGB(0, 0, 0) // Black background
	dc.Clear()

	// Top-left: Artist, Song, Charter
	drawText(dc, 10, 40, "The Beatles", 40)
	drawText(dc, 10, 90, "Hey Jude", 40)
	drawText(dc, 10, 140, "Charter: Custom", 40)

	// Center: Score and Stars
	centerX := float64(width) / 2
	drawText(dc, centerX-200, 60, "1,234,567", 60)
	drawText(dc, centerX-100, 120, "Stars: 5", 40)

	// Main area: Player data
	playerY := float64(height)*25/100 + 20
	players := []string{
		"Player1",
		"Expert",
		"987,654",
		"98.5%",
		"Misses: 2",
		"Combo: 150",
	}

	for _, line := range players {
		drawText(dc, 50, playerY, line, 40)
		playerY += 50
	}

	return dc.Image()
}

func createTestLarge() image.Image {
	// Large image to test resizing: 3840x2160
	largeWidth := 3840
	largeHeight := 2160
	dc := gg.NewContext(largeWidth, largeHeight)
	dc.SetRGB(0, 0, 0) // Black background
	dc.Clear()

	drawText(dc, 100, 120, "Large Test Image", 80)
	drawText(dc, 100, 220, "3840x2160", 60)

	return dc.Image()
}

// Helper functions

// drawText draws text using gg library with a simple built-in font
// Falls back to default font if system fonts aren't available
func drawText(dc *gg.Context, x, y float64, text string, fontSize float64) {
	dc.SetRGB(1, 1, 1) // White text

	// Try to load a common system font, fall back to default if not available
	fontPaths := []string{
		"/usr/share/fonts/truetype/dejavu/DejaVuSans.ttf",
		"/usr/share/fonts/truetype/liberation/LiberationSans-Regular.ttf",
		"/System/Library/Fonts/Helvetica.ttc",
		"/Library/Fonts/Arial.ttf",
	}

	// Try to load a system font, use default if none found
	for _, fontPath := range fontPaths {
		if err := dc.LoadFontFace(fontPath, fontSize); err == nil {
			break
		}
	}

	dc.DrawString(text, x, y)
}

func savePNG(filename string, img image.Image) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	return png.Encode(file, img)
}

func saveJPEG(filename string, img image.Image, quality int) error {
	file, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	return jpeg.Encode(file, img, &jpeg.Options{Quality: quality})
}
