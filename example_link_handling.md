# Link Handling Example

This document demonstrates how to use the enhanced link handling functionality in the go-browser.

## Features Implemented

1. **Enhanced Link Structure**:
   - Number: int (for keyboard selection)
   - Text: string (display text)
   - URL: string (absolute URL)
   - Type: string ("internal", "external", "anchor")

2. **Link Extraction**:
   - Parse HTML from ChromeDP
   - Find all <a> tags
   - Resolve relative URLs to absolute
   - Filter out javascript: and mailto: links
   - Assign sequential numbers
   - Group by type (internal vs external)

3. **Link Display**:
   - Create a section showing all numbered links
   - Format as: [1] Link Text (URL)
   - Optionally truncate long URLs
   - Add visual separators for link groups

4. **Link Navigation**:
   - Find link by number
   - Navigate using ChromeDP
   - Handle external links (open vs. show warning)

5. **Link Search**:
   - Filter links by text or URL matching query
   - Useful for pages with many links

6. **Special Link Handling**:
   - Anchor links (#section): Smooth scroll to element
   - File downloads: Show download prompt
   - External domains: Optional confirmation

## Usage Example

```go
// Extract links from the current page
links, err := browser.ExtractLinks()
if err != nil {
    log.Printf("Error extracting links: %v", err)
    return
}

// Display links
fmt.Println(browser.FormatLinksForDisplay(links, 80))

// Navigate to a specific link
err = browser.NavigateToLink(links, 1, true) // true = confirm external links
if err != nil {
    log.Printf("Error navigating to link: %v", err)
}

// Search for specific links
searchResults := browser.SearchLinks(links, "example")
fmt.Printf("Found %d matching links\n", len(searchResults))

// Check if a link is a download
for _, link := range links {
    if browser.IsDownloadLink(link) {
        fmt.Printf("Download link found: %s\n", link.URL)
    }
}
```

## Integration with UI

The link handling is integrated with the Bubble Tea UI through:

1. **LinkClickMsg**: Sent when a user presses a number key (1-9)
2. **NavigateMsg**: Sent when navigating to a URL
3. **EnhancedLink struct**: Contains all necessary metadata for link handling

The UI automatically displays links below the main content and allows keyboard navigation using number keys 1-9.