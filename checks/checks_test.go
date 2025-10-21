package checks

import (
	"os"
	"path/filepath"
	"testing"
)

func TestIsValidURL(t *testing.T) {
	tests := []struct {
		name     string
		url      string
		expected bool
	}{
		{"valid https", "https://example.com/image.jpg", true},
		{"valid http", "http://example.com/image.jpg", true},
		{"valid with port", "https://example.com:8080/image.jpg", true},
		{"valid with path", "https://example.com/path/to/image.jpg", true},
		{"invalid no scheme", "example.com/image.jpg", false},
		{"invalid ftp scheme", "ftp://example.com/file.txt", false},
		{"invalid no host", "https://", false},
		{"invalid empty", "", false},
		{"invalid malformed", "not a url at all", false},
		{"invalid file scheme", "file:///etc/passwd", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidURL(tt.url)
			if result != tt.expected {
				t.Errorf("IsValidURL(%q) = %v, expected %v", tt.url, result, tt.expected)
			}
		})
	}
}

func TestPathExists(t *testing.T) {
	// Create a temporary file for testing
	tmpFile, err := os.CreateTemp("", "test-path-exists-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "test-dir-exists-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"existing file", tmpPath, true},
		{"existing directory", tmpDir, true},
		{"non-existent path", "/tmp/definitely-does-not-exist-12345", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := PathExists(tt.path)
			if result != tt.expected {
				t.Errorf("PathExists(%q) = %v, expected %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestDirExists(t *testing.T) {
	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "test-dir-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test-file-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"existing directory", tmpDir, true},
		{"file not directory", tmpPath, false},
		{"non-existent path", "/tmp/definitely-does-not-exist-12345", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := DirExists(tt.path)
			if result != tt.expected {
				t.Errorf("DirExists(%q) = %v, expected %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestFileExists(t *testing.T) {
	// Create a temporary file
	tmpFile, err := os.CreateTemp("", "test-file-*.txt")
	if err != nil {
		t.Fatalf("Failed to create temp file: %v", err)
	}
	tmpPath := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpPath)

	// Create a temporary directory
	tmpDir, err := os.MkdirTemp("", "test-dir-*")
	if err != nil {
		t.Fatalf("Failed to create temp dir: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	tests := []struct {
		name     string
		path     string
		expected bool
	}{
		{"existing file", tmpPath, true},
		{"directory not file", tmpDir, false},
		{"non-existent path", "/tmp/definitely-does-not-exist-12345", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FileExists(tt.path)
			if result != tt.expected {
				t.Errorf("FileExists(%q) = %v, expected %v", tt.path, result, tt.expected)
			}
		})
	}
}

func TestIsValidObjectId(t *testing.T) {
	tests := []struct {
		name     string
		oid      string
		expected bool
	}{
		{"valid 24 char hex", "507f1f77bcf86cd799439011", true},
		{"valid uppercase", "507F1F77BCF86CD799439011", true},
		{"invalid too short", "507f1f77bcf86cd7", false},
		{"invalid too long", "507f1f77bcf86cd799439011abc", false},
		{"invalid non-hex chars", "507f1f77bcf86cd79943901g", false},
		{"invalid empty", "", false},
		{"invalid special chars", "507f1f77-bcf8-6cd7-9943-9011", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsValidObjectId(tt.oid)
			if result != tt.expected {
				t.Errorf("IsValidObjectId(%q) = %v, expected %v", tt.oid, result, tt.expected)
			}
		})
	}
}

func TestAbsPath(t *testing.T) {
	// Create a temporary base directory
	tmpBase, err := os.MkdirTemp("", "test-base-*")
	if err != nil {
		t.Fatalf("Failed to create temp base dir: %v", err)
	}
	defer os.RemoveAll(tmpBase)

	// Create a sibling directory to test prefix bypass
	tmpParent := filepath.Dir(tmpBase)
	baseName := filepath.Base(tmpBase)
	siblingDir := filepath.Join(tmpParent, baseName+"-sibling")
	err = os.Mkdir(siblingDir, 0755)
	if err != nil {
		t.Fatalf("Failed to create sibling dir: %v", err)
	}
	defer os.RemoveAll(siblingDir)

	tests := []struct {
		name      string
		base      string
		path      string
		expectErr bool
	}{
		{
			name:      "valid simple path",
			base:      tmpBase,
			path:      "test.txt",
			expectErr: false,
		},
		{
			name:      "valid nested path",
			base:      tmpBase,
			path:      "subdir/test.txt",
			expectErr: false,
		},
		{
			name:      "invalid parent traversal",
			base:      tmpBase,
			path:      "../../../etc/passwd",
			expectErr: true,
		},
		{
			name:      "absolute path becomes relative to base",
			base:      tmpBase,
			path:      "/etc/passwd",
			expectErr: false, // filepath.Join treats absolute paths specially, but result is still within base
		},
		{
			name:      "invalid sibling directory",
			base:      tmpBase,
			path:      "../" + baseName + "-sibling/malicious.txt",
			expectErr: true,
		},
		{
			name:      "valid with dot",
			base:      tmpBase,
			path:      "./test.txt",
			expectErr: false,
		},
		{
			name:      "valid with redundant separators",
			base:      tmpBase,
			path:      "subdir//test.txt",
			expectErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := AbsPath(tt.base, tt.path)

			if tt.expectErr {
				if err == nil {
					t.Errorf("AbsPath(%q, %q) expected error but got result: %q", tt.base, tt.path, result)
				}
			} else {
				if err != nil {
					t.Errorf("AbsPath(%q, %q) unexpected error: %v", tt.base, tt.path, err)
				}

				// Verify the result is an absolute path
				if !filepath.IsAbs(result) {
					t.Errorf("AbsPath(%q, %q) result is not absolute: %q", tt.base, tt.path, result)
				}

				// Verify the result is within the base directory
				absBase, _ := filepath.Abs(tt.base)
				if !filepath.HasPrefix(result, absBase) {
					t.Errorf("AbsPath(%q, %q) result %q is not within base %q", tt.base, tt.path, result, absBase)
				}
			}
		})
	}
}

func TestAbsPathWithNonExistentBase(t *testing.T) {
	// Test with a base directory that doesn't exist yet (should still work for path validation)
	nonExistentBase := "/tmp/non-existent-base-dir-12345"

	result, err := AbsPath(nonExistentBase, "test.txt")
	if err != nil {
		t.Errorf("AbsPath with non-existent base should not error: %v", err)
	}

	expected := filepath.Join(nonExistentBase, "test.txt")
	absExpected, _ := filepath.Abs(expected)
	if result != absExpected {
		t.Errorf("AbsPath result %q does not match expected %q", result, absExpected)
	}
}
