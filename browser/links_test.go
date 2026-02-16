package browser

import (
	"testing"
)

func TestEnhancedLinkStruct(t *testing.T) {
	link := EnhancedLink{
		Number: 1,
		Text:   "Example Link",
		URL:    "https://example.com",
		Type:   "external",
	}
	
	if link.Number != 1 {
		t.Errorf("Expected Number to be 1, got %d", link.Number)
	}
	
	if link.Text != "Example Link" {
		t.Errorf("Expected Text to be 'Example Link', got %s", link.Text)
	}
	
	if link.URL != "https://example.com" {
		t.Errorf("Expected URL to be 'https://example.com', got %s", link.URL)
	}
	
	if link.Type != "external" {
		t.Errorf("Expected Type to be 'external', got %s", link.Type)
	}
}

func TestIsDownloadLink(t *testing.T) {
	testCases := []struct {
		link     EnhancedLink
		expected bool
	}{
		{EnhancedLink{URL: "https://example.com/file.pdf"}, true},
		{EnhancedLink{URL: "https://example.com/file.docx"}, true},
		{EnhancedLink{URL: "https://example.com/file.zip"}, true},
		{EnhancedLink{URL: "https://example.com/page.html"}, false},
		{EnhancedLink{URL: "https://example.com/"}, false},
	}
	
	for _, tc := range testCases {
		result := IsDownloadLink(tc.link)
		if result != tc.expected {
			t.Errorf("IsDownloadLink(%s) = %v; expected %v", tc.link.URL, result, tc.expected)
		}
	}
}

func TestFormatLinksForDisplay(t *testing.T) {
	links := []EnhancedLink{
		{Number: 1, Text: "Internal Link", URL: "https://example.com/page1", Type: "internal"},
		{Number: 2, Text: "External Link", URL: "https://google.com", Type: "external"},
		{Number: 3, Text: "Anchor Link", URL: "https://example.com#section", Type: "anchor"},
	}
	
	result := FormatLinksForDisplay(links, 50)
	
	if len(result) == 0 {
		t.Error("Expected non-empty result")
	}
	
	if len(links) != 3 {
		t.Errorf("Expected 3 links, got %d", len(links))
	}
}