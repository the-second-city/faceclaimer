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

	slog.Info("Read image data", "format", format)

	return image, err
}

// SaveWebP converts the given image data to WebP and saves at dest. Recommend 90 quality.
func SaveWebP(data []byte, dest string, quality int) error {
	if checks.PathExists(dest) {
		return fmt.Errorf("Error: %s already exists", dest)
	}

	// Create the directory structure if it doesn't exist
	dir := filepath.Dir(dest)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return fmt.Errorf("Error: unable to create directory %s: %v", dir, err)
	}

	outputFile, err := os.Create(dest)
	if err != nil {
		return fmt.Errorf("Error: unable to create %s: %v", dest, err)
	}
	defer outputFile.Close()

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
