package browser

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestHistory_Add(t *testing.T) {
	h := NewHistory(10)

	// Test adding entries
	h.Add("https://example.com", "Example")
	if len(h.entries) != 1 {
		t.Errorf("Expected 1 entry, got %d", len(h.entries))
	}

	// Test adding duplicate consecutive entries (should update timestamp)
	oldTime := h.entries[0].Timestamp
	time.Sleep(10 * time.Millisecond) // Ensure timestamp changes
	h.Add("https://example.com", "Example Updated")
	if len(h.entries) != 1 {
		t.Errorf("Expected 1 entry after duplicate, got %d", len(h.entries))
	}
	if !h.entries[0].Timestamp.After(oldTime) {
		t.Error("Timestamp should be updated for duplicate entry")
	}
	if h.entries[0].Title != "Example Updated" {
		t.Errorf("Expected title 'Example Updated', got '%s'", h.entries[0].Title)
	}

	// Test adding different URL
	h.Add("https://example.org", "Example Org")
	if len(h.entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(h.entries))
	}
	if h.currentIndex != 1 {
		t.Errorf("Expected current index 1, got %d", h.currentIndex)
	}
}

func TestHistory_Navigation(t *testing.T) {
	h := NewHistory(10)

	// Add some entries
	h.Add("https://example.com", "Example")
	h.Add("https://example.org", "Example Org")
	h.Add("https://example.net", "Example Net")

	// Test CanGoBack and CanGoForward
	if !h.CanGoBack() {
		t.Error("Should be able to go back")
	}
	if h.CanGoForward() {
		t.Error("Should not be able to go forward")
	}

	// Test Back
	entry, err := h.Back()
	if err != nil {
		t.Errorf("Back failed: %v", err)
	}
	if entry.URL != "https://example.org" {
		t.Errorf("Expected URL 'https://example.org', got '%s'", entry.URL)
	}
	if !h.CanGoForward() {
		t.Error("Should be able to go forward after going back")
	}

	// Test Forward
	entry, err = h.Forward()
	if err != nil {
		t.Errorf("Forward failed: %v", err)
	}
	if entry.URL != "https://example.net" {
		t.Errorf("Expected URL 'https://example.net', got '%s'", entry.URL)
	}

	// Test Current
	current := h.Current()
	if current.URL != "https://example.net" {
		t.Errorf("Expected current URL 'https://example.net', got '%s'", current.URL)
	}

	// Test going back when at the beginning
	h.Clear()
	h.Add("https://example.com", "Example")
	_, err = h.Back()
	if err == nil {
		t.Error("Expected error when going back with no history")
	}
}

func TestHistory_CircularBuffer(t *testing.T) {
	h := NewHistory(3) // Max size of 3

	// Add more entries than max size
	h.Add("https://example1.com", "Example 1")
	h.Add("https://example2.com", "Example 2")
	h.Add("https://example3.com", "Example 3")
	h.Add("https://example4.com", "Example 4") // This should remove the first entry

	if len(h.entries) != 3 {
		t.Errorf("Expected 3 entries, got %d", len(h.entries))
	}
	if h.entries[0].URL != "https://example2.com" {
		t.Errorf("Expected first entry to be 'https://example2.com', got '%s'", h.entries[0].URL)
	}
	if h.entries[2].URL != "https://example4.com" {
		t.Errorf("Expected last entry to be 'https://example4.com', got '%s'", h.entries[2].URL)
	}
}

func TestHistory_Clear(t *testing.T) {
	h := NewHistory(10)
	h.Add("https://example.com", "Example")
	h.Add("https://example.org", "Example Org")

	h.Clear()
	if len(h.entries) != 0 {
		t.Errorf("Expected 0 entries after clear, got %d", len(h.entries))
	}
	if h.currentIndex != -1 {
		t.Errorf("Expected currentIndex -1 after clear, got %d", h.currentIndex)
	}
}

func TestHistory_ScrollPosition(t *testing.T) {
	h := NewHistory(10)
	h.Add("https://example.com", "Example")
	
	// Set scroll position
	h.SetScrollPosition(100)
	
	// Get scroll position
	pos := h.GetScrollPosition()
	if pos != 100 {
		t.Errorf("Expected scroll position 100, got %d", pos)
	}
	
	// Test with empty history
	h.Clear()
	pos = h.GetScrollPosition()
	if pos != 0 {
		t.Errorf("Expected scroll position 0 for empty history, got %d", pos)
	}
}

func TestHistory_SaveLoad(t *testing.T) {
	h := NewHistory(10)
	
	// Add some entries
	h.Add("https://example.com", "Example")
	h.Add("https://example.org", "Example Org")
	
	// Set scroll position for first entry
	h.currentIndex = 0
	h.SetScrollPosition(50)
	
	// Save to disk
	err := h.SaveToDisk()
	if err != nil {
		t.Errorf("SaveToDisk failed: %v", err)
	}
	
	// Create new history and load from disk
	h2 := NewHistory(10)
	err = h2.LoadFromDisk()
	if err != nil {
		t.Errorf("LoadFromDisk failed: %v", err)
	}
	
	// Check that entries were loaded correctly
	if len(h2.entries) != 2 {
		t.Errorf("Expected 2 entries, got %d", len(h2.entries))
	}
	if h2.entries[0].URL != "https://example.com" {
		t.Errorf("Expected first URL 'https://example.com', got '%s'", h2.entries[0].URL)
	}
	if h2.entries[0].ScrollPosition != 50 {
		t.Errorf("Expected scroll position 50, got %d", h2.entries[0].ScrollPosition)
	}
	
	// Clean up test file
	homeDir, _ := os.UserHomeDir()
	historyFile := filepath.Join(homeDir, ".config", "chromedp-browser", "history.json")
	os.Remove(historyFile)
}

func TestHistory_LoadNonExistent(t *testing.T) {
	h := NewHistory(10)
	
	// Make sure the history file doesn't exist
	homeDir, _ := os.UserHomeDir()
	historyFile := filepath.Join(homeDir, ".config", "chromedp-browser", "history.json")
	os.Remove(historyFile)
	
	// Load from disk (should not error even if file doesn't exist)
	err := h.LoadFromDisk()
	if err != nil {
		t.Errorf("LoadFromDisk failed for non-existent file: %v", err)
	}
}