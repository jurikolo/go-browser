package browser

import (
	"fmt"
	"log"
	"os"
	"path/filepath"
)

// Example of how to use the history system
func ExampleHistory() {
	// Create a new history with a maximum size of 50 entries
	history := NewHistory(50)
	
	// Add some entries to the history
	history.Add("https://example.com", "Example Domain")
	history.Add("https://google.com", "Google")
	history.Add("https://github.com", "GitHub")
	
	// Check if we can go back
	if history.CanGoBack() {
		fmt.Printf("Can go back: true\n")
	}
	
	// Go back in history
	entry, err := history.Back()
	if err != nil {
		log.Printf("Error going back: %v", err)
	} else {
		fmt.Printf("Went back to: %s (%s)\n", entry.Title, entry.URL)
	}
	
	// Check if we can go forward
	if history.CanGoForward() {
		fmt.Printf("Can go forward: true\n")
	}
	
	// Go forward in history
	entry, err = history.Forward()
	if err != nil {
		log.Printf("Error going forward: %v", err)
	} else {
		fmt.Printf("Went forward to: %s (%s)\n", entry.Title, entry.URL)
	}
	
	// Set and get scroll position for current entry
	history.SetScrollPosition(500)
	scrollPos := history.GetScrollPosition()
	fmt.Printf("Current scroll position: %d\n", scrollPos)
	
	// Save history to disk
	if err := history.SaveToDisk(); err != nil {
		log.Printf("Error saving history: %v", err)
	} else {
		fmt.Printf("History saved to disk\n")
	}
	
	// Clean up test file
	homeDir, _ := os.UserHomeDir()
	historyFile := filepath.Join(homeDir, ".config", "chromedp-browser", "history.json")
	os.Remove(historyFile)
	
	// Output:
	// Can go back: true
	// Went back to: Google (https://google.com)
	// Can go forward: true
	// Went forward to: GitHub (https://github.com)
	// Current scroll position: 500
	// History saved to disk
}