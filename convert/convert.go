package convert

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/gen2brain/webp"

	"faceclaimer/checks"
)

// imageFromBytes converts the bytes data to an Image.
func imageFromBytes(data []byte) (image.Image, error) {
	reader := bytes.NewReader(data)
	image, format, err := image.Decode(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to decode image: %w", err)
	}
	slog.Info("Read image data", "format", format)
	return image, nil
}

// SaveWebP converts image data to WebP format and saves it to dest with the specified quality (recommended: 90).
func SaveWebP(data []byte, dest string, quality int) (err error) {
	if checks.PathExists(dest) {
		return fmt.Errorf("%s already exists", dest)
	}

	// Create the directory structure if it doesn't exist
	dir := filepath.Dir(dest)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("unable to create directory %s: %w", dir, err)
	}

	outputFile, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("unable to create %s: %w", dest, err)
	}
	defer func() {
		if closeErr := outputFile.Close(); closeErr != nil && err == nil {
			err = fmt.Errorf("failed to close %s: %w", dest, closeErr)
		}
	}()

	image, err := imageFromBytes(data)
	if err != nil {
		return err
	}

	slog.Info("Converting image to WebP")
	options := webp.Options{
		Quality:  quality,
		Lossless: false,
		Method:   0, // Size difference is minimal; performance difference is massive
		Exact:    false,
	}

	err = webp.Encode(outputFile, image, options)
	if err != nil {
		return err
	}

	slog.Info("Saved WebP image", "dest", dest)
	return nil
}
