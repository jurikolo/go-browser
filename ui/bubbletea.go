package ui

import (
	"fmt"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
	"github.com/jurikolo/go-browser/internal/utils"
)

type Link struct {
	Text string
	URL  string
}

type NavigateMsg struct {
	URL string
}

type LinkClickMsg struct {
	LinkNumber int
}

type Styles struct {
	URLBar         lipgloss.Style
	URLInput       lipgloss.Style
	Content        lipgloss.Style
	StatusBar      lipgloss.Style
	Loading        lipgloss.Style
	CookieCount    lipgloss.Style
	ScrollPosition lipgloss.Style
	Help           lipgloss.Style
	Link           lipgloss.Style
	VisitedLink    lipgloss.Style
	LinkNumber     lipgloss.Style
	Heading        lipgloss.Style
	Error          lipgloss.Style
	Success        lipgloss.Style
}

type Model struct {
	viewport     viewport.Model
	urlInput     textinput.Model
	content      string
	links        []Link
	currentURL   string
	status       string
	mode         string
	width        int
	height       int
	history      []string
	historyIndex int
	styles       Styles
	darkMode     bool
}

type UpdateContentMsg struct {
	Content string
}

// Styles holds the styling for the UI
var (
	// URL Bar styles
	urlBarStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("6")).
			Padding(0, 1).
			Background(lipgloss.Color("#424242")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)

	urlInputStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("6")).
			Padding(0, 1).
			Background(lipgloss.Color("#2a2a2a")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Bold(true)

	// Content Viewport styles
	contentStyle = lipgloss.NewStyle().
			Padding(1, 2).
			MaxWidth(120).
			Foreground(lipgloss.Color("#E0E0E0")).
			Background(lipgloss.Color("#1a1a1a"))

	// Status Bar styles
	statusBarStyle = lipgloss.NewStyle().
			BorderStyle(lipgloss.NormalBorder()).
			BorderTop(true).
			BorderForeground(lipgloss.Color("8")).
			Padding(0, 1).
			Background(lipgloss.Color("#2a2a2a")).
			Foreground(lipgloss.Color("#FFFFFF")).
			Align(lipgloss.Left)

	loadingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("3")).
			Bold(true)

	cookieCountStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("5")).
				Bold(true)

	scrollPositionStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("4")).
				Bold(true)

	// Help Footer styles
	helpStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(lipgloss.Color("8")).
			Align(lipgloss.Center)

	// Link styles
	linkStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("12")).
			Bold(true).
			Underline(true)

	visitedLinkStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("8")).
				Underline(true)

	linkNumberStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("9")).
			Bold(true).
			Background(lipgloss.Color("7"))

	// Heading styles
	headingStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("2")).
			Bold(true).
			Underline(true)

	// Error styles
	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("1")).
			Bold(true)

	// Success styles
	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("2")).
			Bold(true)

	// Light mode styles
	lightURLBarStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("6")).
				Padding(0, 1).
				Background(lipgloss.Color("#E0E0E0")).
				Foreground(lipgloss.Color("#000000")).
				Bold(true)

	lightURLInputStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.RoundedBorder()).
				BorderForeground(lipgloss.Color("6")).
				Padding(0, 1).
				Background(lipgloss.Color("#F5F5F5")).
				Foreground(lipgloss.Color("#000000")).
				Bold(true)

	lightContentStyle = lipgloss.NewStyle().
				Padding(1, 2).
				MaxWidth(120).
				Foreground(lipgloss.Color("#212121")).
				Background(lipgloss.Color("#FFFFFF"))

	lightStatusBarStyle = lipgloss.NewStyle().
				BorderStyle(lipgloss.NormalBorder()).
				BorderTop(true).
				BorderForeground(lipgloss.Color("8")).
				Padding(0, 1).
				Background(lipgloss.Color("#F5F5F5")).
				Foreground(lipgloss.Color("#000000")).
				Align(lipgloss.Left)

	lightHelpStyle = lipgloss.NewStyle().
			Padding(0, 1).
			Foreground(lipgloss.Color("242")).
			Align(lipgloss.Center)

	lightLinkStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("12")).
			Bold(true).
			Underline(true)

	lightVisitedLinkStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("245")).
				Underline(true)

	lightLinkNumberStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("9")).
				Bold(true).
				Background(lipgloss.Color("7"))

	lightHeadingStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("2")).
				Bold(true).
				Underline(true)

	statusStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("2"))

	shortcutsStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("8"))

	headerStyle = urlBarStyle

	footerStyle = statusBarStyle
)

// Return the appropriate styles based on the dark mode setting
func GetStyles(darkMode bool) Styles {
	if darkMode {
		return Styles{
			URLBar:         urlBarStyle,
			URLInput:       urlInputStyle,
			Content:        contentStyle,
			StatusBar:      statusBarStyle,
			Loading:        loadingStyle,
			CookieCount:    cookieCountStyle,
			ScrollPosition: scrollPositionStyle,
			Help:           helpStyle,
			Link:           linkStyle,
			VisitedLink:    visitedLinkStyle,
			LinkNumber:     linkNumberStyle,
			Heading:        headingStyle,
			Error:          errorStyle,
			Success:        successStyle,
		}
	}
	return Styles{
		URLBar:         lightURLBarStyle,
		URLInput:       lightURLInputStyle,
		Content:        lightContentStyle,
		StatusBar:      lightStatusBarStyle,
		Loading:        loadingStyle,
		CookieCount:    cookieCountStyle,
		ScrollPosition: scrollPositionStyle,
		Help:           lightHelpStyle,
		Link:           lightLinkStyle,
		VisitedLink:    lightVisitedLinkStyle,
		LinkNumber:     lightLinkNumberStyle,
		Heading:        lightHeadingStyle,
		Error:          errorStyle,
		Success:        successStyle,
	}
}

// Update the content of the UI model
func (m *Model) UpdateContent(content string) {
	m.content = content
	m.viewport.SetContent(m.styles.Content.Render(m.content))
}

// Update the current URL of the UI model
func (m *Model) UpdateCurrentURL(url string) {
	m.currentURL = url
}

// Update the links of the UI model
func (m *Model) UpdateLinks(links []Link) {
	m.links = links
}

// Update the status of the UI model
func (m *Model) UpdateStatus(status string) {
	m.status = status
}

// Return the current URL
func (m *Model) GetCurrentURL() string {
	return m.currentURL
}

// Return the links
func (m *Model) GetLinks() []Link {
	return m.links
}

// Set the UI to a loading state
func (m *Model) SetLoading(url string) {
	m.status = "loading"
	m.content = fmt.Sprintf("Loading content from %s...", url)
	m.viewport.SetContent(m.styles.Content.Render(m.content))
}

func (m *Model) SetUrlPrefix() {
	m.currentURL = utils.SetUrlPrefix(m.currentURL)
}

// Init the model
func (m Model) Init() tea.Cmd {
	m.styles = GetStyles(m.darkMode)
	return textinput.Blink
}

// Handle events and update the model
func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	var cmds []tea.Cmd

	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.viewport.Width = msg.Width - 2
		m.viewport.Height = msg.Height - 5
		m.urlInput.Width = msg.Width - 4

		m.styles = GetStyles(m.darkMode)

	case tea.KeyMsg:
		switch m.mode {
		case "url-input":
			switch msg.String() {
			case "enter":
				m.currentURL = m.urlInput.Value()
				if m.currentURL == "" {
					m.status = "error"
					m.content = m.styles.Error.Render("Please enter a valid URL")
				} else {
					m.SetUrlPrefix()
					m.status = "loading"
					m.mode = "normal"
					m.content = fmt.Sprintf("Loading content from %s...", m.currentURL)
					m.history = append(m.history[:m.historyIndex+1], m.currentURL)
					m.historyIndex++
					return m, func() tea.Msg {
						return NavigateMsg{URL: m.currentURL}
					}
				}
			case "esc":
				m.mode = "normal"
			default:
				m.urlInput, cmd = m.urlInput.Update(msg)
				cmds = append(cmds, cmd)
				return m, tea.Batch(cmds...)
			}
		case "normal":
			switch msg.String() {
			case "g":
				m.mode = "url-input"
				m.urlInput.Focus()
				m.urlInput.SetValue("")
			case "q", "ctrl+c":
				return m, tea.Quit
			case "r":
				// Let the main application handle reload
				// m.status = "reloading"
				// m.content = fmt.Sprintf("Reloading content from %s...", m.currentURL)
			case "c":
				m.content = "Cookies for current site would be displayed here"
				m.status = "showing cookies"
			case "up", "k":
				m.viewport.LineUp(1)
			case "down", "j":
				m.viewport.LineDown(1)
			case "t":
				// Toggle between dark and light mode
				m.darkMode = !m.darkMode
				m.styles = GetStyles(m.darkMode)
			case "b":
				// Trigger back navigation by returning a special message
				return m, func() tea.Msg {
					return struct{ Back bool }{Back: true}
				}
			case "f":
				// Trigger forward navigation by returning a special message
				return m, func() tea.Msg {
					return struct{ Forward bool }{Forward: true}
				}
			case "1", "2", "3", "4", "5", "6", "7", "8", "9":
				// Handle link navigation
				linkNum := int(msg.String()[0] - '1')
				if linkNum < len(m.links) {
					// Trigger navigation by returning a LinkClickMsg
					return m, func() tea.Msg {
						return LinkClickMsg{LinkNumber: linkNum}
					}
				}
			}
		}
	}

	// Update viewport in all modes
	m.viewport, cmd = m.viewport.Update(msg)
	cmds = append(cmds, cmd)

	// Update urlInput in all modes except when already handled in url-input mode
	if m.mode != "url-input" {
		m.urlInput, cmd = m.urlInput.Update(msg)
		cmds = append(cmds, cmd)
	}

	return m, tea.Batch(cmds...)
}

// View renders the UI
func (m Model) View() string {
	var header, content, footer string

	// Header with URL bar
	if m.mode == "url-input" {
		header = m.styles.URLInput.Render(fmt.Sprintf("URL: %s", m.urlInput.View()))
	} else {
		header = m.styles.URLBar.Render(fmt.Sprintf("Current URL: %s", m.currentURL))
	}

	// Main content viewport
	m.viewport.SetContent(m.styles.Content.Render(m.content))
	content = m.viewport.View()

	// Footer with status and shortcuts
	statusText := fmt.Sprintf("Status: %s", m.status)
	if m.status == "error" {
		statusText = m.styles.Error.Render(statusText)
	} else if m.status == "loading" || m.status == "reloading" {
		statusText = m.styles.Loading.Render(statusText)
	} else {
		statusText = m.styles.Success.Render(statusText)
	}

	shortcuts := m.styles.Help.Render("g: URL | q: Quit | r: Reload | b: Back | f: Forward | c: Cookies | 1-9: Links | ↑/↓: Scroll")

	footer = m.styles.StatusBar.Render(fmt.Sprintf("%s | %s", statusText, shortcuts))

	return fmt.Sprintf("%s\n%s\n%s", header, content, footer)
}

// NewModel creates a new browser model
func NewModel() Model {
	urlInput := textinput.New()
	urlInput.Placeholder = "Enter URL..."
	urlInput.Focus()

	vp := viewport.New(80, 20)
	vp.SetContent("Welcome to the Bubble Tea Browser!\n\nPress 'g' to enter a URL.")

	// Initialize with dark mode by default
	styles := GetStyles(true)

	return Model{
		viewport:     vp,
		urlInput:     urlInput,
		content:      "Welcome to the Bubble Tea Browser!\n\nPress 'g' to enter a URL.",
		links:        []Link{},
		currentURL:   "https://jurikolo.name",
		status:       "ready",
		mode:         "normal",
		width:        80,
		height:       25,
		history:      []string{"https://jurikolo.name"},
		historyIndex: 0,
		styles:       styles,
		darkMode:     true,
	}
}
