/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"faceclaimer/checks"
	"faceclaimer/routes"
)

var (
	port      int
	imagesDir string
	baseURL   string
	quality   int
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "faceclaimer",
	Short: "API for managing character profile images.",
	Long: `A web API for uploading and deleting character images uploaded to Discord.

Images are converted to WebP before being saved to local storage, inside a
sub-directory of guildId/userId/charId/imageId.webp, with imageId being a
BSON ObjectId.

*THIS API TAKES NO AUTHENTICATION!* It is recommended to run it in a jail
without an internet connection.`,
	PreRunE: func(cmd *cobra.Command, args []string) error {
		if !checks.IsValidURL(baseURL) {
			return errors.New("base-url must be a valid URL (e.g., https://example.com)")
		}
		if !checks.DirExists(imagesDir) {
			return fmt.Errorf("images-dir does not exist: %s", imagesDir)
		}
		if quality < 1 || quality > 100 {
			return errors.New("quality must be between 1 and 100")
		}

		// Convert imagesDir to absolute path for consistency and reliability
		absPath, err := filepath.Abs(imagesDir)
		if err != nil {
			return fmt.Errorf("failed to resolve absolute path for images-dir: %w", err)
		}
		imagesDir = absPath

		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		slog.Info("Starting images-processor", "imagesDir", imagesDir, "baseURL", baseURL, "quality", quality)
		routes.Run(baseURL, imagesDir, port, quality)
	},
}

// Execute runs the root command. Called by main.main().
func Execute() {
	err := rootCmd.Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	// Define command-line flags
	rootCmd.Flags().IntVar(&port, "port", 8080, "Port to run the server on")
	rootCmd.Flags().StringVar(&imagesDir, "images-dir", "images", "Directory to store images")
	rootCmd.Flags().StringVar(&baseURL, "base-url", "", "Base URL for constructing image URLs (e.g., https://example.com)")
	rootCmd.Flags().IntVar(&quality, "quality", 90, "WebP quality (1-100)")
	rootCmd.MarkFlagRequired("base-url")
}
