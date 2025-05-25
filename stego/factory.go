package stego

import (
	"fmt"
)

// Factory creates and returns the appropriate steganography algorithm
func Factory(algorithmName string) (Steganographer, error) {
	switch algorithmName {
	case "Fractal":
		return NewFractalStego(), nil
	default:
		return nil, fmt.Errorf("unknown steganography algorithm: %s", algorithmName)
	}
}
