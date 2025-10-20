/*
Copyright © 2025 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/spf13/cobra"

	"image-processor/checks"
	"image-processor/routes"
)

var (
	port      int
	imagesDir string
	baseURL   string
)

// rootCmd represents the base command when called without any subcommands
var rootCmd = &cobra.Command{
	Use:   "image-processor",
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
		return nil
	},
	Run: func(cmd *cobra.Command, args []string) {
		slog.Info("Starting images-processor", "imagesDir", imagesDir, "baseURL", baseURL)
		routes.Run(baseURL, imagesDir, port)
	},
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
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
	rootCmd.MarkFlagRequired("base-url")
}
