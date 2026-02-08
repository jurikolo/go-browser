# Go Text Browser with ChromeDP

## Project Overview
A single-tab text-based web browser for Linux built in Go, using ChromeDP (headless Chrome) with persistent cookie authentication support.

## Prerequisites
- Go 1.21 or later installed
- Linux development environment (Ubuntu/Debian recommended)
- Chrome/Chromium browser installed (for chromedp)
- Basic understanding of Go syntax and package management

---

## Technology Stack Updates

### Browser Engine: ChromeDP
**Chosen: github.com/chromedp/chromedp**

**Advantages:**
- ✅ Native Go library (no FFI complexity)
- ✅ Full Chrome rendering engine
- ✅ Built-in cookie management
- ✅ JavaScript execution support
- ✅ Mature, well-documented library
- ✅ Active community and maintenance

**Installation:**
```bash
go get -u github.com/chromedp/chromedp
```

### Terminal UI Library Options

Since you don't like tcell, here are three excellent alternatives:

#### Option 1: **Bubble Tea (RECOMMENDED)** ⭐
**Package:** github.com/charmbracelet/bubbletea

**Why it's great:**
- Modern, elegant API based on The Elm Architecture
- Part of the Charm ecosystem (includes Lipgloss for styling, Bubbles for components)
- Excellent for building interactive TUIs
- Great documentation and examples
- Active development

**Pros:**
- Clean, functional programming style
- Built-in components for lists, text inputs, viewports
- Easy to test and reason about
- Beautiful default styling with Lipgloss

**Cons:**
- Different paradigm (model-update-view) requires learning
- Slightly more complex initial setup

**Installation:**
```bash
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss
go get github.com/charmbracelet/bubbles
```

---

#### Option 2: **Tview**
**Package:** github.com/rivo/tview

**Why it's great:**
- High-level components (forms, tables, lists, text views)
- Built on top of tcell but with easier API
- Lots of pre-built widgets
- Excellent for quickly building complex UIs

**Pros:**
- Fast to build complex UIs
- Many examples and good documentation
- Less boilerplate than tcell
- Built-in layouts and widgets

**Cons:**
- Still uses tcell under the hood
- Less control over low-level details
- Heavier than alternatives

**Installation:**
```bash
go get github.com/rivo/tview
```

---

#### Option 3: **Gocui**
**Package:** github.com/jroimartin/gocui

**Why it's great:**
- Minimalist, straightforward API
- View-based architecture
- Lightweight and simple
- Good for simpler UIs

**Pros:**
- Very simple and easy to learn
- Minimal dependencies
- Good for basic TUI needs

**Cons:**
- Less actively maintained (still stable)
- Fewer built-in components
- Less sophisticated styling

**Installation:**
```bash
go get github.com/jroimartin/gocui
```

---

### **Recommendation: Bubble Tea**

For this project, I recommend **Bubble Tea** because:
1. Modern, well-maintained library
2. Great for interactive applications
3. Excellent viewport component for scrolling content
4. Clean separation of concerns
5. Growing ecosystem (Charm.sh tools)

---

## Updated Development Phases

### Phase 1: Environment Setup (Week 1)
**Objectives:**
- Set up Go development environment with ChromeDP
- Install Chrome/Chromium
- Set up terminal UI library

**Tasks:**
```bash
# Install dependencies
go get github.com/chromedp/chromedp
go get github.com/charmbracelet/bubbletea
go get github.com/charmbracelet/lipgloss
go get github.com/charmbracelet/bubbles

# Install Chrome (if not present)
sudo apt-get install chromium-browser  # Ubuntu/Debian
```

---

### Phase 2: Core Architecture Design (Week 1-2)
**Objectives:**
- Design application architecture with ChromeDP
- Plan component interactions
- Define data structures

**Updated Components:**
```
├── main.go                 # Entry point
├── browser/
│   ├── chromedp.go        # ChromeDP wrapper and page operations
│   ├── renderer.go        # HTML to text conversion
│   └── navigation.go      # URL handling & navigation
├── cookie/
│   ├── manager.go         # Cookie sync with ChromeDP
│   └── persistence.go     # File-based cookie persistence
├── ui/
│   ├── bubbletea.go       # Bubble Tea model and UI
│   ├── viewport.go        # Scrollable content viewport
│   └── styles.go          # Lipgloss styling
└── config/
    └── config.go          # Application configuration
```

---

### Phase 3: ChromeDP Integration (Week 2-3)
**Objectives:**
- Set up ChromeDP context
- Implement page navigation
- Extract page content and links
- Handle JavaScript-rendered content

**Key ChromeDP Features to Use:**
- `chromedp.Navigate()` - Load URLs
- `chromedp.Text()` - Extract text content
- `chromedp.OuterHTML()` - Get full HTML
- `chromedp.Cookies()` - Get cookies from browser
- `chromedp.SetCookies()` - Set cookies in browser
- `chromedp.WaitVisible()` - Wait for content to load

---

### Phase 4: Cookie Management with ChromeDP (Week 3)
**Objectives:**
- Sync cookies between ChromeDP and persistent storage
- Load cookies on browser startup
- Save cookies on page navigation and shutdown

**ChromeDP Cookie Workflow:**
1. On startup: Load cookies from disk → Set in ChromeDP context
2. After navigation: Get cookies from ChromeDP → Save to disk
3. On shutdown: Get all cookies → Save to disk

**Benefits of ChromeDP for Cookies:**
- ChromeDP handles cookie parsing automatically
- Respects all cookie attributes (HttpOnly, Secure, SameSite)
- Automatic cookie expiration
- No manual HTTP header parsing needed

---

### Phase 5: Bubble Tea UI Implementation (Week 4-5)
**Objectives:**
- Create interactive terminal UI
- Implement keyboard navigation
- Display rendered content with scrolling
- Add status bar and URL input

**Bubble Tea Architecture:**
```go
Model (State) → Update (Event Handler) → View (Renderer)
```

**UI Components:**
- URL input (text input component)
- Content viewport (bubbles/viewport)
- Status bar (showing URL, loading state, cookie count)
- Help footer (keyboard shortcuts)

---

## Updated AI Development Prompts

### Prompt 3A: ChromeDP Setup and Basic Navigation
```
I'm building a text-based browser using ChromeDP in Go. Please create:

1. File: browser/chromedp.go
   - Initialize a chromedp context with options:
     * Headless mode enabled
     * User data directory: ~/.config/chromedp-browser/
     * Disable images for faster loading (optional)
   - Create a Browser struct that wraps the chromedp context
   - Implement methods:
     * NewBrowser() - Initialize browser with context
     * Navigate(url string) - Navigate to URL
     * GetPageText() - Extract all visible text from page
     * GetPageTitle() - Get page title
     * GetLinks() - Extract all links with their text and href
     * Close() - Cleanup and close browser context
   - Handle timeouts (30 seconds default)
   - Include proper error handling

2. Example usage in main.go showing:
   - Browser initialization
   - Navigation to a URL
   - Content extraction
   - Proper cleanup with defer

Please provide complete, well-documented code with comments.
```

---

### Prompt 3B: ChromeDP Cookie Integration
```
Create cookie management that integrates with ChromeDP:

1. File: cookie/chromedp_manager.go
   - CookieManager struct with chromedp integration
   - Methods:
     * LoadCookiesIntoChromedp(ctx context.Context) - Load saved cookies into browser
     * ExtractCookiesFromChromedp(ctx context.Context) - Get cookies from browser
     * SaveCookiesToDisk() - Persist cookies to JSON file
     * LoadCookiesFromDisk() - Load cookies from JSON file
     * GetCookiesForDomain(domain string) - Filter cookies by domain

2. Cookie storage format:
   - File location: ~/.config/chromedp-browser/cookies.json
   - Store as array of cookie objects with:
     * Name, Value, Domain, Path
     * Expires, HttpOnly, Secure, SameSite
   - Implement AES encryption for cookie values (optional but recommended)

3. Integration workflow:
   - On browser startup: Load cookies from disk → Set in ChromeDP
   - After each navigation: Extract cookies from ChromeDP → Save to disk
   - Handle cookie expiration (filter out expired cookies on load)

4. Use chromedp.Cookies() and chromedp.SetCookies() functions
5. Include comprehensive error handling

Provide complete code with examples of the workflow.
```

---

### Prompt 4: HTML to Text Conversion for ChromeDP
```
Create a sophisticated HTML-to-text renderer for ChromeDP output:

File: browser/renderer.go

The renderer should:
1. Take HTML string from ChromeDP (via OuterHTML)
2. Parse and convert to readable plain text preserving structure
3. Handle these elements specially:
   - Headings (h1-h6): Show with underlines (=== or ---)
   - Paragraphs: Add blank line between paragraphs
   - Links: Show as [text](url) or [1] text with footnote list
   - Lists (ul, ol): Use bullets/numbers with indentation
   - Code blocks: Preserve formatting with indentation
   - Tables: Convert to simple text table (use library if needed)
   - Images: Show as [Image: alt text]

4. Number all links sequentially for easy navigation
5. Return both:
   - Formatted text for display
   - Array of Link objects with number, text, and URL

6. Use golang.org/x/net/html for parsing
7. Strip JavaScript and style tags completely
8. Handle nested elements correctly

Include:
- Link struct definition
- ExtractLinks() helper function
- FormatHTML() main function
- Unit tests for common HTML patterns

Provide production-ready code with comments.
```

### Implementation Status

✅ **Completed**: The HTML renderer has been implemented with full support for:
- Headings with underlines
- Paragraph spacing
- Numbered links with footnotes
- Lists with indentation
- Code blocks with indentation
- Image placeholders with alt text
- Complete stripping of script and style tags
- Nested element handling
- Comprehensive unit tests

**Example Usage:**
```go
htmlContent := `
<html>
  <body>
    <h1>Welcome</h1>
    <p>Visit <a href="https://example.com">Example</a></p>
    <ul>
      <li>Item 1</li>
      <li>Item 2</li>
    </ul>
  </body>
</html>`

formattedText, links, err := browser.FormatHTML(htmlContent)
if err != nil {
  log.Fatal(err)
}

fmt.Println(formattedText)
// Output:
// Welcome
// =======
// Visit [1] Example
//
//   Item 1
//   Item 2
//
// Links:
// [1] https://example.com

for _, link := range links {
  fmt.Printf("[%d] %s -> %s\n", link.Number, link.Text, link.URL)
}
```

---

### Prompt 5A: Bubble Tea UI Setup
```
Create a Bubble Tea terminal UI for the browser:

File: ui/bubbletea.go

1. Define the Model struct with fields:
   - viewport (bubbles/viewport.Viewport) for scrollable content
   - urlInput (bubbles/textinput.TextInput) for URL entry
   - content (string) - current page content
   - links ([]Link) - extracted links from page
   - currentURL (string)
   - status (string) - loading, ready, error
   - mode (string) - "normal", "url-input"
   - width, height (int) - terminal dimensions

2. Implement required Bubble Tea methods:
   - Init() tea.Cmd - Initialize the model
   - Update(msg tea.Msg) (tea.Model, tea.Cmd) - Handle events
   - View() string - Render the UI

3. Handle keyboard input:
   - 'g': Switch to URL input mode
   - Enter: Navigate to URL (when in URL input mode)
   - Esc: Exit URL input mode
   - Arrow up/down: Scroll content
   - 'q': Quit application
   - '1-9': Navigate to numbered link
   - 'r': Reload current page
   - 'b': Go back in history
   - 'c': Show cookies for current site

4. Layout the UI with sections:
   - Top: URL bar (always visible)
   - Middle: Scrollable content viewport (main area)
   - Bottom: Status bar + keyboard shortcuts

5. Use Lipgloss for styling (colors, borders, padding)

Please provide complete implementation with:
- All required imports
- Proper initialization
- Event handling logic
- View rendering with Lipgloss styles
```

---

### Prompt 5B: Bubble Tea Styling and Layout
```
Create beautiful styling for the Bubble Tea browser UI:

File: ui/styles.go

Define Lipgloss styles for:

1. URL Bar:
   - Border with rounded corners
   - Distinct background color
   - Padding
   - Show current URL and input mode

2. Content Viewport:
   - Scrollable area
   - Padding for readability
   - Line wrapping
   - Syntax highlighting for code blocks (if possible)

3. Status Bar:
   - Different color scheme
   - Show: loading indicator, cookie count, scroll position
   - Aligned layout (left, center, right sections)

4. Help Footer:
   - Compact keyboard shortcuts
   - Faded colors (less prominent)

5. Color Scheme:
   - Use a coherent color palette
   - Support both light and dark terminals
   - Highlight active elements

6. Link Styling:
   - Make numbered links stand out
   - Use different color for visited links (optional)

Example styles to create:
- urlBarStyle
- contentStyle  
- statusBarStyle
- helpStyle
- linkStyle
- headingStyle

Include:
- Responsive design (adapts to terminal width)
- Clear visual hierarchy
- Accessibility considerations

Provide complete styles.go with examples.
```

---

### Prompt 6: Integration - Connecting Browser Engine and UI
```
Update main.go to integrate ChromeDP browser with Bubble Tea UI:

1. Application initialization:
   - Load configuration
   - Initialize ChromeDP browser
   - Load cookies from disk into browser
   - Create Bubble Tea model
   - Start Bubble Tea program

2. Create a channel-based architecture:
   - Browser operations run in goroutines
   - Results sent through channels to UI
   - UI remains responsive during page loads

3. Navigation workflow:
   - User enters URL in Bubble Tea UI
   - Trigger ChromeDP navigation in goroutine
   - Show loading indicator in UI
   - Extract content and links when loaded
   - Update viewport with new content
   - Save cookies to disk

4. Implement these integration functions:
   - navigateToURL(url string) - Wrapper for ChromeDP navigation
   - updateUIWithPage() - Extract and format content for display
   - handleLinkClick(linkNumber int) - Navigate to numbered link
   - syncCookies() - Save cookies after navigation

5. Graceful shutdown:
   - Save cookies on exit
   - Close ChromeDP context
   - Cleanup terminal state

6. Error handling:
   - Network errors
   - Invalid URLs
   - ChromeDP context errors
   - Display user-friendly error messages in status bar

Provide complete main.go that:
- Uses proper goroutine patterns
- Handles context cancellation
- Includes comprehensive error handling
- Has clear separation between browser and UI layers
```

---

### Prompt 7: Navigation History and Back/Forward
```
Create a navigation history system that works with ChromeDP:

File: browser/history.go

1. History struct with:
   - entries []HistoryEntry
   - currentIndex int
   - maxSize int (default 100)

2. HistoryEntry struct:
   - URL string
   - Title string
   - Timestamp time.Time
   - ScrollPosition int (for restoring scroll)

3. Methods:
   - Add(url, title string) - Add to history
   - Back() (HistoryEntry, error) - Go back
   - Forward() (HistoryEntry, error) - Go forward
   - Current() HistoryEntry - Get current entry
   - CanGoBack() bool
   - CanGoForward() bool
   - Clear() - Clear all history
   - SaveToDisk() - Persist to ~/.config/chromedp-browser/history.json
   - LoadFromDisk() - Load history from file

4. Integration with ChromeDP:
   - When navigating forward/back, use ChromeDP's native back/forward
   - Sync internal history with browser history
   - Restore scroll position when possible

5. Avoid duplicate consecutive entries
6. Implement circular buffer for max size

Include proper error handling and tests.
```

---

### Prompt 8: Configuration Management
```
Create a configuration system for the browser:

File: config/config.go

1. Config struct with:
   - UserAgent string
   - Homepage string
   - CookieStorePath string
   - HistoryStorePath string
   - UserDataDir string (for ChromeDP)
   - HeadlessMode bool
   - DisableImages bool (for faster loading)
   - NavigationTimeout time.Duration
   - MaxHistorySize int
   - ChromeFlags []string (custom Chrome flags)

2. Load configuration from:
   - Default values (constants)
   - YAML file: ~/.config/chromedp-browser/config.yaml
   - Environment variables (override YAML)
   - Command-line flags (final override)

3. Use libraries:
   - github.com/spf13/viper for config loading
   - github.com/spf13/pflag for CLI flags

4. CLI flags to support:
   - --url <url> : Start with specific URL
   - --config <path> : Custom config file
   - --headless : Force headless mode
   - --user-data-dir <path> : Custom Chrome profile location

5. Validation:
   - Ensure paths exist or create them
   - Validate URL format for homepage
   - Reasonable timeout values

6. Method to save current config back to file

Include example config.yaml with comprehensive comments.
```

---

### Prompt 9: ChromeDP Advanced Features
```
Add advanced ChromeDP features to browser/chromedp.go:

1. JavaScript Execution:
   - ExecuteJS(script string) (interface{}, error)
   - Example: Extract data not easily accessible via DOM

2. Screenshot capability (for debugging):
   - TakeScreenshot(filename string) error
   - Save to ~/.config/chromedp-browser/screenshots/

3. Wait strategies:
   - WaitForSelector(selector string, timeout time.Duration)
   - WaitForNavigation()
   - Useful for dynamic content

4. Performance optimizations:
   - Disable images: chromedp.Flag("disable-images", true)
   - Disable JavaScript (optional, for text-only mode)
   - Custom user agent for mobile/desktop rendering

5. Network monitoring:
   - Track request count
   - Measure page load time
   - Display in status bar

6. Form filling capability (for future authentication features):
   - FillForm(selector string, value string)
   - SubmitForm(selector string)

7. Error recovery:
   - Auto-restart ChromeDP on crashes
   - Retry failed navigations
   - Handle common errors gracefully

Include comprehensive error handling and logging.
Provide examples of each feature.
```

---

### Prompt 10: Link Navigation and Selection
```
Enhance link handling for better user experience:

File: browser/links.go

1. Link struct:
   - Number int (for keyboard selection)
   - Text string (display text)
   - URL string (absolute URL)
   - Type string ("internal", "external", "anchor")

2. ExtractLinks function:
   - Parse HTML from ChromeDP
   - Find all <a> tags
   - Resolve relative URLs to absolute
   - Filter out javascript: and mailto: (or handle separately)
   - Assign sequential numbers (1, 2, 3, ...)
   - Group by type (internal vs external)

3. FormatLinksForDisplay:
   - Create a section showing all numbered links
   - Format as: [1] Link Text (URL)
   - Optionally truncate long URLs
   - Add visual separators for link groups

4. NavigateToLink(number int):
   - Find link by number
   - Navigate using ChromeDP
   - Handle external links (open vs. show warning)

5. SearchLinks(query string):
   - Filter links by text or URL matching query
   - Useful for pages with many links

6. Special link handling:
   - Anchor links (#section): Smooth scroll to element
   - File downloads: Show download prompt
   - External domains: Optional confirmation

Integration with UI:
- Display links below main content
- Highlight selected link
- Show link preview in status bar on hover

Provide complete implementation with tests.
```

---

### Prompt 11: Logging and Debugging
```
Implement comprehensive logging system:

File: browser/logger.go

1. Logger setup:
   - Log to file: ~/.config/chromedp-browser/browser.log
   - Rotate logs (max 10MB, keep 5 files)
   - Log levels: DEBUG, INFO, WARN, ERROR
   - Include timestamp, level, and source location

2. Use structured logging:
   - Library: github.com/sirupsen/logrus or zerolog
   - JSON format for easy parsing
   - Include context (URL, action, duration)

3. What to log:
   - Page navigation (URL, time taken)
   - Cookie operations (saved, loaded, count)
   - ChromeDP errors and retries
   - User actions (link clicks, searches)
   - Performance metrics (page load time, size)

4. Debug mode:
   - Enable with --debug flag
   - Log ChromeDP protocol messages
   - Log HTML content (truncated)
   - Verbose error messages

5. Log viewer helper:
   - Function to tail logs in real-time
   - Filter logs by level
   - Search logs by keyword

6. Privacy considerations:
   - Don't log sensitive data (passwords, tokens)
   - Option to disable logging entirely
   - Clear logs command

Integration:
- Log viewer accessible via 'l' key in UI
- Show recent errors in status bar
- Performance stats in debug panel

Provide complete logger implementation with examples.
```

---

### Prompt 12: Testing Strategy
```
Create comprehensive test suite for the browser:

1. Unit Tests:

File: browser/chromedp_test.go
- Test URL validation
- Test link extraction from sample HTML
- Test cookie serialization/deserialization
- Mock ChromeDP operations

File: cookie/manager_test.go
- Test cookie CRUD operations
- Test expiration handling
- Test domain/path matching
- Test persistence (save/load)

File: browser/renderer_test.go
- Test HTML to text conversion
- Test link numbering
- Test formatting preservation
- Use golden files for expected outputs

2. Integration Tests:

File: integration_test.go
- Test full navigation workflow
- Test cookie persistence across restarts
- Test history navigation (back/forward)
- Use testify/suite for setup/teardown

3. UI Tests:

File: ui/bubbletea_test.go
- Test keyboard input handling
- Test model updates
- Test view rendering
- Use bubble tea's testing utilities

4. Test Utilities:
- Mock HTTP server for testing
- Sample HTML files for parsing tests
- Helper functions for creating test browser instances

5. Coverage:
- Aim for >70% code coverage
- Focus on critical paths (navigation, cookies, UI)
- Use `go test -cover` and `go tool cover`

6. Benchmark tests:
- Page load performance
- HTML parsing speed
- UI rendering performance

7. Testing dependencies:
```bash
go get github.com/stretchr/testify
go get github.com/chromedp/chromedp
```

Provide complete test suite with:
- Table-driven tests
- Proper mocking
- Clear test names
- Test documentation
```

---

## Recommended Bubble Tea Resources

### Official Examples
- Basic example: https://github.com/charmbracelet/bubbletea/tree/master/examples/simple
- Text input: https://github.com/charmbracelet/bubbletea/tree/master/examples/textinput
- Viewport: https://github.com/charmbracelet/bubbletea/tree/master/examples/pager

### Charm Ecosystem
- **Bubbles**: Pre-built components (github.com/charmbracelet/bubbles)
- **Lipgloss**: Styling (github.com/charmbracelet/lipgloss)
- **Glamour**: Markdown rendering (github.com/charmbracelet/glamour)

### Video Tutorials
- "Building Terminal UIs with Bubble Tea" - Charm YouTube channel

---

## ChromeDP Resources

### Official Documentation
- ChromeDP Examples: https://github.com/chromedp/examples
- API Documentation: https://pkg.go.dev/github.com/chromedp/chromedp

### Useful ChromeDP Patterns
```go
// Navigate and wait for content
chromedp.Navigate(url),
chromedp.WaitVisible(`body`, chromedp.ByQuery),

// Extract text
chromedp.Text(`body`, &content, chromedp.ByQuery),

// Get all links
chromedp.Evaluate(`Array.from(document.querySelectorAll('a')).map(a => ({text: a.innerText, href: a.href}))`, &links),

// Set cookies
chromedp.ActionFunc(func(ctx context.Context) error {
    return network.SetCookies(cookieParams...).Do(ctx)
}),

// Get cookies
chromedp.ActionFunc(func(ctx context.Context) error {
    cookies, err := network.GetAllCookies().Do(ctx)
    // ... process cookies
    return err
}),
```

---

## Updated Timeline Estimate

With ChromeDP and Bubble Tea:
- **Beginner Go Developer**: 4-5 weeks (part-time)
- **Experienced Go Developer**: 2-3 weeks (part-time)

**Time savings**: ChromeDP eliminates FFI complexity, Bubble Tea provides robust UI framework.

---

## Key Advantages of This Approach

1. **No FFI Complexity**: Pure Go, no Rust/C bindings needed
2. **Full Browser Features**: JavaScript support, modern web standards
3. **Cookie Management**: Handled automatically by Chrome
4. **Modern UI Framework**: Bubble Tea provides excellent TUI capabilities
5. **Active Ecosystem**: Both ChromeDP and Charm are actively maintained
6. **Better Documentation**: More examples and community support
7. **Easier Debugging**: Chrome DevTools protocol provides detailed info

---

## Architecture Comparison

### Old Plan (Servo + tcell)
- Complex FFI layer
- Low-level terminal handling
- Manual cookie parsing
- Limited JavaScript support

### New Plan (ChromeDP + Bubble Tea)
- Pure Go, no FFI
- High-level UI framework
- Built-in cookie management
- Full JavaScript support
- Better developer experience

---

## Next Steps

1. **Start with Prompt 3A**: Set up ChromeDP and basic navigation
2. **Move to Prompt 3B**: Integrate cookie management
3. **Then Prompt 5A**: Build the Bubble Tea UI
4. **Continue sequentially** through remaining prompts

Each prompt builds on the previous one, creating a complete, functional browser.

---

## Success Criteria

- ✅ Browser loads pages using ChromeDP
- ✅ Cookies persist between sessions
- ✅ Beautiful, responsive Bubble Tea UI
- ✅ Keyboard navigation works smoothly
- ✅ Links are numbered and easily accessible
- ✅ History navigation (back/forward)
- ✅ Clean code with >70% test coverage
- ✅ Good performance (<3s page loads)

---

## Troubleshooting Common Issues

### ChromeDP Issues

**Chrome not found:**
```bash
# Ubuntu/Debian
sudo apt-get install chromium-browser

# Arch
sudo pacman -S chromium

# Or set custom Chrome path
chromedp.ExecPath("/path/to/chrome")
```

**Context deadline exceeded:**
- Increase timeout in chromedp.Navigate
- Check network connectivity
- Verify page is loading (try in regular Chrome)

**Cookies not persisting:**
- Ensure UserDataDir is set correctly
- Check file permissions on cookie storage
- Verify cookies are being extracted after navigation

### Bubble Tea Issues

**UI not rendering correctly:**
- Check terminal supports required features
- Test with different terminal emulators
- Verify window size detection

**Keyboard input not working:**
- Ensure terminal is in raw mode
- Check for conflicting key bindings
- Test with tea.WithAltScreen()

---

Good luck with your improved browser implementation! 🚀

The combination of ChromeDP and Bubble Tea will give you a much smoother development experience compared to the original Servo + tcell approach.