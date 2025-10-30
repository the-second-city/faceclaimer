package cmd

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestPreRunE_BaseURLValidation(t *testing.T) {
	tests := []struct {
		name      string
		baseURL   string
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid https URL",
			baseURL:   "https://images.example.com",
			wantError: false,
		},
		{
			name:      "valid http URL",
			baseURL:   "http://localhost",
			wantError: false,
		},
		{
			name:      "valid URL with port",
			baseURL:   "https://example.com:8080",
			wantError: false,
		},
		{
			name:      "valid URL with path",
			baseURL:   "https://example.com/images",
			wantError: false,
		},
		{
			name:      "empty baseURL",
			baseURL:   "",
			wantError: true,
			errorMsg:  "base-url must be a valid URL",
		},
		{
			name:      "invalid URL no scheme",
			baseURL:   "example.com",
			wantError: true,
			errorMsg:  "base-url must be a valid URL",
		},
		{
			name:      "invalid ftp scheme",
			baseURL:   "ftp://example.com",
			wantError: true,
			errorMsg:  "base-url must be a valid URL",
		},
		{
			name:      "invalid file scheme",
			baseURL:   "file:///path/to/images",
			wantError: true,
			errorMsg:  "base-url must be a valid URL",
		},
		{
			name:      "malformed URL",
			baseURL:   "ht!tp://invalid",
			wantError: true,
			errorMsg:  "base-url must be a valid URL",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for tests
			tmpDir, err := os.MkdirTemp("", "test-images-*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			// Reset flags to defaults
			baseURL = tt.baseURL
			imagesDir = tmpDir
			quality = 90

			// Call PreRunE
			err = rootCmd.PreRunE(rootCmd, []string{})

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error containing %q, got nil", tt.errorMsg)
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}

func TestPreRunE_ImagesDirValidation(t *testing.T) {
	tests := []struct {
		name      string
		setupDir  func() string
		wantError bool
		errorMsg  string
	}{
		{
			name: "valid existing directory",
			setupDir: func() string {
				dir, _ := os.MkdirTemp("", "test-images-*")
				return dir
			},
			wantError: false,
		},
		{
			name: "non-existent directory",
			setupDir: func() string {
				return "/tmp/nonexistent-faceclaimer-test-dir-12345"
			},
			wantError: true,
			errorMsg:  "images-dir does not exist",
		},
		{
			name: "file instead of directory",
			setupDir: func() string {
				f, _ := os.CreateTemp("", "test-file-*")
				defer f.Close()
				return f.Name()
			},
			wantError: true,
			errorMsg:  "images-dir does not exist",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			testPath := tt.setupDir()
			if !tt.wantError {
				defer os.RemoveAll(testPath)
			}

			// Reset flags to defaults
			baseURL = "https://example.com"
			imagesDir = testPath
			quality = 90

			// Call PreRunE
			err := rootCmd.PreRunE(rootCmd, []string{})

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error containing %q, got nil", tt.errorMsg)
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}

func TestPreRunE_QualityValidation(t *testing.T) {
	tests := []struct {
		name      string
		quality   int
		wantError bool
		errorMsg  string
	}{
		{
			name:      "valid quality 1",
			quality:   1,
			wantError: false,
		},
		{
			name:      "valid quality 50",
			quality:   50,
			wantError: false,
		},
		{
			name:      "valid quality 90",
			quality:   90,
			wantError: false,
		},
		{
			name:      "valid quality 100",
			quality:   100,
			wantError: false,
		},
		{
			name:      "invalid quality 0",
			quality:   0,
			wantError: true,
			errorMsg:  "quality must be between 1 and 100",
		},
		{
			name:      "invalid quality -1",
			quality:   -1,
			wantError: true,
			errorMsg:  "quality must be between 1 and 100",
		},
		{
			name:      "invalid quality 101",
			quality:   101,
			wantError: true,
			errorMsg:  "quality must be between 1 and 100",
		},
		{
			name:      "invalid quality 1000",
			quality:   1000,
			wantError: true,
			errorMsg:  "quality must be between 1 and 100",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create temporary directory for tests
			tmpDir, err := os.MkdirTemp("", "test-images-*")
			if err != nil {
				t.Fatalf("Failed to create temp dir: %v", err)
			}
			defer os.RemoveAll(tmpDir)

			// Reset flags to defaults
			baseURL = "https://example.com"
			imagesDir = tmpDir
			quality = tt.quality

			// Call PreRunE
			err = rootCmd.PreRunE(rootCmd, []string{})

			if tt.wantError {
				if err == nil {
					t.Errorf("Expected error containing %q, got nil", tt.errorMsg)
				} else if !strings.Contains(err.Error(), tt.errorMsg) {
					t.Errorf("Expected error containing %q, got %q", tt.errorMsg, err.Error())
				}
			} else {
				if err != nil {
					t.Errorf("Expected no error, got: %v", err)
				}
			}
		})
	}
}

func TestPreRunE_AbsolutePathResolution(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "test-images-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Use a relative path
	relPath := filepath.Base(tmpDir)

	// Change to parent directory so relative path works
	originalWd, err := os.Getwd()
	if err != nil {
		t.Fatalf("Failed to get working directory: %v", err)
	}
	defer os.Chdir(originalWd)

	err = os.Chdir(filepath.Dir(tmpDir))
	if err != nil {
		t.Fatalf("Failed to change directory: %v", err)
	}

	// Reset flags with relative path
	baseURL = "https://example.com"
	imagesDir = relPath
	quality = 90

	// Call PreRunE
	err = rootCmd.PreRunE(rootCmd, []string{})
	if err != nil {
		t.Fatalf("PreRunE failed: %v", err)
	}

	// Verify imagesDir was converted to absolute path
	if !filepath.IsAbs(imagesDir) {
		t.Errorf("Expected imagesDir to be absolute path, got: %s", imagesDir)
	}

	// Verify it points to the correct directory
	expectedAbs, _ := filepath.Abs(relPath)
	if imagesDir != expectedAbs {
		t.Errorf("Expected imagesDir to be %q, got %q", expectedAbs, imagesDir)
	}
}

func TestPreRunE_AllValidationsCombined(t *testing.T) {
	// Create temporary directory
	tmpDir, err := os.MkdirTemp("", "test-images-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name      string
		baseURL   string
		imagesDir string
		quality   int
		wantError bool
	}{
		{
			name:      "all valid",
			baseURL:   "https://example.com",
			imagesDir: tmpDir,
			quality:   90,
			wantError: false,
		},
		{
			name:      "invalid baseURL valid others",
			baseURL:   "not-a-url",
			imagesDir: tmpDir,
			quality:   90,
			wantError: true,
		},
		{
			name:      "valid baseURL invalid dir",
			baseURL:   "https://example.com",
			imagesDir: "/nonexistent/dir",
			quality:   90,
			wantError: true,
		},
		{
			name:      "valid baseURL and dir invalid quality",
			baseURL:   "https://example.com",
			imagesDir: tmpDir,
			quality:   0,
			wantError: true,
		},
		{
			name:      "all invalid",
			baseURL:   "invalid",
			imagesDir: "/nonexistent",
			quality:   150,
			wantError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Reset flags
			baseURL = tt.baseURL
			imagesDir = tt.imagesDir
			quality = tt.quality

			// Call PreRunE
			err := rootCmd.PreRunE(rootCmd, []string{})

			if tt.wantError && err == nil {
				t.Error("Expected error, got nil")
			}
			if !tt.wantError && err != nil {
				t.Errorf("Expected no error, got: %v", err)
			}
		})
	}
}

