// Package browser provides a text-based browser implementation using ChromeDP
package browser

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/chromedp/chromedp"
)

// Browser wraps a ChromeDP context to provide a simplified browser interface
type Browser struct {
	ctx    context.Context
	cancel context.CancelFunc
}

// NewBrowser initializes a new browser instance with ChromeDP context
// It sets up a headless browser with a custom user data directory and disables images for faster loading
func NewBrowser() (*Browser, error) {
	// Create a custom user data directory
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}
	userDataDir := filepath.Join(homeDir, ".config", "chromedp-browser")

	// Create the user data directory if it doesn't exist
	if err := os.MkdirAll(userDataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create user data directory: %w", err)
	}

	// Set up ChromeDP options
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Headless,
		chromedp.UserDataDir(userDataDir),
		// Disable images for faster loading
		chromedp.Flag("disable-images", true),
		// Allow ChromeDP to download browser if not available
		chromedp.NoSandbox,
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	)

	// Create the allocator context
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)

	// Create the browser context
	ctx, cancelCtx := chromedp.NewContext(allocCtx)

	// Ensure the browser is started
	if err := chromedp.Run(ctx); err != nil {
		cancel()
		cancelCtx()
		return nil, fmt.Errorf("failed to start browser: %w", err)
	}

	return &Browser{
		ctx:    ctx,
		cancel: cancelCtx,
	}, nil
}

// Navigate navigates the browser to the specified URL
func (b *Browser) Navigate(url string) error {
	return chromedp.Run(b.ctx, chromedp.Navigate(url))
}

// GetPageText extracts all visible text from the current page
func (b *Browser) GetPageText() (string, error) {
	var text string
	err := chromedp.Run(b.ctx,
		chromedp.Evaluate(`
			(function() {
				var walker = document.createTreeWalker(
					document.body,
					NodeFilter.SHOW_TEXT,
					null,
					false
				);
				var textNodes = [];
				var node;
				while (node = walker.nextNode()) {
					if (node.parentElement.tagName !== 'SCRIPT' &&
					    node.parentElement.tagName !== 'STYLE') {
						textNodes.push(node.nodeValue);
					}
				}
				return textNodes.join(' ').replace(/\s+/g, ' ').trim();
			})()
		`, &text),
	)
	return text, err
}

// GetPageTitle retrieves the title of the current page
func (b *Browser) GetPageTitle() (string, error) {
	var title string
	err := chromedp.Run(b.ctx,
		chromedp.Title(&title),
	)
	return title, err
}

// Link represents a hyperlink with its text and href
type Link struct {
	Text string
	URL  string
}

// GetLinks extracts all links from the current page with their text and href attributes
func (b *Browser) GetLinks() ([]Link, error) {
	var links []Link
	err := chromedp.Run(b.ctx,
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
		`, &links),
	)
	return links, err
}

// Context returns the browser's context
func (b *Browser) Context() context.Context {
	return b.ctx
}

// Close cleans up the browser context and releases resources
func (b *Browser) Close() {
	if b.cancel != nil {
		b.cancel()
	}
}
