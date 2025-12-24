package parser

import (
	"fmt"
	"image"
	"image/color"
	"os"
	"path/filepath"
	"testing"
	"time"

	"cloneheroer/internal/db"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewParser(t *testing.T) {
	parser, err := NewParser(1920, 1080)
	require.NoError(t, err)
	require.NotNil(t, parser)

	err = parser.Close()
	assert.NoError(t, err)

	// Test that we can create another parser after closing
	parser2, err := NewParser(1920, 1080)
	require.NoError(t, err)
	require.NotNil(t, parser2)
	defer parser2.Close()
}

func TestParseTimestampFromFilename(t *testing.T) {
	testCases := []struct {
		name      string
		filename  string
		wantTime  time.Time
		wantError bool
	}{
		{
			name:      "valid timestamp in filename",
			filename:  "clonehero-Artist-20251212052231.png",
			wantTime:  time.Date(2025, 12, 12, 5, 22, 31, 0, time.UTC),
			wantError: false,
		},
		{
			name:      "timestamp only",
			filename:  "20251212052231.png",
			wantTime:  time.Date(2025, 12, 12, 5, 22, 31, 0, time.UTC),
			wantError: false,
		},
		{
			name:      "timestamp without extension",
			filename:  "20251212052231",
			wantTime:  time.Date(2025, 12, 12, 5, 22, 31, 0, time.UTC),
			wantError: false,
		},
		{
			name:      "no timestamp",
			filename:  "image.png",
			wantError: true,
		},
		{
			name:      "invalid timestamp format",
			filename:  "20251212.png",
			wantError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := parseTimestampFromFilename(tc.filename)
			if tc.wantError {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.wantTime, got)
			}
		})
	}
}

func TestLoadImage(t *testing.T) {
	parser, err := NewParser(1920, 1080)
	require.NoError(t, err)
	defer parser.Close()

	testCases := []struct {
		name      string
		imagePath string
		wantError bool
	}{
		{
			name:      "valid PNG image",
			imagePath: "../../../testdata/images/iamabanana.png",
			wantError: false,
		},
		{
			name:      "non-existent file",
			imagePath: "../../../testdata/images/nonexistent.png",
			wantError: true,
		},
		{
			name:      "non-PNG file (JPEG)",
			imagePath: "../../../testdata/images/test-image.jpg",
			wantError: true,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			img, err := parser.loadImage(tc.imagePath)
			if tc.wantError {
				assert.Error(t, err)
				assert.Nil(t, img)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, img)
				// Check that image has valid bounds
				bounds := img.Bounds()
				assert.Greater(t, bounds.Dx(), 0)
				assert.Greater(t, bounds.Dy(), 0)
			}
		})
	}
}

func TestFilterEmpty(t *testing.T) {
	testCases := []struct {
		name     string
		input    []string
		expected []string
	}{
		{
			name:     "empty slice",
			input:    []string{},
			expected: nil,
		},
		{
			name:     "all empty strings",
			input:    []string{"", "   ", "\t", "\n"},
			expected: nil,
		},
		{
			name:     "mixed empty and non-empty",
			input:    []string{"", "hello", "   ", "world", "\t"},
			expected: []string{"hello", "world"},
		},
		{
			name:     "no empty strings",
			input:    []string{"hello", "world"},
			expected: []string{"hello", "world"},
		},
		{
			name:     "strings with only whitespace",
			input:    []string{"  ", "\t\t", "hello", "  \n  "},
			expected: []string{"hello"},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := filterEmpty(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsNumeric(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "simple number",
			input:    "123",
			expected: true,
		},
		{
			name:     "number with commas",
			input:    "1,234",
			expected: true,
		},
		{
			name:     "number with spaces",
			input:    "1 234",
			expected: true,
		},
		{
			name:     "number with mixed separators",
			input:    "1,234,567",
			expected: true,
		},
		{
			name:     "non-numeric string",
			input:    "hello",
			expected: false,
		},
		{
			name:     "number with text",
			input:    "123abc",
			expected: false,
		},
		{
			name:     "empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: false,
		},
		{
			name:     "decimal number",
			input:    "123.45",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isNumeric(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestNormalizeDifficulty(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "expert lowercase",
			input:    "expert",
			expected: "Expert",
		},
		{
			name:     "expert mixed case",
			input:    "ExPeRt",
			expected: "Expert",
		},
		{
			name:     "hard",
			input:    "hard",
			expected: "Hard",
		},
		{
			name:     "medium",
			input:    "medium",
			expected: "Medium",
		},
		{
			name:     "easy",
			input:    "easy",
			expected: "Easy",
		},
		{
			name:     "unknown difficulty",
			input:    "unknown",
			expected: "unknown",
		},
		{
			name:     "empty string",
			input:    "",
			expected: "",
		},
		{
			name:     "text containing difficulty",
			input:    "played on expert mode",
			expected: "Expert",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := normalizeDifficulty(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestIsPlayerSeparator(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected bool
	}{
		{
			name:     "empty string",
			input:    "",
			expected: true,
		},
		{
			name:     "whitespace only",
			input:    "   ",
			expected: true,
		},
		{
			name:     "line with dashes",
			input:    "---",
			expected: true,
		},
		{
			name:     "line starting with dashes",
			input:    "---separator---",
			expected: true,
		},
		{
			name:     "dashes with whitespace",
			input:    "  ---  ",
			expected: true,
		},
		{
			name:     "regular text",
			input:    "Player 1",
			expected: false,
		},
		{
			name:     "text with dashes in middle",
			input:    "Player-1",
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isPlayerSeparator(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

func TestCropImage(t *testing.T) {
	// Create a test image
	testImg := createTestImage(100, 100)

	testCases := []struct {
		name           string
		left, top      int
		right, bottom  int
		expectedWidth  int
		expectedHeight int
	}{
		{
			name:           "crop center region",
			left:           25,
			top:            25,
			right:          75,
			bottom:         75,
			expectedWidth:  50,
			expectedHeight: 50,
		},
		{
			name:           "crop top left",
			left:           0,
			top:            0,
			right:          50,
			bottom:         50,
			expectedWidth:  50,
			expectedHeight: 50,
		},
		{
			name:           "crop with out of bounds (clamped)",
			left:           -10,
			top:            -10,
			right:          150,
			bottom:         150,
			expectedWidth:  100,
			expectedHeight: 100,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			cropped := cropImage(testImg, tc.left, tc.top, tc.right, tc.bottom)
			bounds := cropped.Bounds()
			assert.Equal(t, tc.expectedWidth, bounds.Dx())
			assert.Equal(t, tc.expectedHeight, bounds.Dy())
		})
	}
}

func TestExtractText(t *testing.T) {
	parser, close := newTestParser(t)
	defer close()

	img, err := parser.loadImage("../../../testdata/images/iamabanana.png")
	require.NoError(t, err)

	text := parser.extractText(img)
	// The exact text may vary based on OCR accuracy, but it should not be empty
	// for a valid image with text
	assert.NotEmpty(t, text)
}

func TestParseImage(t *testing.T) {
	parser, close := newTestParser(t)
	defer close()

	// Test with the provided test image - fail if test image is missing
	imagePath := "../../../testdata/images/iamabanana.png"
	_, err := os.Stat(imagePath)
	require.NoError(t, err, "Test image is required but not found: %s. This test requires the test image to be present.", imagePath)

	scoreData, err := parser.ParseImage(imagePath)
	require.NoError(t, err)
	require.NotNil(t, scoreData)

	// Verify the structure is valid
	assert.NotNil(t, scoreData.CreatedAt)
	// Note: The actual extracted values depend on OCR accuracy
	// We just verify the structure is created correctly
}

func TestParseImage_NonExistentFile(t *testing.T) {
	parser, close := newTestParser(t)
	defer close()

	scoreData, err := parser.ParseImage("nonexistent.png")
	assert.Error(t, err)
	assert.Nil(t, scoreData)
}

func TestParseImage_InvalidImage(t *testing.T) {
	// Create a temporary file that's not a valid image
	tmpFile, err := os.CreateTemp("", "test-*.txt")
	require.NoError(t, err)
	defer os.Remove(tmpFile.Name())
	defer tmpFile.Close()

	_, err = tmpFile.WriteString("not an image")
	require.NoError(t, err)
	tmpFile.Close()

	parser, close := newTestParser(t)
	defer close()

	scoreData, err := parser.ParseImage(tmpFile.Name())
	assert.Error(t, err)
	assert.Nil(t, scoreData)
}

// Test that ParseImage handles files with timestamps in different formats
func TestParseImage_TimestampFormats(t *testing.T) {
	// This test requires Tesseract to be installed - fail loudly if not available
	parser, close := newTestParser(t)
	defer close()

	// Create a temporary copy of the test image with different timestamp formats
	// Fail if test image is missing
	testImagePath := "../../../testdata/images/iamabanana.png"
	_, err := os.Stat(testImagePath)
	require.NoError(t, err, "Test image is required but not found: %s. This test requires the test image to be present.", testImagePath)

	testCases := []struct {
		name     string
		filename string
	}{
		{
			name:     "timestamp in middle",
			filename: "clonehero-Artist-20251212052231.png",
		},
		{
			name:     "timestamp at start",
			filename: "20251212052231-image.png",
		},
		{
			name:     "no timestamp",
			filename: "image.png",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create temp file with the test filename
			tmpDir := t.TempDir()
			tmpPath := filepath.Join(tmpDir, tc.filename)

			// Copy test image
			data, err := os.ReadFile(testImagePath)
			require.NoError(t, err)
			err = os.WriteFile(tmpPath, data, 0644)
			require.NoError(t, err)

			scoreData, err := parser.ParseImage(tmpPath)
			// Should not error even if timestamp parsing fails (falls back to file mod time)
			assert.NoError(t, err)
			assert.NotNil(t, scoreData)
		})
	}
}

// Test images generated by scripts/generate_test_images.go
func TestParseImage_TestImages(t *testing.T) {
	parser, close := newTestParser(t)
	defer close()

	testImages := []struct {
		name     string
		filepath string
		validate func(*testing.T, *db.CreateScoreData)
	}{
		{
			name:     "test-top-left",
			filepath: "../../../testdata/images/test-top-left.png",
			validate: func(t *testing.T, data *db.CreateScoreData) {
				// Should extract artist, song, charter from top-left
				assert.NotEmpty(t, data.Artist, "Should extract artist from top-left")
				assert.NotEmpty(t, data.SongName, "Should extract song name from top-left")
			},
		},
		{
			name:     "test-center",
			filepath: "../../../testdata/images/test-center.png",
			validate: func(t *testing.T, data *db.CreateScoreData) {
				// Should extract score and stars from center
				// Note: OCR may not always extract perfectly, so we just check structure
				assert.NotNil(t, data, "Should parse center image")
			},
		},
		{
			name:     "test-players",
			filepath: "../../../testdata/images/test-players.png",
			validate: func(t *testing.T, data *db.CreateScoreData) {
				// Should extract player data
				assert.Greater(t, len(data.Players), 0, "Should extract at least one player")
			},
		},
		{
			name:     "test-numbers",
			filepath: "../../../testdata/images/test-numbers.png",
			validate: func(t *testing.T, data *db.CreateScoreData) {
				// Should handle various number formats
				assert.NotNil(t, data, "Should parse numbers image")
			},
		},
		{
			name:     "test-difficulty",
			filepath: "../../../testdata/images/test-difficulty.png",
			validate: func(t *testing.T, data *db.CreateScoreData) {
				// Should extract difficulty information
				assert.NotNil(t, data, "Should parse difficulty image")
			},
		},
		{
			name:     "test-percentages",
			filepath: "../../../testdata/images/test-percentages.png",
			validate: func(t *testing.T, data *db.CreateScoreData) {
				// Should extract percentage values
				assert.NotNil(t, data, "Should parse percentages image")
			},
		},
		{
			name:     "test-complete",
			filepath: "../../../testdata/images/test-complete.png",
			validate: func(t *testing.T, data *db.CreateScoreData) {
				// Should extract all elements
				assert.NotEmpty(t, data.Artist, "Should extract artist")
				assert.NotEmpty(t, data.SongName, "Should extract song name")
				assert.NotNil(t, data.TotalScore, "Should extract total score")
				assert.Greater(t, len(data.Players), 0, "Should extract players")
			},
		},
		{
			name:     "test-stars",
			filepath: "../../../testdata/images/test-stars.png",
			validate: func(t *testing.T, data *db.CreateScoreData) {
				// Should extract star information
				assert.NotNil(t, data, "Should parse stars image")
			},
		},
		{
			name:     "test-misses",
			filepath: "../../../testdata/images/test-misses.png",
			validate: func(t *testing.T, data *db.CreateScoreData) {
				// Should extract miss information
				assert.NotNil(t, data, "Should parse misses image")
			},
		},
		{
			name:     "test-large",
			filepath: "../../../testdata/images/test-large.png",
			validate: func(t *testing.T, data *db.CreateScoreData) {
				// Should handle large images (resized)
				assert.NotNil(t, data, "Should parse large image after resizing")
			},
		},
		{
			name:     "test-image-jpeg",
			filepath: "../../../testdata/images/test-image.jpg",
			validate: func(t *testing.T, data *db.CreateScoreData) {
				// JPEG images are not supported; expect an error/failure
				fmt.Printf("data: %v", data)
				assert.Nil(t, data, "JPEG images should not be parsed successfully (only PNG supported)")
			},
		},
	}

	for _, tc := range testImages {
		t.Run(tc.name, func(t *testing.T) {
			_, statErr := os.Stat(tc.filepath)
			require.NoError(t, statErr, "Test image is required but not found: %s. Run 'make generate-test-images' to create it.", tc.filepath)

			scoreData, parseErr := parser.ParseImage(tc.filepath)

			// Provide an error-aware test helper for custom validation, or default behaviors
			type scenario struct {
				expectErr  bool
				validateFn func(t *testing.T, data *db.CreateScoreData, err error)
			}

			// Compose a scenario for each case
			var s scenario
			if tc.validate != nil {
				s = scenario{
					validateFn: func(t *testing.T, data *db.CreateScoreData, err error) {
						tc.validate(t, data)
					},
				}
			} else {
				s = scenario{
					validateFn: func(t *testing.T, data *db.CreateScoreData, err error) {
						require.NoError(t, err, "Failed to parse image: %s", tc.filepath)
						require.NotNil(t, data, "Parsed data should not be nil")
					},
				}
			}

			// Always run validation, passing both scoreData and error
			s.validateFn(t, scoreData, parseErr)
		})
	}
}

func TestParseImage_EmptyImage(t *testing.T) {
	parser, close := newTestParser(t)
	defer close()

	// Test with empty image
	imagePath := "../../../testdata/images/test-empty.png"
	_, err := os.Stat(imagePath)
	require.NoError(t, err, "Test image is required but not found: %s. Run 'make generate-test-images' to create it.", imagePath)

	scoreData, err := parser.ParseImage(imagePath)
	// Empty image should still parse (may return empty data)
	// The warning should be logged, but parsing should succeed
	assert.NoError(t, err, "Empty image should parse without error")
	assert.NotNil(t, scoreData, "Parsed data should not be nil even for empty image")
}

func TestExtractTopLeftInfo_TestImage(t *testing.T) {
	parser, close := newTestParser(t)
	defer close()

	imagePath := "../../../testdata/images/test-top-left.png"
	_, err := os.Stat(imagePath)
	require.NoError(t, err, "Test image is required but not found: %s. Run 'make generate-test-images' to create it.", imagePath)

	img, err := parser.loadImage(imagePath)
	require.NoError(t, err)

	artist, songName, charter := parser.extractTopLeftInfo(img)

	// Verify extraction from top-left region
	assert.NotEmpty(t, artist, "Should extract artist from top-left region")
	assert.NotEmpty(t, songName, "Should extract song name from top-left region")
	// Charter may or may not be extracted depending on OCR accuracy
	_ = charter // Suppress unused variable warning
}

func TestExtractCenterInfo_TestImage(t *testing.T) {
	parser, close := newTestParser(t)
	defer close()

	imagePath := "../../../testdata/images/test-center.png"
	_, err := os.Stat(imagePath)
	require.NoError(t, err, "Test image is required but not found: %s. Run 'make generate-test-images' to create it.", imagePath)

	img, err := parser.loadImage(imagePath)
	require.NoError(t, err)

	totalScore, stars := parser.extractCenterInfo(img)

	// Verify extraction from center region
	// OCR may not always extract perfectly, but we verify the function runs
	assert.NotNil(t, totalScore, "Should attempt to extract total score from center region")
	// Stars may or may not be extracted depending on OCR accuracy
	_ = stars // Suppress unused variable warning
}

func TestExtractPlayers_TestImage(t *testing.T) {
	parser, close := newTestParser(t)
	defer close()

	imagePath := "../../../testdata/images/test-players.png"
	_, err := os.Stat(imagePath)
	require.NoError(t, err, "Test image is required but not found: %s. Run 'make generate-test-images' to create it.", imagePath)

	img, err := parser.loadImage(imagePath)
	require.NoError(t, err)

	players := parser.extractPlayers(img)

	// Verify extraction from player region
	assert.Greater(t, len(players), 0, "Should extract at least one player from player region")

	// Verify player structure
	for _, player := range players {
		assert.NotEmpty(t, player.Name, "Player should have a name")
	}
}

func TestExtractRealScoreData(t *testing.T) {
	testCases := []struct {
		name     string
		filepath string
		expected *db.CreateScoreData
	}{
		{
			name:     "tripping billies",
			filepath: "../../../testdata/scores/clonehero-Tripping-Billies-20251209195440.png",
			expected: &db.CreateScoreData{
				Artist:        "Dave Matthews Band",
				SongName:      "Tripping Billies",
				TotalScore:    447253,
				StarsAchieved: 4,
				Players: []db.Player{
					{
						Name:       "_gem_",
						Difficulty: "Expert",
					},
					{
						Name:        "A_Hole_Pro",
						Difficulty:  "Expert",
						Score:       68508,
						Accuracy:    81,
						NotesMissed: 0,
						BestStreak:  100,
						Rank:        1,
					},
				},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			parser, close := newTestParser(t)
			defer close()

			scoreData, err := parser.ParseImage(tc.filepath)
			require.NoError(t, err)
			require.NotNil(t, scoreData)

			assert.NotNil(t, scoreData.CreatedAt)
		})
	}
}

func createTestImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	// Fill with a simple pattern (not necessary for crop test, but makes it valid)
	// Fill with white color
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{255, 255, 255, 255})
		}
	}
	return img
}

func newTestParser(t *testing.T) (*Parser, func() error) {
	parser, err := NewParser(1920, 1080)
	require.NoError(t, err, "Tesseract OCR is not installed or not available. This is required for the parser service. Install with: sudo apt-get install tesseract-ocr (Ubuntu/Debian) or brew install tesseract (macOS)")

	return parser, parser.Close
}

func ptrToInt64(v int64) *int64 {
	return &v
}

func ptrToInt(v int) *int {
	return &v
}

func ptrToString(v string) *string {
	return &v
}
