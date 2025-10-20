// checks has various validators.
package checks

import (
	"net/url"
	"os"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// IsValidURL returns true if the string is a valid URL with http or https scheme.
func IsValidURL(urlStr string) bool {
	u, err := url.Parse(urlStr)
	if err != nil {
		return false
	}
	// Ensure the URL has http/https scheme and a host
	return (u.Scheme == "http" || u.Scheme == "https") && u.Host != ""
}

// PathExists returns true if the given path exists (file or directory).
func PathExists(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

// DirExists returns true if the given directory exists.
func DirExists(dir string) bool {
	info, err := os.Stat(dir)
	if err != nil {
		return false
	}
	return info.IsDir()
}

// FileExists returns true if the given file exists (not a directory).
func FileExists(file string) bool {
	info, err := os.Stat(file)
	if err != nil {
		return false
	}
	return !info.IsDir()
}

// IsValidObjectId returns true if the string represents a valid ObjectId.
func IsValidObjectId(oid string) bool {
	_, err := primitive.ObjectIDFromHex(oid)
	return err == nil
}
