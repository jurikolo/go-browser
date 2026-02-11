package utils

import "strings"

// SetUrlPrefix ensures that a URL has a proper prefix (http://, https://, or /)
// If the URL doesn't have a prefix, it will be prefixed with "https://"
func SetUrlPrefix(url string) string {
	if !(strings.HasPrefix(url, "http://") || strings.HasPrefix(url, "https://") || strings.HasPrefix(url, "/")) {
		return "https://" + url
	}
	return url
}