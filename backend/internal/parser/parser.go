package parser

import (
	"fmt"
	"image"
	"image/png"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"

	"cloneheroer/internal/db"

	"github.com/otiai10/gosseract/v2"
	"golang.org/x/image/draw"
)

// Parser extracts score data from Clone Hero screenshot images.
type Parser struct {
	client    *gosseract.Client
	closed    bool
	maxWidth  int
	maxHeight int
}

// findTessdataPrefix attempts to find the Tesseract data directory.
func findTessdataPrefix() string {
	// Check if TESSDATA_PREFIX is already set
	if prefix := os.Getenv("TESSDATA_PREFIX"); prefix != "" {
		return prefix
	}

	// Common locations for tessdata
	commonPaths := []string{
		"/usr/share/tesseract-ocr/5/tessdata",
		"/usr/share/tesseract-ocr/4.00/tessdata",
		"/usr/share/tesseract-ocr/tessdata",
		"/usr/local/share/tesseract-ocr/5/tessdata",
		"/usr/local/share/tesseract-ocr/4.00/tessdata",
		"/usr/local/share/tesseract-ocr/tessdata",
		"/opt/homebrew/share/tesseract-ocr/5/tessdata", // macOS Homebrew
		"/opt/homebrew/share/tesseract-ocr/tessdata",   // macOS Homebrew
	}

	// Check each path for eng.traineddata
	for _, path := range commonPaths {
		engFile := filepath.Join(path, "eng.traineddata")
		if _, err := os.Stat(engFile); err == nil {
			// Return the parent directory (tessdata's parent)
			return filepath.Dir(path)
		}
	}

	// Try to query tesseract for tessdata location
	if out, err := exec.Command("tesseract", "--print-parameters").Output(); err == nil {
		// Look for tessdata path in output (format may vary)
		output := string(out)
		if strings.Contains(output, "tessdata") {
			// Try to extract path (this is a best-effort approach)
			lines := strings.Split(output, "\n")
			for _, line := range lines {
				if strings.Contains(line, "tessdata") {
					// Try to extract directory path
					parts := strings.Fields(line)
					for _, part := range parts {
						if strings.Contains(part, "tessdata") {
							// Return parent of tessdata directory
							if strings.HasSuffix(part, "tessdata") {
								return filepath.Dir(part)
							}
						}
					}
				}
			}
		}
	}

	return ""
}

// NewParser creates a new parser instance.
func NewParser(maxWidth, maxHeight int) (*Parser, error) {
	// Set TESSDATA_PREFIX if not already set
	if os.Getenv("TESSDATA_PREFIX") == "" {
		prefix := findTessdataPrefix()
		if prefix != "" {
			os.Setenv("TESSDATA_PREFIX", prefix)
			log.Printf("Set TESSDATA_PREFIX to: %s", prefix)
		} else {
			log.Printf("warning: TESSDATA_PREFIX not set and could not be auto-detected. Tesseract may fail to find language data files.")
			log.Printf("Please set TESSDATA_PREFIX environment variable to the directory containing the 'tessdata' folder.")
			log.Printf("Example: export TESSDATA_PREFIX=/usr/share/tesseract-ocr/5")
		}
	}

	client := gosseract.NewClient()
	if err := client.SetLanguage("eng"); err != nil {
		return nil, fmt.Errorf("failed to set OCR language (check TESSDATA_PREFIX): %w", err)
	}
	return &Parser{
		client:    client,
		maxWidth:  maxWidth,
		maxHeight: maxHeight,
	}, nil
}

// Close releases resources.
func (p *Parser) Close() error {
	if p.closed {
		return nil // Already closed, no-op
	}
	if p.client != nil {
		err := p.client.Close()
		p.closed = true
		return err
	}
	p.closed = true
	return nil
}

// ParseImage extracts score data from a screenshot image file.
func (p *Parser) ParseImage(imagePath string) (*db.CreateScoreData, error) {
	// Parse timestamp from filename
	// Format: clonehero-Artist-20251212052231.png or 20251212052231.png
	filename := filepath.Base(imagePath)
	createdAt, err := parseTimestampFromFilename(filename)
	if err != nil {
		// If we can't parse timestamp, use file modification time
		info, err := os.Stat(imagePath)
		if err != nil {
			return nil, fmt.Errorf("failed to get file info: %w", err)
		}
		createdAt = info.ModTime()
	}

	// Load and preprocess image
	img, err := p.loadImage(imagePath)
	if err != nil {
		return nil, fmt.Errorf("failed to load image: %w", err)
	}

	// Extract text from different regions
	artist, songName, charter := p.extractTopLeftInfo(img)
	totalScore, stars := p.extractCenterInfo(img)
	players := p.extractPlayers(img)

	// Check if OCR failed to extract meaningful data
	hasData := false
	var missingFields []string

	if strings.TrimSpace(artist) != "" {
		hasData = true
	} else {
		missingFields = append(missingFields, "artist")
	}

	if strings.TrimSpace(songName) != "" {
		hasData = true
	} else {
		missingFields = append(missingFields, "song name")
	}

	if totalScore != nil {
		hasData = true
	} else {
		missingFields = append(missingFields, "total score")
	}

	if len(players) > 0 {
		hasData = true
	} else {
		missingFields = append(missingFields, "players")
	}

	// Output prominent warning if critical data is missing
	if !hasData || len(missingFields) >= 3 {
		const red = "\033[31m"
		const reset = "\033[0m"
		const bold = "\033[1m"

		log.Printf("%s%s════════════════════════════════════════════════════════════════════════════════%s", red, bold, reset)
		log.Printf("%s%s⚠️  OCR WARNING: Failed to extract meaningful data from image%s", red, bold, reset)
		log.Printf("%s%s════════════════════════════════════════════════════════════════════════════════%s", red, bold, reset)
		log.Printf("%s%s  File: %s%s", red, bold, imagePath, reset)
		log.Printf("%s%s  Missing fields: %s%s", red, bold, strings.Join(missingFields, ", "), reset)
		log.Printf("%s%s  Extracted - Artist: '%s' | Song: '%s' | Score: %v | Stars: %v | Players: %d%s",
			red, bold, artist, songName, totalScore, stars, len(players), reset)
		log.Printf("%s%s  This may result in empty database records. Check OCR/Tesseract installation and image quality.%s", red, bold, reset)
		log.Printf("%s%s════════════════════════════════════════════════════════════════════════════════%s", red, bold, reset)
	}

	return &db.CreateScoreData{
		Artist:        artist,
		SongName:      songName,
		Charter:       charter,
		TotalScore:    totalScore,
		StarsAchieved: stars,
		Players:       players,
		CreatedAt:     createdAt,
	}, nil
}

// parseTimestampFromFilename extracts timestamp from filename.
// Supports formats: "20251212052231" or "clonehero-Artist-20251212052231.png"
func parseTimestampFromFilename(filename string) (time.Time, error) {
	// Remove extension
	filename = strings.TrimSuffix(filename, filepath.Ext(filename))

	// Try to find 14-digit timestamp (yyyyMMddHHmmss)
	re := regexp.MustCompile(`(\d{14})`)
	matches := re.FindStringSubmatch(filename)
	if len(matches) < 2 {
		return time.Time{}, fmt.Errorf("no timestamp found in filename")
	}

	timestamp := matches[1]
	layout := "20060102150405"
	return time.Parse(layout, timestamp)
}

// loadImage loads a PNG image file and returns it as an image.Image.
// It resizes the image if it exceeds the configured maximum dimensions.
func (p *Parser) loadImage(path string) (image.Image, error) {
	// Only accept PNG files
	ext := strings.ToLower(filepath.Ext(path))
	if ext != ".png" {
		return nil, fmt.Errorf("only PNG files are supported, got: %s", ext)
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	// Decode PNG
	img, err := png.Decode(file)
	if err != nil {
		return nil, fmt.Errorf("failed to decode PNG: %w", err)
	}

	// Resize if too large (OCR works better on smaller images)
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// If image is larger than max dimensions, scale it down proportionally
	if p.maxWidth > 0 && p.maxHeight > 0 && (width > p.maxWidth || height > p.maxHeight) {
		scale := float64(p.maxWidth) / float64(width)
		if float64(p.maxHeight)/float64(height) < scale {
			scale = float64(p.maxHeight) / float64(height)
		}
		newWidth := int(float64(width) * scale)
		newHeight := int(float64(height) * scale)

		resized := image.NewRGBA(image.Rect(0, 0, newWidth, newHeight))
		draw.BiLinear.Scale(resized, resized.Bounds(), img, img.Bounds(), draw.Src, nil)
		img = resized
	}

	return img, nil
}

// extractTopLeftInfo extracts artist, song name, and charter from top left of image.
func (p *Parser) extractTopLeftInfo(img image.Image) (artist, songName string, charter *string) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Top left region: approximately first 30% of width, first 20% of height
	left := 0
	top := 0
	right := width * 30 / 100
	bottom := height * 20 / 100

	region := cropImage(img, left, top, right, bottom)

	// Debug: log region size
	regionBounds := region.Bounds()
	if regionBounds.Dx() == 0 || regionBounds.Dy() == 0 {
		log.Printf("warning: extracted region is empty (top-left: %dx%d)", regionBounds.Dx(), regionBounds.Dy())
		return "", "", nil
	}

	text := p.extractText(region)
	if text == "" {
		log.Printf("warning: OCR returned empty text for top-left region (%dx%d)", regionBounds.Dx(), regionBounds.Dy())
	}

	lines := strings.Split(strings.TrimSpace(text), "\n")
	lines = filterEmpty(lines)

	// Debug: log extracted lines to help diagnose issues
	log.Printf("extracted top-left OCR lines (%d): %q", len(lines), lines)

	if len(lines) > 0 {
		artist = strings.TrimSpace(lines[0])
	}
	if len(lines) > 1 {
		songName = strings.TrimSpace(lines[1])
	}
	if len(lines) > 2 {
		c := strings.TrimSpace(lines[2])
		// Check if line contains "charter" keyword
		if strings.HasPrefix(strings.ToLower(c), "charter") {
			// Remove "charter:" prefix if present
			c = strings.TrimPrefix(strings.TrimPrefix(c, "Charter:"), "charter:")
			c = strings.TrimSpace(c)
		}
		if c != "" {
			charter = &c
		}
	}

	// Validation: If songName is empty but artist is set, there might be an issue
	// This could mean OCR only extracted one line, and we're not sure if it's artist or song
	if artist != "" && songName == "" {
		log.Printf("warning: extracted artist '%s' but no song name. OCR may have only detected one line.", artist)
	}

	return
}

// extractCenterInfo extracts total score and stars from center top of image.
func (p *Parser) extractCenterInfo(img image.Image) (totalScore *int64, stars *int) {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Center region: middle 40% of width, first 25% of height
	left := width * 30 / 100
	top := 0
	right := width * 70 / 100
	bottom := height * 25 / 100

	region := cropImage(img, left, top, right, bottom)
	text := p.extractText(region)

	// Look for large numbers (total score) and star indicators
	lines := strings.Split(strings.TrimSpace(text), "\n")
	lines = filterEmpty(lines)

	// First large number is likely total score
	for _, line := range lines {
		// Remove non-digit characters except commas
		cleaned := regexp.MustCompile(`[^\d]`).ReplaceAllString(line, "")
		if cleaned != "" {
			if val, err := strconv.ParseInt(cleaned, 10, 64); err == nil {
				totalScore = &val
				break
			}
		}
	}

	// Look for star count (usually "★" or "Stars: X" or just a number)
	for _, line := range lines {
		lower := strings.ToLower(line)
		if strings.Contains(lower, "star") {
			re := regexp.MustCompile(`(\d+)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				if val, err := strconv.Atoi(matches[1]); err == nil {
					stars = &val
					break
				}
			}
		}
		// Also check for just a single digit (common pattern)
		if len(strings.TrimSpace(line)) == 1 {
			if val, err := strconv.Atoi(strings.TrimSpace(line)); err == nil && val >= 0 && val <= 7 {
				stars = &val
				break
			}
		}
	}

	return
}

// extractPlayers extracts player data from the main area of the image.
// TODO: Instrument detection currently relies on OCR text. For better accuracy,
// implement template matching using instrum-icons.png to detect instrument icons.
// This would require loading the reference icons and using image comparison/template matching.
func (p *Parser) extractPlayers(img image.Image) []db.Player {
	bounds := img.Bounds()
	width := bounds.Dx()
	height := bounds.Dy()

	// Main player area: full width, from 25% to 90% of height
	left := 0
	top := height * 25 / 100
	right := width
	bottom := height * 90 / 100

	region := cropImage(img, left, top, right, bottom)
	text := p.extractText(region)

	// Parse player data from text
	// This is a simplified parser - may need refinement based on actual screenshot format
	players := []db.Player{}
	lines := strings.Split(strings.TrimSpace(text), "\n")
	lines = filterEmpty(lines)

	// Group lines into player blocks (heuristic: look for patterns)
	currentPlayer := db.Player{}
	for i, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Try to identify player name (usually first non-numeric line in a block)
		if currentPlayer.Name == "" && !isNumeric(line) {
			currentPlayer.Name = line
			continue
		}

		// Look for difficulty (Easy, Medium, Hard, Expert)
		if strings.Contains(strings.ToLower(line), "easy") ||
			strings.Contains(strings.ToLower(line), "medium") ||
			strings.Contains(strings.ToLower(line), "hard") ||
			strings.Contains(strings.ToLower(line), "expert") {
			d := normalizeDifficulty(line)
			currentPlayer.Difficulty = &d
			continue
		}

		// Look for score (large number)
		if isNumeric(line) {
			cleaned := regexp.MustCompile(`[^\d]`).ReplaceAllString(line, "")
			if len(cleaned) > 3 { // Score is usually a large number
				if val, err := strconv.ParseInt(cleaned, 10, 64); err == nil {
					if currentPlayer.Score == nil {
						currentPlayer.Score = &val
					}
				}
			}
		}

		// Look for accuracy (percentage)
		if strings.Contains(line, "%") {
			re := regexp.MustCompile(`(\d+\.?\d*)%`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				if val, err := strconv.ParseFloat(matches[1], 64); err == nil {
					currentPlayer.Accuracy = &val
				}
			}
		}

		// Look for misses
		if strings.Contains(strings.ToLower(line), "miss") {
			re := regexp.MustCompile(`(\d+)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				if val, err := strconv.Atoi(matches[1]); err == nil {
					currentPlayer.Misses = &val
				}
			}
		}

		// Look for combo/best streak
		if strings.Contains(strings.ToLower(line), "combo") ||
			strings.Contains(strings.ToLower(line), "streak") {
			re := regexp.MustCompile(`(\d+)`)
			matches := re.FindStringSubmatch(line)
			if len(matches) > 1 {
				if val, err := strconv.Atoi(matches[1]); err == nil {
					currentPlayer.Combo = &val
				}
			}
		}

		// If we've collected enough info or hit a separator, save player
		if currentPlayer.Name != "" && (i == len(lines)-1 || isPlayerSeparator(line)) {
			if currentPlayer.Name != "" {
				players = append(players, currentPlayer)
			}
			currentPlayer = db.Player{}
		}
	}

	// Add last player if exists
	if currentPlayer.Name != "" {
		players = append(players, currentPlayer)
	}

	return players
}

// extractText performs OCR on an image region.
func (p *Parser) extractText(img image.Image) string {
	// Check if image is valid
	bounds := img.Bounds()
	if bounds.Dx() == 0 || bounds.Dy() == 0 {
		log.Printf("warning: extractText called with empty image")
		return ""
	}

	// Save image to temp file for OCR
	tmpFile, err := os.CreateTemp("", "clonehero-ocr-*.png")
	if err != nil {
		log.Printf("failed to create temp file for OCR: %v", err)
		return ""
	}
	tmpPath := tmpFile.Name()
	defer os.Remove(tmpPath)

	// Encode image to PNG
	if err := encodePNG(tmpFile, img); err != nil {
		log.Printf("failed to encode image to PNG: %v", err)
		tmpFile.Close()
		return ""
	}

	// Close and flush the file before Tesseract reads it
	if err := tmpFile.Close(); err != nil {
		log.Printf("failed to close temp file: %v", err)
		return ""
	}

	// Verify file exists and has content
	fileInfo, err := os.Stat(tmpPath)
	if err != nil {
		log.Printf("failed to stat temp file: %v", err)
		return ""
	}
	if fileInfo.Size() == 0 {
		log.Printf("warning: encoded PNG file is empty")
		return ""
	}

	// Perform OCR - SetImage needs the file path
	if err := p.client.SetImage(tmpPath); err != nil {
		log.Printf("failed to set image for OCR: %v", err)
		return ""
	}

	text, err := p.client.Text()
	if err != nil {
		log.Printf("failed to extract text via OCR: %v", err)
		return ""
	}

	// Debug: log extracted text (first 100 chars to avoid spam)
	if text != "" {
		preview := text
		if len(preview) > 100 {
			preview = preview[:100] + "..."
		}
		log.Printf("OCR extracted text (%d bytes): %q", len(text), preview)
	} else {
		log.Printf("warning: OCR returned empty text for image %dx%d", bounds.Dx(), bounds.Dy())
	}

	return text
}

// Helper functions

func cropImage(img image.Image, left, top, right, bottom int) image.Image {
	bounds := img.Bounds()
	if left < bounds.Min.X {
		left = bounds.Min.X
	}
	if top < bounds.Min.Y {
		top = bounds.Min.Y
	}
	if right > bounds.Max.X {
		right = bounds.Max.X
	}
	if bottom > bounds.Max.Y {
		bottom = bounds.Max.Y
	}

	cropped := image.NewRGBA(image.Rect(0, 0, right-left, bottom-top))
	draw.Draw(cropped, cropped.Bounds(), img, image.Pt(left, top), draw.Src)
	return cropped
}

func filterEmpty(lines []string) []string {
	var result []string
	for _, line := range lines {
		if strings.TrimSpace(line) != "" {
			result = append(result, line)
		}
	}
	return result
}

func isNumeric(s string) bool {
	re := regexp.MustCompile(`^\d+([,\s]\d+)*$`)
	return re.MatchString(strings.TrimSpace(s))
}

func normalizeDifficulty(s string) string {
	lower := strings.ToLower(s)
	if strings.Contains(lower, "expert") {
		return "Expert"
	}
	if strings.Contains(lower, "hard") {
		return "Hard"
	}
	if strings.Contains(lower, "medium") {
		return "Medium"
	}
	if strings.Contains(lower, "easy") {
		return "Easy"
	}
	return s
}

func isPlayerSeparator(s string) bool {
	// Heuristic: empty line or line with dashes/separators
	return strings.TrimSpace(s) == "" || strings.HasPrefix(strings.TrimSpace(s), "---")
}

func encodePNG(w *os.File, img image.Image) error {
	return png.Encode(w, img)
}
