package convert

import (
	"bytes"
	"fmt"
	"image"
	_ "image/jpeg"
	_ "image/png"
	"log/slog"
	"os"

	"github.com/gen2brain/webp"
)

// fileExists returns true if the given file or directory exists.
func fileExists(filename string) bool {
	_, err := os.Stat(filename)
	if os.IsNotExist(err) {
		return false
	}
	return true
}

// imageFromBytes converts the bytes data to an Image.
func imageFromBytes(data []byte) (image.Image, error) {
	reader := bytes.NewReader(data)
	image, _, err := image.Decode(reader)
	return image, err
}

// SaveWebP converts the given image data to WebP and saves at dest. Recommend 90 quality.
func SaveWebP(data []byte, dest string, quality int) error {
	if fileExists(dest) {
		return fmt.Errorf("Error: %s already exists", dest)
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
