package browser

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/chromedp/chromedp"
)

// HistoryEntry represents a single entry in the browser history
type HistoryEntry struct {
	URL            string    `json:"url"`
	Title          string    `json:"title"`
	Timestamp      time.Time `json:"timestamp"`
	ScrollPosition int       `json:"scroll_position"`
}

// History manages the browser navigation history
type History struct {
	entries     []HistoryEntry
	currentIndex int
	maxSize     int
}

// NewHistory creates a new history instance with the specified max size
func NewHistory(maxSize int) *History {
	if maxSize <= 0 {
		maxSize = 100 // default size
	}
	return &History{
		entries:     make([]HistoryEntry, 0, maxSize),
		currentIndex: -1,
		maxSize:     maxSize,
	}
}

// Add adds a new entry to the history
func (h *History) Add(url, title string) {
	// Avoid duplicate consecutive entries
	if h.currentIndex >= 0 && h.currentIndex < len(h.entries) {
		if h.entries[h.currentIndex].URL == url {
			// Update the timestamp of the current entry instead of adding a new one
			h.entries[h.currentIndex].Timestamp = time.Now()
			h.entries[h.currentIndex].Title = title
			return
		}
	}

	// Create new entry
	entry := HistoryEntry{
		URL:       url,
		Title:    title,
		Timestamp: time.Now(),
	}

	// If we're not at the end of the history, truncate everything after current index
	if h.currentIndex < len(h.entries)-1 {
		h.entries = h.entries[:h.currentIndex+1]
	}

	// Add the new entry
	h.entries = append(h.entries, entry)
	h.currentIndex++

	// Implement circular buffer for max size
	if len(h.entries) > h.maxSize {
		// Remove the oldest entry
		h.entries = h.entries[1:]
		h.currentIndex--
	}
}

// Back navigates backward in history
func (h *History) Back() (HistoryEntry, error) {
	if !h.CanGoBack() {
		return HistoryEntry{}, fmt.Errorf("cannot go back, no previous entries")
	}
	h.currentIndex--
	return h.entries[h.currentIndex], nil
}

// Forward navigates forward in history
func (h *History) Forward() (HistoryEntry, error) {
	if !h.CanGoForward() {
		return HistoryEntry{}, fmt.Errorf("cannot go forward, no next entries")
	}
	h.currentIndex++
	return h.entries[h.currentIndex], nil
}

// Current returns the current history entry
func (h *History) Current() HistoryEntry {
	if h.currentIndex >= 0 && h.currentIndex < len(h.entries) {
		return h.entries[h.currentIndex]
	}
	return HistoryEntry{}
}

// CanGoBack returns true if we can navigate back in history
func (h *History) CanGoBack() bool {
	return h.currentIndex > 0
}

// CanGoForward returns true if we can navigate forward in history
func (h *History) CanGoForward() bool {
	return h.currentIndex < len(h.entries)-1
}

// Clear clears all history entries
func (h *History) Clear() {
	h.entries = h.entries[:0]
	h.currentIndex = -1
}

// SetScrollPosition updates the scroll position for the current entry
func (h *History) SetScrollPosition(position int) {
	if h.currentIndex >= 0 && h.currentIndex < len(h.entries) {
		h.entries[h.currentIndex].ScrollPosition = position
	}
}

// GetScrollPosition returns the scroll position for the current entry
func (h *History) GetScrollPosition() int {
	if h.currentIndex >= 0 && h.currentIndex < len(h.entries) {
		return h.entries[h.currentIndex].ScrollPosition
	}
	return 0
}

// SaveToDisk persists the history to disk
func (h *History) SaveToDisk() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}
	
	configDir := filepath.Join(homeDir, ".config", "chromedp-browser")
	if err := os.MkdirAll(configDir, 0755); err != nil {
		return fmt.Errorf("failed to create config directory: %w", err)
	}
	
	historyFile := filepath.Join(configDir, "history.json")
	
	// Create a copy of entries to avoid concurrency issues
	entries := make([]HistoryEntry, len(h.entries))
	copy(entries, h.entries)
	
	data, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal history: %w", err)
	}
	
	if err := os.WriteFile(historyFile, data, 0644); err != nil {
		return fmt.Errorf("failed to write history file: %w", err)
	}
	
	return nil
}

// LoadFromDisk loads history from disk
func (h *History) LoadFromDisk() error {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}
	
	historyFile := filepath.Join(homeDir, ".config", "chromedp-browser", "history.json")
	
	data, err := os.ReadFile(historyFile)
	if err != nil {
		if os.IsNotExist(err) {
			// File doesn't exist, which is fine - start with empty history
			return nil
		}
		return fmt.Errorf("failed to read history file: %w", err)
	}
	
	var entries []HistoryEntry
	if err := json.Unmarshal(data, &entries); err != nil {
		return fmt.Errorf("failed to unmarshal history: %w", err)
	}
	
	// Apply circular buffer logic if needed
	if len(entries) > h.maxSize {
		entries = entries[len(entries)-h.maxSize:]
	}
	
	h.entries = entries
	h.currentIndex = len(entries) - 1
	
	return nil
}

// NavigateBack navigates back using ChromeDP
func (h *History) NavigateBack(ctx context.Context) error {
	if !h.CanGoBack() {
		return fmt.Errorf("cannot go back, no previous entries")
	}
	
	// Navigate using ChromeDP
	if err := chromedp.Run(ctx, chromedp.NavigateBack()); err != nil {
		return fmt.Errorf("failed to navigate back: %w", err)
	}
	
	// Move our internal pointer back
	h.currentIndex--
	return nil
}

// NavigateForward navigates forward using ChromeDP
func (h *History) NavigateForward(ctx context.Context) error {
	if !h.CanGoForward() {
		return fmt.Errorf("cannot go forward, no next entries")
	}
	
	// Navigate using ChromeDP
	if err := chromedp.Run(ctx, chromedp.NavigateForward()); err != nil {
		return fmt.Errorf("failed to navigate forward: %w", err)
	}
	
	// Move our internal pointer forward
	h.currentIndex++
	return nil
}