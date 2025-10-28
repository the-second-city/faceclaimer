package routes

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	"faceclaimer/checks"
)

func init() {
	// Suppress gin debug output during tests
	gin.SetMode(gin.TestMode)
}

func TestCleanEmptyDirs(t *testing.T) {
	// Create a temporary base directory
	baseDir, err := os.MkdirTemp("", "test-clean-empty-*")
	if err != nil {
		t.Fatalf("Failed to create temp base dir: %v", err)
	}
	defer os.RemoveAll(baseDir)

	// Test case 1: Simple empty directory
	t.Run("single empty directory", func(t *testing.T) {
		emptyDir := filepath.Join(baseDir, "empty1")
		if err := os.Mkdir(emptyDir, 0755); err != nil {
			t.Fatalf("Failed to create empty dir: %v", err)
		}

		if err := cleanEmptyDirs(baseDir); err != nil {
			t.Errorf("cleanEmptyDirs failed: %v", err)
		}

		// Verify the empty directory was deleted
		if _, err := os.Stat(emptyDir); !os.IsNotExist(err) {
			t.Errorf("Empty directory was not deleted: %s", emptyDir)
		}

		// Verify baseDir still exists
		if _, err := os.Stat(baseDir); os.IsNotExist(err) {
			t.Errorf("Base directory was deleted but should not be")
		}
	})

	// Test case 2: Nested empty directories
	t.Run("nested empty directories", func(t *testing.T) {
		nestedPath := filepath.Join(baseDir, "level1", "level2", "level3")
		if err := os.MkdirAll(nestedPath, 0755); err != nil {
			t.Fatalf("Failed to create nested dirs: %v", err)
		}

		if err := cleanEmptyDirs(baseDir); err != nil {
			t.Errorf("cleanEmptyDirs failed: %v", err)
		}

		// Verify all nested empty directories were deleted
		if _, err := os.Stat(filepath.Join(baseDir, "level1")); !os.IsNotExist(err) {
			t.Errorf("Nested empty directories were not deleted")
		}
	})

	// Test case 3: Directory with file should not be deleted
	t.Run("directory with file preserved", func(t *testing.T) {
		dirWithFile := filepath.Join(baseDir, "with-file")
		if err := os.Mkdir(dirWithFile, 0755); err != nil {
			t.Fatalf("Failed to create dir: %v", err)
		}

		testFile := filepath.Join(dirWithFile, "test.txt")
		if err := os.WriteFile(testFile, []byte("content"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		if err := cleanEmptyDirs(baseDir); err != nil {
			t.Errorf("cleanEmptyDirs failed: %v", err)
		}

		// Verify directory with file was NOT deleted
		if _, err := os.Stat(dirWithFile); os.IsNotExist(err) {
			t.Errorf("Directory with file was deleted but should be preserved")
		}

		// Clean up for next test
		os.RemoveAll(dirWithFile)
	})

	// Test case 4: Mixed - empty dirs alongside dirs with content
	t.Run("mixed empty and non-empty directories", func(t *testing.T) {
		// Create structure:
		// baseDir/
		//   empty1/
		//   nonempty/
		//     file.txt
		//     subdir_empty/
		//     subdir_nonempty/
		//       another.txt

		empty1 := filepath.Join(baseDir, "empty1")
		nonempty := filepath.Join(baseDir, "nonempty")
		subdirEmpty := filepath.Join(nonempty, "subdir_empty")
		subdirNonempty := filepath.Join(nonempty, "subdir_nonempty")

		os.Mkdir(empty1, 0755)
		os.MkdirAll(subdirEmpty, 0755)
		os.MkdirAll(subdirNonempty, 0755)

		os.WriteFile(filepath.Join(nonempty, "file.txt"), []byte("test"), 0644)
		os.WriteFile(filepath.Join(subdirNonempty, "another.txt"), []byte("test"), 0644)

		if err := cleanEmptyDirs(baseDir); err != nil {
			t.Errorf("cleanEmptyDirs failed: %v", err)
		}

		// Verify empty1 was deleted
		if _, err := os.Stat(empty1); !os.IsNotExist(err) {
			t.Errorf("Empty directory empty1 should be deleted")
		}

		// Verify nonempty still exists
		if _, err := os.Stat(nonempty); os.IsNotExist(err) {
			t.Errorf("Non-empty directory was deleted")
		}

		// Verify subdir_empty was deleted
		if _, err := os.Stat(subdirEmpty); !os.IsNotExist(err) {
			t.Errorf("Empty subdirectory should be deleted")
		}

		// Verify subdir_nonempty still exists
		if _, err := os.Stat(subdirNonempty); os.IsNotExist(err) {
			t.Errorf("Non-empty subdirectory was deleted")
		}

		// Clean up
		os.RemoveAll(nonempty)
	})

	// Test case 5: Becomes empty after cleaning subdirs
	t.Run("parent becomes empty after cleaning children", func(t *testing.T) {
		// Create structure:
		// baseDir/
		//   parent/
		//     child1/  (empty)
		//     child2/  (empty)
		// After cleaning, parent should also be deleted

		parent := filepath.Join(baseDir, "parent")
		child1 := filepath.Join(parent, "child1")
		child2 := filepath.Join(parent, "child2")

		os.MkdirAll(child1, 0755)
		os.MkdirAll(child2, 0755)

		if err := cleanEmptyDirs(baseDir); err != nil {
			t.Errorf("cleanEmptyDirs failed: %v", err)
		}

		// Verify parent was deleted (became empty after children removed)
		if _, err := os.Stat(parent); !os.IsNotExist(err) {
			t.Errorf("Parent directory should be deleted after children removed")
		}
	})
}

func TestCleanEmptyDirsEdgeCases(t *testing.T) {
	t.Run("non-existent directory", func(t *testing.T) {
		err := cleanEmptyDirs("/tmp/definitely-does-not-exist-12345")
		if err == nil {
			t.Error("Expected error for non-existent directory")
		}
	})

	t.Run("empty base directory", func(t *testing.T) {
		baseDir, err := os.MkdirTemp("", "test-empty-base-*")
		if err != nil {
			t.Fatalf("Failed to create temp base dir: %v", err)
		}
		defer os.RemoveAll(baseDir)

		// cleanEmptyDirs on already empty dir should succeed and not delete baseDir
		if err := cleanEmptyDirs(baseDir); err != nil {
			t.Errorf("cleanEmptyDirs failed on empty base: %v", err)
		}

		// Verify baseDir still exists
		if _, err := os.Stat(baseDir); os.IsNotExist(err) {
			t.Errorf("Base directory should not be deleted")
		}
	})
}

func TestPrepImageName(t *testing.T) {
	t.Run("valid request", func(t *testing.T) {
		req := UploadRequest{
			Guild:  123,
			User:   456,
			CharID: "507f1f77bcf86cd799439011",
		}

		parts, err := prepImageNameParts(req)
		if err != nil {
			t.Errorf("prepImageNameParts failed: %v", err)
		}

		// Verify path structure: charID/imageID.webp
		if len(parts) != 2 {
			t.Errorf("Expected 2 path parts, got %d: %v", len(parts), parts)
		}

		if parts[0] != "507f1f77bcf86cd799439011" {
			t.Errorf("Expected charID, got %s", parts[0])
		}
		if !strings.HasSuffix(parts[1], ".webp") {
			t.Errorf("Expected .webp extension, got %s", parts[1])
		}

		// Verify imageID is valid ObjectID (24 hex chars + .webp)
		imageID := strings.TrimSuffix(parts[1], ".webp")
		if len(imageID) != 24 {
			t.Errorf("Expected 24-char ObjectID, got %d chars: %s", len(imageID), imageID)
		}
		if !checks.IsValidObjectId(imageID) {
			t.Errorf("ImageID is not a valid ObjectID: %s", imageID)
		}
	})

	t.Run("invalid CharID", func(t *testing.T) {
		req := UploadRequest{
			Guild:  123,
			User:   456,
			CharID: "invalid-not-objectid",
		}

		_, err := prepImageNameParts(req)
		if err == nil {
			t.Error("Expected error for invalid CharID")
		}
		if !strings.Contains(err.Error(), "not a valid character ID") {
			t.Errorf("Unexpected error message: %v", err)
		}
	})

	t.Run("zero values", func(t *testing.T) {
		req := UploadRequest{
			Guild:  0,
			User:   0,
			CharID: "507f1f77bcf86cd799439011",
		}

		parts, err := prepImageNameParts(req)
		if err != nil {
			t.Errorf("prepImageNameParts failed with zero values: %v", err)
		}

		if parts[0] != "507f1f77bcf86cd799439011" {
			t.Errorf("CharID not correct: got %s", parts[0])
		}
		if !strings.HasSuffix(parts[1], ".webp") {
			t.Errorf("Image name should end with .webp: %s", parts[1])
		}
	})

	t.Run("large numbers", func(t *testing.T) {
		req := UploadRequest{
			Guild:  999999999,
			User:   888888888,
			CharID: "507f1f77bcf86cd799439011",
		}

		parts, err := prepImageNameParts(req)
		if err != nil {
			t.Errorf("prepImageNameParts failed with large numbers: %v", err)
		}

		// With new format, guild/user are not in the path anymore
		if parts[0] != "507f1f77bcf86cd799439011" {
			t.Errorf("CharID not correct: got %s", parts[0])
		}
		if !strings.HasSuffix(parts[1], ".webp") {
			t.Errorf("Image name should end with .webp: %s", parts[1])
		}
	})
}

func TestHandleSingleDelete(t *testing.T) {
	t.Run("successful deletion", func(t *testing.T) {
		// Setup: Create temp dir and test image
		tmpDir, err := os.MkdirTemp("", "test-delete-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		imagePath := "123/456/abc123def456789012345678/test.webp"
		fullPath := filepath.Join(tmpDir, imagePath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("Failed to create directories: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte("test image"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		cfg := &Config{ImagesDir: tmpDir, BaseURL: "https://example.com", Quality: 90}
		router := setupRouter(cfg)

		// Execute: DELETE request
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/image/"+imagePath, nil)
		router.ServeHTTP(w, req)

		// Assert: 200 OK and file deleted
		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
		}
		if checks.PathExists(fullPath) {
			t.Error("File should have been deleted")
		}

		// Verify response contains deleted path
		if !strings.Contains(w.Body.String(), imagePath) {
			t.Errorf("Response should mention deleted path: %s", w.Body.String())
		}
	})

	t.Run("non-existent image", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test-delete-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		cfg := &Config{ImagesDir: tmpDir, BaseURL: "https://example.com", Quality: 90}
		router := setupRouter(cfg)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/image/123/456/abc123/nonexistent.webp", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
		if !strings.Contains(w.Body.String(), "not found") {
			t.Errorf("Expected 'not found' error: %s", w.Body.String())
		}
	})

	t.Run("path traversal attack", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test-delete-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		cfg := &Config{ImagesDir: tmpDir, BaseURL: "https://example.com", Quality: 90}
		router := setupRouter(cfg)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/image/../../../etc/passwd", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for path traversal, got %d", w.Code)
		}
	})

	t.Run("empty directory cleanup", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test-delete-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		// Create nested structure
		imagePath := "123/456/abc123def456789012345678/test.webp"
		fullPath := filepath.Join(tmpDir, imagePath)
		if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
			t.Fatalf("Failed to create directories: %v", err)
		}
		if err := os.WriteFile(fullPath, []byte("test"), 0644); err != nil {
			t.Fatalf("Failed to create test file: %v", err)
		}

		cfg := &Config{ImagesDir: tmpDir, BaseURL: "https://example.com", Quality: 90}
		router := setupRouter(cfg)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/image/"+imagePath, nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Verify empty parent directories were cleaned up
		guildDir := filepath.Join(tmpDir, "123")
		if checks.PathExists(guildDir) {
			t.Error("Empty parent directories should have been cleaned up")
		}
	})

	t.Run("cannot delete directory", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test-delete-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		// Create a directory path (not a file)
		dirPath := "123/456/abc123def456789012345678"
		fullPath := filepath.Join(tmpDir, dirPath)
		if err := os.MkdirAll(fullPath, 0755); err != nil {
			t.Fatalf("Failed to create directories: %v", err)
		}

		cfg := &Config{ImagesDir: tmpDir, BaseURL: "https://example.com", Quality: 90}
		router := setupRouter(cfg)

		// Attempt to delete the directory using single delete endpoint
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/image/"+dirPath, nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 when deleting directory, got %d", w.Code)
		}
		if !strings.Contains(w.Body.String(), "Cannot delete directory") {
			t.Errorf("Expected 'Cannot delete directory' error: %s", w.Body.String())
		}

		// Verify directory still exists
		if !checks.DirExists(fullPath) {
			t.Error("Directory should not have been deleted")
		}
	})
}

func TestHandleCharacterDelete(t *testing.T) {
	t.Run("successful deletion", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test-char-delete-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		// Create character directory with multiple images
		charPath := filepath.Join(tmpDir, "507f1f77bcf86cd799439011")
		if err := os.MkdirAll(charPath, 0755); err != nil {
			t.Fatalf("Failed to create char dir: %v", err)
		}
		os.WriteFile(filepath.Join(charPath, "image1.webp"), []byte("test1"), 0644)
		os.WriteFile(filepath.Join(charPath, "image2.webp"), []byte("test2"), 0644)
		os.WriteFile(filepath.Join(charPath, "image3.webp"), []byte("test3"), 0644)

		cfg := &Config{ImagesDir: tmpDir, BaseURL: "https://example.com", Quality: 90}
		router := setupRouter(cfg)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/character/507f1f77bcf86cd799439011", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d: %s", w.Code, w.Body.String())
		}
		if checks.PathExists(charPath) {
			t.Error("Character directory should have been deleted")
		}
	})

	t.Run("invalid CharID", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test-char-delete-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		cfg := &Config{ImagesDir: tmpDir, BaseURL: "https://example.com", Quality: 90}
		router := setupRouter(cfg)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/character/not-a-valid-objectid", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
		if !strings.Contains(w.Body.String(), "Invalid character ID") {
			t.Errorf("Expected 'Invalid character ID' error: %s", w.Body.String())
		}
	})

	t.Run("non-existent character", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test-char-delete-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		cfg := &Config{ImagesDir: tmpDir, BaseURL: "https://example.com", Quality: 90}
		router := setupRouter(cfg)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/character/507f1f77bcf86cd799439011", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400, got %d", w.Code)
		}
		if !strings.Contains(w.Body.String(), "not found") {
			t.Errorf("Expected 'not found' error: %s", w.Body.String())
		}
	})

	t.Run("path traversal attack via params", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test-char-delete-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		cfg := &Config{ImagesDir: tmpDir, BaseURL: "https://example.com", Quality: 90}
		router := setupRouter(cfg)

		// Try to use path traversal via charID param (e.g. charID="../../../etc/passwd")
		// This should be blocked by checks.AbsPath
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/character/../../../etc/passwd", nil)
		router.ServeHTTP(w, req)

		// Should either be 400 (blocked by path validation) or 404 (route not matched)
		if w.Code != http.StatusBadRequest && w.Code != http.StatusNotFound {
			t.Errorf("Expected status 400 or 404 for path traversal, got %d", w.Code)
		}
	})

	t.Run("empty directory cleanup", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test-char-delete-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		// Create character directory
		charPath := filepath.Join(tmpDir, "507f1f77bcf86cd799439011")
		if err := os.MkdirAll(charPath, 0755); err != nil {
			t.Fatalf("Failed to create char dir: %v", err)
		}
		os.WriteFile(filepath.Join(charPath, "image.webp"), []byte("test"), 0644)

		cfg := &Config{ImagesDir: tmpDir, BaseURL: "https://example.com", Quality: 90}
		router := setupRouter(cfg)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("DELETE", "/character/507f1f77bcf86cd799439011", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Errorf("Expected status 200, got %d", w.Code)
		}

		// Verify character directory was deleted
		if checks.PathExists(charPath) {
			t.Error("Character directory should have been deleted")
		}
	})
}

func TestHandleImageUpload(t *testing.T) {
	t.Run("successful upload", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test-upload-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		cfg := &Config{ImagesDir: tmpDir, BaseURL: "https://example.com", Quality: 90}
		router := setupRouter(cfg)

		uploadReq := UploadRequest{
			Guild:    123,
			User:     456,
			CharID:   "507f1f77bcf86cd799439011",
			ImageURL: "https://i.tiltowait.dev/avatar.jpg",
		}

		body, _ := json.Marshal(uploadReq)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/image/upload", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Errorf("Expected status 201, got %d: %s", w.Code, w.Body.String())
		}

		// Verify response contains URL
		responseURL := strings.Trim(w.Body.String(), "\"")
		if !strings.HasPrefix(responseURL, "https://example.com/507f1f77bcf86cd799439011/") {
			t.Errorf("Unexpected response URL format: %s", responseURL)
		}
		if !strings.HasSuffix(responseURL, ".webp") {
			t.Errorf("Response URL should end with .webp: %s", responseURL)
		}

		// Verify file was actually created
		// Extract path components from URL (they use forward slashes)
		urlPath := strings.TrimPrefix(responseURL, "https://example.com/")
		// Split on forward slashes and rejoin with OS-specific separator
		pathParts := strings.Split(urlPath, "/")
		fullPath := filepath.Join(append([]string{tmpDir}, pathParts...)...)
		if !checks.PathExists(fullPath) {
			t.Errorf("Image file was not created at: %s", fullPath)
		}
	})

	t.Run("invalid JSON", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test-upload-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		cfg := &Config{ImagesDir: tmpDir, BaseURL: "https://example.com", Quality: 90}
		router := setupRouter(cfg)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/image/upload", bytes.NewBufferString("{invalid json"))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for invalid JSON, got %d", w.Code)
		}
	})

	t.Run("invalid URL scheme", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test-upload-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		cfg := &Config{ImagesDir: tmpDir, BaseURL: "https://example.com", Quality: 90}
		router := setupRouter(cfg)

		uploadReq := UploadRequest{
			Guild:    123,
			User:     456,
			CharID:   "507f1f77bcf86cd799439011",
			ImageURL: "file:///etc/passwd",
		}

		body, _ := json.Marshal(uploadReq)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/image/upload", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for file:// URL, got %d", w.Code)
		}
		if !strings.Contains(w.Body.String(), "invalid image URL") {
			t.Errorf("Expected 'invalid image URL' error: %s", w.Body.String())
		}
	})

	t.Run("invalid CharID", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test-upload-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		cfg := &Config{ImagesDir: tmpDir, BaseURL: "https://example.com", Quality: 90}
		router := setupRouter(cfg)

		uploadReq := UploadRequest{
			Guild:    123,
			User:     456,
			CharID:   "not-a-valid-objectid",
			ImageURL: "https://i.tiltowait.dev/avatar.jpg",
		}

		body, _ := json.Marshal(uploadReq)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/image/upload", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Errorf("Expected status 400 for invalid CharID, got %d", w.Code)
		}
		if !strings.Contains(w.Body.String(), "not a valid character ID") {
			t.Errorf("Expected 'not a valid character ID' error: %s", w.Body.String())
		}
	})

	t.Run("download non-existent URL", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test-upload-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		cfg := &Config{ImagesDir: tmpDir, BaseURL: "https://example.com", Quality: 90}
		router := setupRouter(cfg)

		uploadReq := UploadRequest{
			Guild:    123,
			User:     456,
			CharID:   "507f1f77bcf86cd799439011",
			ImageURL: "https://example.com/definitely-does-not-exist-12345.jpg",
		}

		body, _ := json.Marshal(uploadReq)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/image/upload", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusBadGateway {
			t.Errorf("Expected status 502 for download failure, got %d", w.Code)
		}
		if !strings.Contains(w.Body.String(), "failed to download image") {
			t.Errorf("Expected 'failed to download image' error: %s", w.Body.String())
		}
	})

	t.Run("non-image content", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test-upload-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		cfg := &Config{ImagesDir: tmpDir, BaseURL: "https://example.com", Quality: 90}
		router := setupRouter(cfg)

		uploadReq := UploadRequest{
			Guild:    123,
			User:     456,
			CharID:   "507f1f77bcf86cd799439011",
			ImageURL: "https://tilt-assets.s3-us-west-1.amazonaws.com/b.txt",
		}

		body, _ := json.Marshal(uploadReq)
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("POST", "/image/upload", bytes.NewBuffer(body))
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Errorf("Expected status 500 for non-image content, got %d: %s", w.Code, w.Body.String())
		}
	})

	t.Run("duplicate upload same location", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test-upload-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		// Pre-create a file to simulate existing image
		existingPath := filepath.Join(tmpDir, "123", "456", "507f1f77bcf86cd799439011", "existing.webp")
		if err := os.MkdirAll(filepath.Dir(existingPath), 0755); err != nil {
			t.Fatalf("Failed to create directories: %v", err)
		}
		if err := os.WriteFile(existingPath, []byte("existing"), 0644); err != nil {
			t.Fatalf("Failed to create existing file: %v", err)
		}

		// Note: This test verifies the behavior, but actual duplicate prevention
		// happens via unique ObjectID generation, not by checking existing files.
		// The SaveWebP function will reject if the exact file already exists.
	})
}

func TestSetupRouter(t *testing.T) {
	t.Run("routes registered", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test-router-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		cfg := &Config{ImagesDir: tmpDir, BaseURL: "https://example.com", Quality: 90}
		router := setupRouter(cfg)

		// Test POST /image/upload exists
		w := httptest.NewRecorder()
		body := bytes.NewBufferString(`{"guild":1,"user":2,"charid":"507f1f77bcf86cd799439011","image_url":"invalid"}`)
		req, _ := http.NewRequest("POST", "/image/upload", body)
		req.Header.Set("Content-Type", "application/json")
		router.ServeHTTP(w, req)
		// Should not be 404 (might be 400 due to invalid URL, but route exists)
		if w.Code == http.StatusNotFound {
			t.Error("POST /image/upload route not registered")
		}

		// Test DELETE /image/* exists
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("DELETE", "/image/test/path.webp", nil)
		router.ServeHTTP(w, req)
		if w.Code == http.StatusNotFound {
			t.Error("DELETE /image/* route not registered")
		}

		// Test DELETE /character/:charID exists
		w = httptest.NewRecorder()
		req, _ = http.NewRequest("DELETE", "/character/507f1f77bcf86cd799439011", nil)
		router.ServeHTTP(w, req)
		// Should not be 404 (might be 400 due to non-existent dir, but route exists)
		if w.Code == http.StatusNotFound {
			t.Error("DELETE /character/:charID route not registered")
		}
	})

	t.Run("404 for non-existent route", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test-router-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		cfg := &Config{ImagesDir: tmpDir, BaseURL: "https://example.com", Quality: 90}
		router := setupRouter(cfg)

		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/nonexistent", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusNotFound {
			t.Errorf("Expected 404 for non-existent route, got %d", w.Code)
		}
	})

	t.Run("wrong method returns 405", func(t *testing.T) {
		tmpDir, err := os.MkdirTemp("", "test-router-*")
		if err != nil {
			t.Fatalf("Failed to create temp dir: %v", err)
		}
		defer os.RemoveAll(tmpDir)

		cfg := &Config{ImagesDir: tmpDir, BaseURL: "https://example.com", Quality: 90}
		router := setupRouter(cfg)

		// Try GET on POST-only route
		w := httptest.NewRecorder()
		req, _ := http.NewRequest("GET", "/image/upload", nil)
		router.ServeHTTP(w, req)

		if w.Code != http.StatusMethodNotAllowed && w.Code != http.StatusNotFound {
			t.Errorf("Expected 405 or 404 for wrong method, got %d", w.Code)
		}
	})
}
