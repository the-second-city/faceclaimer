package routes

import (
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"image-processor/checks"
	"image-processor/convert"
)

type Config struct {
	ImagesDir string
	BaseURL   string
}

type UploadRequest struct {
	Guild    int    `json:"guild"`
	User     int    `json:"user"`
	CharID   string `json:"charid"`
	ImageURL string `json:"image_url"`
}

func setupRouter(cfg *Config) *gin.Engine {
	r := gin.Default()
	r.SetTrustedProxies(nil)
	r.POST("/image/upload", func(c *gin.Context) {
		handleImageUpload(c, cfg)
	})
	r.DELETE("/image/delete/:imagePath", func(c *gin.Context) {
		handleSingleDelete(c, cfg)
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
	imageName, err := prepImageName(request)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	saveLoc := strings.Join([]string{cfg.ImagesDir, imageName}, "/")
	err = convert.SaveWebP(imageData, saveLoc, 90)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// The web URL doesn't include the images directory. That way, we can place
	// the images at root, e.g. https://example.com/guildId/userId/charId/imageId.webp
	webRoot := strings.Trim(cfg.BaseURL, "/")
	url := strings.Join([]string{webRoot, imageName}, "/")

	c.JSON(http.StatusCreated, url)
}

// handleSingleDelete performs a single image deletion.
func handleSingleDelete(c *gin.Context, cfg *Config) {
	// We keep imagePath separate from loc, for the return value
	imagePath := c.Param("imagePath")
	imageLoc, err := checks.AbsPath(cfg.ImagesDir, imagePath)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if !checks.PathExists(imageLoc) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "Image not found"})
		return
	}
	if err := os.Remove(imageLoc); err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, fmt.Sprintf("Deleted %s", imagePath))
}

func prepImageName(r UploadRequest) (string, error) {
	if !checks.IsValidObjectId(r.CharID) {
		return "", fmt.Errorf("%s is not a valid character ID", r.CharID)
	}
	guild := fmt.Sprint(r.Guild)
	user := fmt.Sprint(r.User)
	charId := fmt.Sprint(r.CharID)
	imageName := fmt.Sprintf("%s.webp", primitive.NewObjectID().Hex())
	return strings.Join([]string{guild, user, charId, imageName}, "/"), nil
}

func Run(baseURL, imagesDir string, port int) {
	cfg := &Config{
		ImagesDir: imagesDir,
		BaseURL:   baseURL,
	}

	r := setupRouter(cfg)
	r.Run(fmt.Sprintf(":%d", port))
}
