package ghost

import (
	"bufio"
	"bytes"
	"image"
	"image/color"
	"image/png"
	"math"
	"math/rand"
	"os"

	"github.com/auyer/steganography"
)

// GetRequiredDimensions calculates the side length of a square image
// needed to fit the provided data using 3 bits per pixel (LSB in RGB).
func getRequiredDimensions(dataLen int) int {
	// Total bits needed
	bitsNeeded := float64(dataLen * 8)
	// Each pixel provides 3 bits (R, G, B)
	pixelsNeeded := math.Ceil(bitsNeeded / 3.0)
	// Calculate side of a square (sqrt) and add a small buffer for metadata
	side := int(math.Ceil(math.Sqrt(pixelsNeeded))) + 5
	return side
}

// EncodeToImage creates a random noise image based on data size,
// embeds data, and saves to disk.
func EncodeToImage(data []byte, path string) error {
	side := getRequiredDimensions(len(data))

	// Generate random carrier image
	img := image.NewNRGBA(image.Rect(0, 0, side, side))
	for y := 0; y < side; y++ {
		for x := 0; x < side; x++ {
			img.Set(x, y, color.NRGBA{
				R: uint8(rand.Intn(256)),
				G: uint8(rand.Intn(256)),
				B: uint8(rand.Intn(256)),
				A: 255,
			})
		}
	}

	// Prepare the output file
	outFile, err := os.Create(path)
	if err != nil {
		return err
	}
	defer outFile.Close()

	// Embed and write to file
	// auyer/steganography uses the LSB of the RGB channels
	var buf bytes.Buffer
	if err := steganography.Encode(&buf, img, data); err != nil {
		return err
	}
	if _, err := buf.WriteTo(outFile); err != nil {
		return err
	}
	return nil
}

// DecodeFromImage extracts the hidden data from an image file.
func DecodeFromImage(path string) ([]byte, error) {
	inFile, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer inFile.Close()

	img, err := png.Decode(bufio.NewReader(inFile))
	if err != nil {
		return nil, err
	}

	// The library stores the size in the first few pixels of the image
	size := steganography.GetMessageSizeFromImage(img)
	return steganography.Decode(size, img), nil
}
