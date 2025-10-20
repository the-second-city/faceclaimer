package main

import (
	"fmt"
	"io"
	"log"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"strings"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"

	"image-processor/convert"
)

const (
	ImagesDir = "./images"
	Port      = 8080
)

type UploadRequest struct {
	Guild    int    `json:"guild"`
	User     int    `json:"user"`
	CharID   string `json:"charid"`
	ImageURL string `json:"image_url"`
}

func imagesDirExists() bool {
	info, err := os.Stat(ImagesDir)
	if err != nil {
		return false
	}
	return info.IsDir()
}

func setupRouter() *gin.Engine {
	r := gin.Default()
	r.SetTrustedProxies(nil)
	r.POST("/image/upload", handleImageUpload)

	return r
}

func handleImageUpload(c *gin.Context) {
	var request UploadRequest
	if err := c.BindJSON(&request); err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	slog.Info("Image upload request", "user", request.User, "guild", request.Guild, "charId", request.CharID)

	// Validate URL
	if !isValidURL(request.ImageURL) {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "invalid image URL"})
		return
	}

	// Download image data
	resp, err := http.Get(request.ImageURL)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()

	imageData, err := io.ReadAll(resp.Body)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	slog.Info("Downloaded image data", "url", request.ImageURL)

	imageName, err := prepImageName(request)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	saveLoc := strings.Join([]string{ImagesDir, imageName}, "/")
	err = convert.SaveWebP(imageData, saveLoc, 90)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, imageName)
}

func isValidURL(urlStr string) bool {
	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}
	// Ensure the URL has a scheme (http/https) and a host
	return u.Scheme != "" && u.Host != ""
}

func isValidObjectId(oid string) bool {
	_, err := primitive.ObjectIDFromHex(oid)
	return err == nil
}

func prepImageName(r UploadRequest) (string, error) {
	if !isValidObjectId(r.CharID) {
		return "", fmt.Errorf("%s is not a valid character ID", r.CharID)
	}
	guild := fmt.Sprint(r.Guild)
	user := fmt.Sprint(r.User)
	charId := fmt.Sprint(r.CharID)
	imageName := fmt.Sprintf("%s.webp", primitive.NewObjectID().Hex())
	return strings.Join([]string{guild, user, charId, imageName}, "/"), nil
}

func main() {
	if !imagesDirExists() {
		log.Fatalf("%s does not exist. Please create and mount the write-only nullfs.", ImagesDir)
	}

	r := setupRouter()
	r.Run(fmt.Sprintf(":%d", Port))
}
