package browser

import (
	"fmt"
	"net/url"
	"strings"

	"github.com/chromedp/chromedp"
)

// EnhancedLink represents a hyperlink with additional metadata for user interaction
type EnhancedLink struct {
	Number int    `json:"number"` // Sequential number for keyboard selection
	Text   string `json:"text"`   // Display text of the link
	URL    string `json:"url"`    // Absolute URL
	Type   string `json:"type"`   // "internal", "external", "anchor"
}

// ExtractLinks parses HTML from ChromeDP, finds all <a> tags, resolves relative URLs to absolute,
// filters out javascript: and mailto: links, assigns sequential numbers, and groups by type
func (b *Browser) ExtractLinks() ([]EnhancedLink, error) {
	var links []EnhancedLink
	
	// Get the current page URL for resolving relative URLs
	var currentURL string
	err := chromedp.Run(b.ctx, chromedp.Location(&currentURL))
	if err != nil {
		return nil, fmt.Errorf("failed to get current URL: %w", err)
	}
	
	baseURL, err := url.Parse(currentURL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse current URL: %w", err)
	}
	
	// Extract all links from the page
	var linkElements []struct {
		Text string `json:"text"`
		URL  string `json:"url"`
	}
	
	err = chromedp.Run(b.ctx,
		chromedp.Evaluate(`
			(function() {
				var linkElements = document.querySelectorAll('a[href]');
				var links = [];
				for (var i = 0; i < linkElements.length; i++) {
					links.push({
						text: linkElements[i].textContent.trim(),
						url: linkElements[i].href
					});
				}
				return links;
			})()
		`, &linkElements),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to extract links: %w", err)
	}
	
	// Process and categorize links
	linkCounter := 1
	for _, element := range linkElements {
		// Skip javascript and mailto links
		if strings.HasPrefix(element.URL, "javascript:") || strings.HasPrefix(element.URL, "mailto:") {
			continue
		}
		
		// Resolve relative URLs to absolute
		linkURL, err := url.Parse(element.URL)
		if err != nil {
			continue // Skip invalid URLs
		}
		
		if !linkURL.IsAbs() {
			linkURL = baseURL.ResolveReference(linkURL)
		}
		
		// Determine link type
		linkType := "external"
		if linkURL.Host == baseURL.Host {
			if strings.HasPrefix(element.URL, "#") {
				linkType = "anchor"
			} else {
				linkType = "internal"
			}
		}
		
		links = append(links, EnhancedLink{
			Number: linkCounter,
			Text:   element.Text,
			URL:    linkURL.String(),
			Type:   linkType,
		})
		
		linkCounter++
	}
	
	return links, nil
}

// FormatLinksForDisplay creates a formatted string showing all numbered links
// with visual separators for link groups
func FormatLinksForDisplay(links []EnhancedLink, maxWidth int) string {
	if len(links) == 0 {
		return "No links found on this page."
	}
	
	var result strings.Builder
	result.WriteString("\n\n=== Links ===\n")
	
	currentType := ""
	for _, link := range links {
		// Add section header when link type changes
		if link.Type != currentType {
			if currentType != "" {
				result.WriteString("\n")
			}
			currentType = link.Type
			switch link.Type {
			case "internal":
				result.WriteString("--- Internal Links ---\n")
			case "external":
				result.WriteString("--- External Links ---\n")
			case "anchor":
				result.WriteString("--- Anchor Links ---\n")
			default:
				result.WriteString(fmt.Sprintf("--- %s Links ---\n", strings.Title(link.Type)))
			}
		}
		
		// Format the link display
		linkText := fmt.Sprintf("[%d] %s", link.Number, link.Text)
		linkURL := link.URL
		
		// Truncate long URLs if needed
		if maxWidth > 0 && len(linkURL) > maxWidth {
			linkURL = linkURL[:maxWidth-3] + "..."
		}
		
		result.WriteString(fmt.Sprintf("%s (%s)\n", linkText, linkURL))
	}
	
	return result.String()
}

// NavigateToLink finds a link by number and navigates to it using ChromeDP
// Handles external links with optional confirmation
func (b *Browser) NavigateToLink(links []EnhancedLink, number int, confirmExternal bool) error {
	var targetLink *EnhancedLink
	for _, link := range links {
		if link.Number == number {
			targetLink = &link
			break
		}
	}
	
	if targetLink == nil {
		return fmt.Errorf("link with number %d not found", number)
	}
	
	// Handle different link types
	switch targetLink.Type {
	case "anchor":
		// For anchor links, scroll to the element
		return chromedp.Run(b.ctx, chromedp.ScrollIntoView(fmt.Sprintf("a[href='%s']", targetLink.URL)))
		
	case "external":
		// For external links, optionally show confirmation
		if confirmExternal {
			// In a real implementation, this would show a confirmation dialog
			// For now, we'll just log it
			fmt.Printf("Warning: Navigating to external site: %s\n", targetLink.URL)
		}
		fallthrough // Continue with navigation
		
	default:
		// Navigate to the link
		return b.Navigate(targetLink.URL)
	}
}

// SearchLinks filters links by text or URL matching query
func SearchLinks(links []EnhancedLink, query string) []EnhancedLink {
	var results []EnhancedLink
	query = strings.ToLower(query)
	
	for _, link := range links {
		if strings.Contains(strings.ToLower(link.Text), query) || 
		   strings.Contains(strings.ToLower(link.URL), query) {
			results = append(results, link)
		}
	}
	
	return results
}

// IsDownloadLink checks if a link points to a downloadable file
func IsDownloadLink(link EnhancedLink) bool {
	// Common file extensions that are typically downloads
	downloadExtensions := []string{
		".pdf", ".doc", ".docx", ".xls", ".xlsx", ".ppt", ".pptx",
		".zip", ".rar", ".7z", ".tar", ".gz",
		".exe", ".dmg", ".deb", ".rpm",
		".mp3", ".mp4", ".avi", ".mov", ".wmv",
	}
	
	linkURL, err := url.Parse(link.URL)
	if err != nil {
		return false
	}
	
	path := strings.ToLower(linkURL.Path)
	for _, ext := range downloadExtensions {
		if strings.HasSuffix(path, ext) {
			return true
		}
	}
	
	return false
}