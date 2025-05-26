package stego

import (
	"image"
)

// Steganographer is the interface that all steganography algorithms must implement
type Steganographer interface {
	// Embed embeds the provided data into the cover image
	Embed(cover image.Image, data []byte, config Config) (image.Image, error)
	// Extract extracts the hidden data from the stego image
	Extract(stego image.Image, config Config) ([]byte, error)
	// Name returns the name of the algorithm
	Name() string
}

// Config holds configuration data needed for steganography algorithms
type Config struct {
	// EmbeddingRate is the proportion of available cover elements to use (0.0-1.0)
	EmbeddingRate float64
	// FractalParams contains parameters for fractal-based steganography
	FractalParams *FractalParams
}

// FractalParams contains configuration for fractal-based steganography
type FractalParams struct {
	// Type is the fractal type (e.g., "Mandelbrot", "Julia")
	Type string
	// Iterations is the maximum number of iterations for the fractal calculation
	Iterations int
	// Threshold is the escape radius for the fractal calculation
	Threshold float64
}
