package stego

import (
	"encoding/binary"
	"errors"
	"image"
	"image/color"
	"image/draw"
	"math/cmplx"
)

type FractalStego struct{}

func NewFractalStego() *FractalStego {
	return &FractalStego{}
}

func (f *FractalStego) Name() string {
	return "Fractal"
}

func (f *FractalStego) Embed(cover image.Image, data []byte, config Config) (image.Image, error) {
	if config.FractalParams == nil {
		return nil, errors.New("fractal parameters are required")
	}

	bounds := cover.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	// Calculate required pixels (32 bits for length + data bits)
	requiredPixels := 32 + len(data)*8
	if width*height < requiredPixels {
		return nil, errors.New("image too small to embed data")
	}

	stego := image.NewRGBA(bounds)
	draw.Draw(stego, bounds, cover, bounds.Min, draw.Src)

	// Generate fractal pattern
	pattern := f.generateFractalPattern(width, height, config.FractalParams)

	// Prepare all bits: [32-bit length][data bits]
	lengthBits := make([]byte, 4)
	binary.BigEndian.PutUint32(lengthBits, uint32(len(data)))
	allBits := append(bytesToBits(lengthBits), bytesToBits(data)...)

	// Embed bits in fractal pattern pixels
	bitPos := 0
	for y := 0; y < height && bitPos < len(allBits); y++ {
		for x := 0; x < width && bitPos < len(allBits); x++ {
			if pattern[y*width+x] {
				c := stego.RGBAAt(x, y)
				c.B = (c.B & 0xFE) | allBits[bitPos]
				stego.SetRGBA(x, y, c)
				bitPos++
			}
		}
	}

	return stego, nil
}

func (f *FractalStego) Extract(stego image.Image, config Config) ([]byte, error) {
	if config.FractalParams == nil {
		return nil, errors.New("fractal parameters are required")
	}

	bounds := stego.Bounds()
	width, height := bounds.Dx(), bounds.Dy()

	// Generate the same fractal pattern used in embedding
	pattern := f.generateFractalPattern(width, height, config.FractalParams)

	// First extract 32-bit length (4 bytes)
	var length uint32
	extractedBits := 0
	lengthBytes := make([]byte, 4)

	for y := 0; y < height && extractedBits < 32; y++ {
		for x := 0; x < width && extractedBits < 32; x++ {
			if pattern[y*width+x] {
				c := color.RGBAModel.Convert(stego.At(x, y)).(color.RGBA)
				bytePos := extractedBits / 8
				bitPos := 7 - (extractedBits % 8)
				lengthBytes[bytePos] |= (c.B & 1) << bitPos
				extractedBits++
			}
		}
	}

	length = binary.BigEndian.Uint32(lengthBytes)
	if length == 0 || length > uint32(width*height) {
		return nil, errors.New("invalid data length extracted")
	}

	// Then extract the actual data (without length prefix)
	dataBits := make([]byte, length*8)
	extractedBits = 0

	skippedTimes := 0
	for y := 0; y < height && extractedBits < len(dataBits); y++ {
		for x := 0; x < width && extractedBits < len(dataBits); x++ {
			pixelIndex := y*width + x
			if skippedTimes < 32 && pattern[pixelIndex] { // Skip first 32 pattern pixels
				skippedTimes++
				continue
			}
			if pattern[pixelIndex] { // Skip first 32 pattern pixels
				c := color.RGBAModel.Convert(stego.At(x, y)).(color.RGBA)
				dataBits[extractedBits] = c.B & 1
				extractedBits++
			}
		}
	}

	// Convert bits to bytes and return only the actual data (without length prefix)
	return bitsToBytes(dataBits), nil
}

func (f *FractalStego) generateFractalPattern(width, height int, params *FractalParams) []bool {
	pattern := make([]bool, width*height)

	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			// Map pixel coordinates to complex plane
			nx := float64(x)/float64(width)*3.5 - 2.5
			ny := float64(y)/float64(height)*2.0 - 1.0
			c := complex(nx, ny)
			var z complex128

			// Perform iterations
			iter := 0
			for ; iter < params.Iterations; iter++ {
				if params.Type == "Julia" {
					z = z*z + complex(-0.8, 0.156) // Fixed Julia set constant
				} else { // Mandelbrot
					z = z*z + c
				}
				if cmplx.Abs(z) > params.Threshold {
					break
				}
			}

			// Mark as embeddable if stayed within threshold
			pattern[y*width+x] = iter == params.Iterations
		}
	}

	return pattern
}

func bytesToBits(data []byte) []byte {
	bits := make([]byte, len(data)*8)
	for i, b := range data {
		for j := 0; j < 8; j++ {
			bits[i*8+j] = (b >> (7 - j)) & 1
		}
	}
	return bits
}

func bitsToBytes(bits []byte) []byte {
	bytes := make([]byte, (len(bits)+7)/8)
	for i, bit := range bits {
		if bit == 1 {
			bytes[i/8] |= 1 << (7 - (i % 8))
		}
	}
	return bytes
}
