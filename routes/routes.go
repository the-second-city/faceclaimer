package routes

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"faceclaimer/checks"
	"faceclaimer/convert"
)

type Config struct {
	ImagesDir string
	BaseURL   string
	Quality   int
}

type UploadRequest struct {
	Guild    int    `json:"guild"`
	User     int    `json:"user"`
	CharID   string `json:"charid"`
	ImageURL string `json:"image_url"`
}

// setupRouter sets up gin's route handlers.
func setupRouter(cfg *Config) *gin.Engine {
	r := gin.Default()
	r.SetTrustedProxies(nil)
	r.POST("/image/upload", func(c *gin.Context) {
		handleImageUpload(c, cfg)
	})
	r.DELETE("/image/*imagePath", func(c *gin.Context) {
		handleSingleDelete(c, cfg)
	})
	r.DELETE("/character/:charID", func(c *gin.Context) {
		handleCharacterDelete(c, cfg)
	})

	return r
}

func handleImageUpload(c *gin.Context, cfg *Config) {
	var request UploadRequest
	if err := c.BindJSON(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	slog.Info("Image upload request", "user", request.User, "guild", request.Guild, "charId", request.CharID)

	if !checks.IsValidURL(request.ImageURL) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid image URL"})
		return
	}

	// Download the image from the remote w/ timeout and size limit
	client := &http.Client{
		Timeout: 30 * time.Second,
	}
	resp, err := client.Get(request.ImageURL)
	if err != nil || resp.StatusCode != http.StatusOK {
		c.AbortWithStatusJSON(http.StatusBadGateway, gin.H{"error": "failed to download image"})
		return
	}
	defer resp.Body.Close()

	// Discord Nitro users can upload up to 500MB, but anyone doing that with
	// images is clearly insane, and we're being generous with a 100MB limit.
	limitedReader := io.LimitReader(resp.Body, 100*1024*1024)
	imageData, err := io.ReadAll(limitedReader)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	slog.Info("Downloaded image data", "url", request.ImageURL)

	// We need the image name and the save location separately so we can construct
	// the URL to return to the user.
	imageNameParts, err := prepImageNameParts(request)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	saveLoc := filepath.Join(append([]string{cfg.ImagesDir}, imageNameParts...)...)
	err = convert.SaveWebP(imageData, saveLoc, cfg.Quality)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// The web URL doesn't include the images directory. That way, we can place
	// the images at root, e.g. https://example.com/guildId/userId/charId/imageId.webp
	imageURL := strings.Join(imageNameParts, "/")
	webRoot := strings.Trim(cfg.BaseURL, "/")
	imageURL = strings.Join([]string{webRoot, imageURL}, "/")

	c.JSON(http.StatusCreated, imageURL)
}

// cleanEmptyDirs recursively deletes empty directories within baseDir.
// It does NOT delete baseDir itself, only empty subdirectories.
func cleanEmptyDirs(baseDir string) error {
	entries, err := os.ReadDir(baseDir)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}

		fullPath := filepath.Join(baseDir, entry.Name())

		// Recursively clean subdirectories first
		if err := cleanEmptyDirs(fullPath); err != nil {
			return err
		}

		// After cleaning subdirectories, check if this directory is now empty
		subEntries, err := os.ReadDir(fullPath)
		if err != nil {
			return err
		}

		if len(subEntries) == 0 {
			if err := os.Remove(fullPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// handleSingleDelete performs a single image deletion.
func handleSingleDelete(c *gin.Context, cfg *Config) {
	// We keep imagePath separate from loc, for the return value
	// Wildcard params include a leading slash, so strip it
	imagePath := strings.TrimPrefix(c.Param("imagePath"), "/")
	imageLoc, err := checks.AbsPath(cfg.ImagesDir, imagePath)
	fmt.Println(imageLoc)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !checks.PathExists(imageLoc) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Image not found"})
		return
	}
	if checks.DirExists(imageLoc) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Cannot delete directory"})
		return
	}
	if err := os.Remove(imageLoc); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := cleanEmptyDirs(cfg.ImagesDir); err != nil {
		slog.Warn("Failed to clean empty directories", "error", err)
	}

	c.JSON(http.StatusOK, fmt.Sprintf("Deleted %s", imagePath))
}

// handleCharacterDelete deletes all of a character's images.
func handleCharacterDelete(c *gin.Context, cfg *Config) {
	charID := c.Param("charID")

	if !checks.IsValidObjectId(charID) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Invalid character ID"})
		return
	}

	charPath, err := checks.AbsPath(cfg.ImagesDir, charID)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !checks.PathExists(charPath) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Character directory not found"})
		return
	}

	slog.Info("Deleting", "path", charPath)
	if err = os.RemoveAll(charPath); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	if err := cleanEmptyDirs(cfg.ImagesDir); err != nil {
		slog.Warn("Failed to clean empty directories", "error", err)
	}

	c.JSON(http.StatusOK, fmt.Sprintf("Deleted all images: %s", charID))
}

// prepImageNameParts generates path components for a new image upload.
// It returns a slice containing [guild, user, charID, imageID.webp] that can be
// used with filepath.Join for OS-specific file paths or strings.Join for URLs.
// The function validates that CharID is a valid MongoDB ObjectID and generates
// a unique ObjectID for the image filename.
func prepImageNameParts(r UploadRequest) ([]string, error) {
	if !checks.IsValidObjectId(r.CharID) {
		return nil, fmt.Errorf("%s is not a valid character ID", r.CharID)
	}
	charId := fmt.Sprint(r.CharID)
	imageName := fmt.Sprintf("%s.webp", primitive.NewObjectID().Hex())
	return []string{charId, imageName}, nil
}

// Run starts the HTTP server with the given configuration.
func Run(baseURL, imagesDir string, port, quality int) {
	cfg := &Config{
		ImagesDir: imagesDir,
		BaseURL:   baseURL,
		Quality:   quality,
	}

	r := setupRouter(cfg)
	r.Run(fmt.Sprintf(":%d", port))
}
