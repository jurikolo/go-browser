// Package browser provides a text-based browser implementation using ChromeDP
package browser

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/chromedp/chromedp"
)

type Browser struct {
	ctx     context.Context
	cancel  context.CancelFunc
	history *History
}

type Link struct {
	Text string
	URL  string
}

// NewBrowser initializes a new browser instance with ChromeDP context
// It sets up a headless browser with a custom user data directory and disables images for faster loading
func NewBrowser() (*Browser, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, fmt.Errorf("failed to get user home directory: %w", err)
	}
	userDataDir := filepath.Join(homeDir, ".config", "chromedp-browser")

	if err := os.MkdirAll(userDataDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create user data directory: %w", err)
	}

	// Set up ChromeDP options
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Headless,
		chromedp.UserDataDir(userDataDir),
		chromedp.Flag("disable-images", true),
		chromedp.NoSandbox,
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-dev-shm-usage", true),
	)

	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	ctx, cancelCtx := chromedp.NewContext(allocCtx)

	if err := chromedp.Run(ctx); err != nil {
		cancel()
		cancelCtx()
		return nil, fmt.Errorf("failed to start browser: %w", err)
	}

	return &Browser{
		ctx:     ctx,
		cancel:  cancelCtx,
		history: NewHistory(100),
	}, nil
}

// Navigates the browser to the specified URL
func (b *Browser) Navigate(url string) error {
	if err := chromedp.Run(b.ctx, chromedp.Navigate(url)); err != nil {
		return err
	}

	title, err := b.GetPageTitle()
	if err != nil {
		title = url // fallback to URL if title unavailable
	}

	b.history.Add(url, title)

	return nil
}

// Extract all visible text from the current page
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

// Retrieve the title of the current page
func (b *Browser) GetPageTitle() (string, error) {
	var title string
	err := chromedp.Run(b.ctx,
		chromedp.Title(&title),
	)
	return title, err
}

// Extract all links from the current page with their text and href attributes
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

// Return the browser's context
func (b *Browser) Context() context.Context {
	return b.ctx
}

// Clean up the browser context and releases resources
func (b *Browser) Close() {
	if b.cancel != nil {
		b.cancel()
	}

	if b.history != nil {
		if err := b.history.SaveToDisk(); err != nil {
			fmt.Printf("Warning: Failed to save history: %v\n", err)
		}
	}
}

// Return the browser's history manager
func (b *Browser) History() *History {
	return b.history
}

// Navigates back in browser history
func (b *Browser) NavigateBack() error {
	entry, err := b.history.Back()
	if err != nil {
		return err
	}

	if err := chromedp.Run(b.ctx, chromedp.Navigate(entry.URL)); err != nil {
		b.history.Forward()
		return err
	}

	return nil
}

// Navigate forward in browser history
func (b *Browser) NavigateForward() error {
	entry, err := b.history.Forward()
	if err != nil {
		return err
	}

	if err := chromedp.Run(b.ctx, chromedp.Navigate(entry.URL)); err != nil {
		return err
	}

	return nil
}
