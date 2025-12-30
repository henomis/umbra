package ghost

import (
	"encoding/base64"
	"fmt"
	"image"
	"os"

	"github.com/makiuchi-d/gozxing"
	"github.com/makiuchi-d/gozxing/qrcode"
	goqrcode "github.com/skip2/go-qrcode"
)

const maxQRBufferSize = 2953

// EncodeToQR generates a QR code image containing the binary data.
func EncodeToQR(data []byte, path string) error {
	if len(data) > maxQRBufferSize {
		return fmt.Errorf("data size %d exceeds max QR code capacity of %d bytes", len(data), maxQRBufferSize)
	}

	// Encode binary data as base64
	encoded := base64.StdEncoding.EncodeToString(data)

	qr, err := goqrcode.New(encoded, goqrcode.Highest)
	if err != nil {
		return fmt.Errorf("failed to create QR: %w", err)
	}

	// Calculate a reasonable pixel size.
	// For high-density QR codes, we want at least 512x512 for scannability.
	pixelSize := 512
	if len(data) > 1000 {
		pixelSize = 1024
	}

	return qr.WriteFile(pixelSize, path)
}

// DecodeFromQR reads a QR code image file and extracts the embedded data.
func DecodeFromQR(path string) ([]byte, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	img, _, err := image.Decode(file)
	if err != nil {
		return nil, err
	}

	// prepare BinaryBitmap
	bmp, err := gozxing.NewBinaryBitmapFromImage(img)
	if err != nil {
		return nil, err
	}
	// decode image
	qrReader := qrcode.NewQRCodeReader()
	result, err := qrReader.Decode(bmp, nil)
	if err != nil {
		return nil, err
	}

	// Decode base64 to get original binary data
	decoded, err := base64.StdEncoding.DecodeString(result.GetText())
	if err != nil {
		return nil, fmt.Errorf("failed to decode base64: %w", err)
	}

	return decoded, nil
}
