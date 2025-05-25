package stego

import (
	"github.com/stretchr/testify/assert"
	"image"
	"image/color"
	"math/rand"
	"testing"
)

// createTestImage creates a test image with random pixel values
func createTestImage(width, height int) image.Image {
	img := image.NewRGBA(image.Rect(0, 0, width, height))

	// Fill with random pixel values
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{
				R: uint8(rand.Intn(256)),
				G: uint8(rand.Intn(256)),
				B: uint8(rand.Intn(256)),
				A: 255,
			})
		}
	}

	return img
}

// TestFractalEmbedExtract tests the embedding and extraction of data
func TestFractalEmbedExtract(t *testing.T) {
	t.Parallel()

	// Create a new fractal steganographer
	stego := NewFractalStego()

	// Create test parameters
	testCases := []struct {
		name       string
		data       []byte
		fracType   string
		iterations int
		threshold  float64
		imgWidth   int
		imgHeight  int
	}{
		{
			name:       "Small data in Mandelbrot",
			data:       []byte("Hello, World!"),
			fracType:   "Mandelbrot",
			iterations: 100,
			threshold:  2.0,
			imgWidth:   200,
			imgHeight:  200,
		},
		{
			name:       "Medium data in Julia",
			data:       []byte("The quick brown fox jumps over the lazy dog. This is a longer test string to embed."),
			fracType:   "Julia",
			iterations: 50,
			threshold:  2.0,
			imgWidth:   300,
			imgHeight:  300,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a test image
			cover := createTestImage(tc.imgWidth, tc.imgHeight)

			// Create config
			config := Config{
				EmbeddingRate: 0.5,
				FractalParams: &FractalParams{
					Type:       tc.fracType,
					Iterations: tc.iterations,
					Threshold:  tc.threshold,
				},
			}

			// Embed the data
			stegoImg, err := stego.Embed(cover, tc.data, config)
			if err != nil {
				t.Fatalf("Failed to embed data: %v", err)
			}

			// Extract the data
			extractedData, err := stego.Extract(stegoImg, config)
			if err != nil {
				t.Fatalf("Failed to extract data: %v", err)
			}

			if len(extractedData) != len(tc.data) {
				t.Errorf("Extracted data length doesn't match original data length.\nOriginal: %d\nExtracted: %d",
					len(tc.data), len(extractedData))
			}
			assert.Equal(t, tc.data, extractedData)
		})
	}
}

// TestFractalPatternConsistency tests that the fractal pattern generation is consistent
func TestFractalPatternConsistency(t *testing.T) {
	t.Parallel()

	stego := NewFractalStego()

	// Test cases for different fractal parameters
	testCases := []struct {
		name       string
		fracType   string
		iterations int
		threshold  float64
		width      int
		height     int
	}{
		{
			name:       "Mandelbrot 100 iterations",
			fracType:   "Mandelbrot",
			iterations: 100,
			threshold:  2.0,
			width:      100,
			height:     100,
		},
		{
			name:       "Julia 50 iterations",
			fracType:   "Julia",
			iterations: 50,
			threshold:  2.0,
			width:      100,
			height:     100,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Generate the pattern twice with the same parameters
			pattern1 := stego.generateFractalPattern(tc.width, tc.height, &FractalParams{Type: tc.fracType, Iterations: tc.iterations, Threshold: tc.threshold})
			pattern2 := stego.generateFractalPattern(tc.width, tc.height, &FractalParams{Type: tc.fracType, Iterations: tc.iterations, Threshold: tc.threshold})

			// The patterns should be identical
			for i := 0; i < len(pattern1); i++ {
				if pattern1[i] != pattern2[i] {
					t.Errorf("Pattern generation is not consistent at index %d", i)
					break
				}
			}
		})
	}
}

// TestFractalCapacity tests that the steganography algorithm correctly handles capacity limits
func TestFractalCapacity(t *testing.T) {
	stego := NewFractalStego()

	// Create a small test image
	imgWidth, imgHeight := 50, 50
	cover := createTestImage(imgWidth, imgHeight)

	// Create data that's too large for the image
	// The fractal pattern uses approximately half the pixels, so this is definitely too large
	largeData := make([]byte, imgWidth*imgHeight)
	for i := range largeData {
		largeData[i] = byte(rand.Intn(256))
	}

	config := Config{
		EmbeddingRate: 0.5,
		FractalParams: &FractalParams{
			Type:       "Mandelbrot",
			Iterations: 100,
			Threshold:  2.0,
		},
	}

	// Try to embed the data
	_, err := stego.Embed(cover, largeData, config)

	// We expect an error because the data is too large
	if err == nil {
		t.Error("Expected an error when embedding data that exceeds capacity, but got none")
	}
}

// TestFractalParamsRequired tests that the algorithm requires fractal parameters
func TestFractalParamsRequired(t *testing.T) {
	stego := NewFractalStego()

	// Create a test image
	cover := createTestImage(100, 100)

	// Create config without fractal parameters
	config := Config{
		EmbeddingRate: 0.5,
		FractalParams: nil,
	}

	// Try to embed data
	_, err := stego.Embed(cover, []byte("test"), config)

	// We expect an error because fractal parameters are required
	if err == nil {
		t.Error("Expected an error when fractal parameters are missing, but got none")
	}

	// Create a stego image
	validConfig := Config{
		EmbeddingRate: 0.5,
		FractalParams: &FractalParams{
			Type:       "Mandelbrot",
			Iterations: 100,
			Threshold:  2.0,
		},
	}
	stegoImg, _ := stego.Embed(cover, []byte("test"), validConfig)

	// Try to extract without fractal parameters
	_, err = stego.Extract(stegoImg, config)

	// We expect an error because fractal parameters are required
	if err == nil {
		t.Error("Expected an error when fractal parameters are missing for extraction, but got none")
	}
}
