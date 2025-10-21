package routes

import (
	"os"
	"path/filepath"
	"testing"
)

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
