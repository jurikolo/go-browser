package utils

import (
	"net/url"
	"strings"
)

// SetUrlPrefix ensures that a URL has a proper prefix (http://, https://, file://, or /)
// If the URL doesn't have a prefix, it will be prefixed with "https://"
func SetUrlPrefix(urlStr string) string {
	// Check if it's already a valid URL with a scheme
	if strings.HasPrefix(urlStr, "http://") || strings.HasPrefix(urlStr, "https://") || strings.HasPrefix(urlStr, "file://") {
		return urlStr
	}
	
	// Check if it's an absolute path (Unix/Linux)
	if strings.HasPrefix(urlStr, "/") {
		return "file://" + urlStr
	}
	
	// Check if it looks like a file path that should be treated as a local file
	// This includes paths that end with .html, .htm, or contain a dot in the last segment
	if isLikelyFilePath(urlStr) {
		// Convert to file:// URL
		return "file://" + urlStr
	}
	
	// Default to https:// for web URLs
	return "https://" + urlStr
}

// isLikelyFilePath checks if a string looks like a file path that should be treated as a local file
func isLikelyFilePath(path string) bool {
	// Check if it's an absolute path
	if strings.HasPrefix(path, "/") {
		return true
	}
	
	// Check if it has a file extension commonly associated with HTML files
	if strings.HasSuffix(path, ".html") || strings.HasSuffix(path, ".htm") {
		return true
	}
	
	// Parse as URL to check if it has a scheme
	if u, err := url.Parse(path); err == nil && u.Scheme != "" {
		// If it has a scheme, it's not a file path
		return false
	}
	
	return false
}