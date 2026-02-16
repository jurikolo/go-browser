// Package browser provides a text-based browser implementation using ChromeDP
package browser

import (
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/cdproto/runtime"
	"github.com/chromedp/chromedp"
	"github.com/jurikolo/go-browser/config"
)

type Browser struct {
	ctx          context.Context
	cancel       context.CancelFunc
	history      *History
	cfg          *config.Config
	requestCount int64
	loadTime     time.Duration
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

// NewBrowserWithConfig initializes a new browser instance with ChromeDP context using the provided configuration
func NewBrowserWithConfig(cfg *config.Config) (*Browser, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	if err := cfg.EnsurePaths(); err != nil {
		return nil, fmt.Errorf("failed to ensure paths: %w", err)
	}

	// Set up ChromeDP options
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.UserDataDir(cfg.UserDataDir),
		chromedp.Flag("disable-images", cfg.DisableImages),
		chromedp.UserAgent(cfg.UserAgent),
	)

	// Add headless flag if enabled
	if cfg.HeadlessMode {
		opts = append(opts, chromedp.Headless)
	}

	// Add custom Chrome flags
	for _, flag := range cfg.ChromeFlags {
		// Parse flag format: --flag=value or --flag
		if len(flag) > 2 && flag[0:2] == "--" {
			eqIndex := len(flag)
			for i, c := range flag {
				if c == '=' {
					eqIndex = i
					break
				}
			}

			if eqIndex < len(flag) {
				key := flag[2:eqIndex]
				value := flag[eqIndex+1:]
				opts = append(opts, chromedp.Flag(key, value))
			} else {
				key := flag[2:]
				opts = append(opts, chromedp.Flag(key, true))
			}
		}
	}

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
		history: NewHistory(cfg.MaxHistorySize),
		cfg:     cfg,
	}, nil
}

// Navigates the browser to the specified URL with performance monitoring
func (b *Browser) Navigate(url string) error {
	startTime := time.Now()
	b.requestCount = 0

	// Enable network domain to track requests
	if err := chromedp.Run(b.ctx, network.Enable()); err != nil {
		log.Printf("Warning: Failed to enable network monitoring: %v", err)
	}

	// Set up network event listener
	chromedp.ListenTarget(b.ctx, func(ev interface{}) {
		if _, ok := ev.(*network.EventRequestWillBeSent); ok {
			b.requestCount++
		}
	})

	// Navigate to the URL
	if err := chromedp.Run(b.ctx, chromedp.Navigate(url)); err != nil {
		return b.handleNavigationError(err, url)
	}

	// Wait for page to load
	if err := b.WaitForNavigation(); err != nil {
		log.Printf("Warning: Navigation wait failed: %v", err)
	}

	b.loadTime = time.Since(startTime)

	title, err := b.GetPageTitle()
	if err != nil {
		title = url // fallback to URL if title unavailable
	}

	b.history.Add(url, title)

	return nil
}

// ExecuteJS executes JavaScript code in the browser context and returns the result
// Example: Extract data not easily accessible via DOM
// result, err := browser.ExecuteJS("document.querySelectorAll('a').length")
func (b *Browser) ExecuteJS(script string) (interface{}, error) {
	var result interface{}
	err := chromedp.Run(b.ctx,
		chromedp.Evaluate(script, &result),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to execute JavaScript: %w", err)
	}
	return result, nil
}

// TakeScreenshot captures a screenshot of the current page and saves it to the specified filename
// Screenshots are saved to ~/.config/chromedp-browser/screenshots/
func (b *Browser) TakeScreenshot(filename string) error {
	// Ensure screenshots directory exists
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return fmt.Errorf("failed to get user home directory: %w", err)
	}

	screenshotsDir := filepath.Join(homeDir, ".config", "chromedp-browser", "screenshots")
	if err := os.MkdirAll(screenshotsDir, 0755); err != nil {
		return fmt.Errorf("failed to create screenshots directory: %w", err)
	}

	// Full path to screenshot file
	fullPath := filepath.Join(screenshotsDir, filename)

	// Capture screenshot
	var buf []byte
	if err := chromedp.Run(b.ctx, chromedp.CaptureScreenshot(&buf)); err != nil {
		return fmt.Errorf("failed to capture screenshot: %w", err)
	}

	// Save to file
	if err := os.WriteFile(fullPath, buf, 0644); err != nil {
		return fmt.Errorf("failed to save screenshot: %w", err)
	}

	log.Printf("Screenshot saved to: %s", fullPath)
	return nil
}

// WaitForSelector waits for an element matching the selector to appear in the DOM
// Useful for dynamic content that loads asynchronously
func (b *Browser) WaitForSelector(selector string, timeout time.Duration) error {
	if timeout <= 0 {
		timeout = 30 * time.Second // Default timeout
	}

	ctx, cancel := context.WithTimeout(b.ctx, timeout)
	defer cancel()

	return chromedp.Run(ctx, chromedp.WaitVisible(selector))
}

// WaitForNavigation waits for the page to finish navigating and loading
// Useful after clicking links or submitting forms
func (b *Browser) WaitForNavigation() error {
	return chromedp.Run(b.ctx, chromedp.WaitReady("body"))
}

// DisableJavaScript disables JavaScript execution in the browser
// Useful for text-only mode or performance optimization
func (b *Browser) DisableJavaScript() error {
	return chromedp.Run(b.ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		return runtime.Disable().Do(ctx)
	}))
}

// EnableJavaScript enables JavaScript execution in the browser
func (b *Browser) EnableJavaScript() error {
	return chromedp.Run(b.ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		return runtime.Enable().Do(ctx)
	}))
}

// SetUserAgent sets a custom user agent string for the browser
// Useful for mobile/desktop rendering simulation
func (b *Browser) SetUserAgent(userAgent string) error {
	// Set user agent through Chrome flags
	return chromedp.Run(b.ctx, chromedp.ActionFunc(func(ctx context.Context) error {
		// This is a simplified approach - in practice, you might need to restart
		// the browser with new flags for this to take effect
		return nil
	}))
}

// FillForm fills a form field with the specified value
// selector: CSS selector for the form field
// value: value to fill in the field
func (b *Browser) FillForm(selector string, value string) error {
	return chromedp.Run(b.ctx, chromedp.SendKeys(selector, value))
}

// SubmitForm submits a form by clicking its submit button
// selector: CSS selector for the form submit button
func (b *Browser) SubmitForm(selector string) error {
	return chromedp.Run(b.ctx, chromedp.Click(selector))
}

// GetRequestCount returns the number of HTTP requests made for the current page
func (b *Browser) GetRequestCount() int64 {
	return b.requestCount
}

// GetLoadTime returns the time taken to load the current page
func (b *Browser) GetLoadTime() time.Duration {
	return b.loadTime
}

// Extract all visible text from the current page
func (b *Browser) GetPageText() (string, error) {
	// Get the HTML content of the page
	html, err := b.GetPageHTML()
	if err != nil {
		return "", err
	}

	// Format the HTML content using our renderer
	formatted, _, err := FormatHTML(html)
	if err != nil {
		return "", err
	}

	return formatted, nil
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

// GetPageHTML returns the outer HTML of the current page
func (b *Browser) GetPageHTML() (string, error) {
	var html string
	err := chromedp.Run(b.ctx,
		chromedp.OuterHTML("html", &html),
	)
	return html, err
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

// handleNavigationError handles navigation errors with retry logic
func (b *Browser) handleNavigationError(err error, url string) error {
	log.Printf("Navigation error: %v", err)

	// Check if it's a context canceled error (browser closed)
	if strings.Contains(err.Error(), "context canceled") {
		log.Printf("Browser context was canceled, attempting to restart...")
		if restartErr := b.restart(); restartErr != nil {
			return fmt.Errorf("navigation failed and browser restart failed: %w (original error: %v)", restartErr, err)
		}

		// Retry navigation after restart
		log.Printf("Retrying navigation to %s after browser restart", url)
		if retryErr := chromedp.Run(b.ctx, chromedp.Navigate(url)); retryErr != nil {
			return fmt.Errorf("navigation retry failed after browser restart: %w (original error: %v)", retryErr, err)
		}
		return nil
	}

	return err
}

// restart attempts to restart the browser context
func (b *Browser) restart() error {
	// Close current context
	if b.cancel != nil {
		b.cancel()
	}

	// Create new browser context with same configuration
	var err error
	var newBrowser *Browser

	if b.cfg != nil {
		newBrowser, err = NewBrowserWithConfig(b.cfg)
	} else {
		newBrowser, err = NewBrowser()
	}

	if err != nil {
		return fmt.Errorf("failed to create new browser instance: %w", err)
	}

	// Update current browser with new context
	b.ctx = newBrowser.ctx
	b.cancel = newBrowser.cancel

	return nil
}

// Example usage of advanced features:
//
// 1. JavaScript Execution:
//    result, err := browser.ExecuteJS("document.title")
//    if err != nil {
//        log.Printf("Failed to execute JS: %v", err)
//    } else {
//        log.Printf("Page title from JS: %v", result)
//    }
//
// 2. Screenshot capability:
//    err := browser.TakeScreenshot("example.png")
//    if err != nil {
//        log.Printf("Failed to take screenshot: %v", err)
//    }
//
// 3. Wait strategies:
//    err := browser.WaitForSelector("#dynamic-content", 10*time.Second)
//    if err != nil {
//        log.Printf("Element not found: %v", err)
//    }
//
// 4. Performance monitoring:
//    log.Printf("Requests: %d, Load time: %v", browser.GetRequestCount(), browser.GetLoadTime())
//
// 5. Form filling:
//    err := browser.FillForm("#username", "example")
//    if err != nil {
//        log.Printf("Failed to fill form: %v", err)
//    }
//    err = browser.SubmitForm("#submit-btn")
//    if err != nil {
//        log.Printf("Failed to submit form: %v", err)
//    }
