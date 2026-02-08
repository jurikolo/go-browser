package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/jurikolo/go-browser/browser"
	"github.com/jurikolo/go-browser/cookie"
	"github.com/jurikolo/go-browser/ui"
)

// PageContent represents the content extracted from a web page
type PageContent struct {
	Title string
	Text  string
	Links []browser.Link
	URL   string
}

// Custom messages for updating the UI
type updateContentMsg struct {
	Content *PageContent
}

type updateErrorMsg struct {
	Error error
}

// Forward declaration of UI message types
type NavigateMsg = ui.NavigateMsg
type LinkClickMsg = ui.LinkClickMsg

// AppModel represents the application state
type AppModel struct {
	browser       *browser.Browser
	cookieManager *cookie.CookieManager
	uiModel       ui.Model
	loading       bool
	currentURL    string
}

// navigateToURL navigates to a URL in a goroutine and returns a command
func (m *AppModel) navigateToURL(url string) tea.Cmd {
	log.Printf("navigateToURL called with URL: %s", url)
	m.currentURL = url
	m.loading = true
	// Set the UI to loading state
	m.uiModel.SetLoading(url)
	
	return func() tea.Msg {
		// Create a timeout context for the navigation
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		
		// Channel for navigation result
		resultChan := make(chan interface{}, 1)
		
		// Start navigation in goroutine
		go func() {
			log.Printf("Starting navigation to: %s", url)
			// Navigate to the URL
			if err := m.browser.Navigate(url); err != nil {
				log.Printf("Navigation error: %v", err)
				resultChan <- err
				return
			}
			
			// Extract page content
			title, err := m.browser.GetPageTitle()
			if err != nil {
				log.Printf("Error getting page title: %v", err)
				resultChan <- err
				return
			}
			
			text, err := m.browser.GetPageText()
			if err != nil {
				log.Printf("Error getting page text: %v", err)
				resultChan <- err
				return
			}
			
			links, err := m.browser.GetLinks()
			if err != nil {
				log.Printf("Error getting page links: %v", err)
				resultChan <- err
				return
			}
			
			// Send the result
			resultChan <- &PageContent{
				Title: title,
				Text:  text,
				Links: links,
				URL:   url,
			}
		}()
		
		// Wait for result or timeout
		log.Printf("Waiting for navigation result or timeout")
		select {
		case result := <-resultChan:
			switch r := result.(type) {
			case error:
				log.Printf("Navigation resulted in error: %v", r)
				return updateErrorMsg{Error: r}
			case *PageContent:
				// Sync cookies after successful navigation
				m.syncCookies()
				return updateContentMsg{Content: r}
			}
		case <-ctx.Done():
			log.Printf("Navigation timed out or cancelled: %v", ctx.Err())
			if ctx.Err() == context.DeadlineExceeded {
				return updateErrorMsg{Error: fmt.Errorf("navigation timeout")}
			}
			return updateErrorMsg{Error: fmt.Errorf("navigation cancelled")}
		}
		
		// This should never be reached, but required for compilation
		log.Printf("Unexpected code path reached in navigateToURL")
		return updateErrorMsg{Error: fmt.Errorf("unexpected error")}
	}
}

// updateUIWithPage updates the UI with the page content
func (m *AppModel) updateUIWithPage(content *PageContent) tea.Cmd {
	// Format the content for display
	m.formatContent(content)
	
	// Since we can't directly access the UI model's fields,
	// we'll need to work within the existing framework
	// For now, we'll just return nil as we'll handle the update in the Update method
	return nil
}

// handleLinkClick navigates to a numbered link
func (m *AppModel) handleLinkClick(linkNumber int) tea.Cmd {
	// For now, we'll just log the link click
	log.Printf("Handling link click for link number: %d", linkNumber)
	
	// Get the links from the UI model
	links := m.uiModel.GetLinks()
	if linkNumber >= 0 && linkNumber < len(links) {
		// Navigate to the link URL
		return m.navigateToURL(links[linkNumber].URL)
	}
	return nil
}

// syncCookies saves cookies to disk
func (m *AppModel) syncCookies() {
	if err := m.cookieManager.ExtractCookiesFromChromedp(m.browser.Context()); err != nil {
		log.Printf("Warning: Failed to extract cookies: %v", err)
		return
	}
	
	if err := m.cookieManager.SaveCookiesToDisk(); err != nil {
		log.Printf("Warning: Failed to save cookies: %v", err)
		return
	}
}

// formatContent formats the page content for display
func (m *AppModel) formatContent(content *PageContent) string {
	result := fmt.Sprintf("# %s\n\n", content.Title)
	
	// Add the page text
	if content.Text != "" {
		result += content.Text + "\n\n"
	}
	
	// Add links section
	if len(content.Links) > 0 {
		result += "## Links\n\n"
		for i, link := range content.Links {
			if link.Text != "" && link.URL != "" {
				result += fmt.Sprintf("[%d] %s -> %s\n", i+1, link.Text, link.URL)
			}
		}
	}
	
	return result
}

// Init initializes the application
func (m AppModel) Init() tea.Cmd {
	return m.navigateToURL("https://jurikolo.name")
}

// Update handles events and updates the model
func (m AppModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	
	switch msg := msg.(type) {
	case NavigateMsg:
		// Handle navigation request from UI
		log.Printf("Received NavigateMsg for URL: %s", msg.URL)
		return m, m.navigateToURL(msg.URL)
		
	case LinkClickMsg:
		// Handle link click request from UI
		log.Printf("Received LinkClickMsg for link number: %d", msg.LinkNumber)
		return m, m.handleLinkClick(msg.LinkNumber)
		
	case updateContentMsg:
		// Update UI with the page content
		formattedContent := m.formatContent(msg.Content)
		m.loading = false
		
		// Update the UI model with the new content
		m.uiModel.UpdateContent(formattedContent)
		m.uiModel.UpdateCurrentURL(msg.Content.URL)
		m.uiModel.UpdateStatus("loaded")
		// Convert browser links to UI links
		uiLinks := make([]ui.Link, len(msg.Content.Links))
		for i, link := range msg.Content.Links {
			uiLinks[i] = ui.Link{Text: link.Text, URL: link.URL}
		}
		m.uiModel.UpdateLinks(uiLinks)
		log.Printf("UI model updated with content, URL: %s, links: %d", msg.Content.URL, len(uiLinks))
		
	case updateErrorMsg:
		// Handle navigation error
		log.Printf("Received updateErrorMsg: %v", msg.Error)
		log.Printf("Navigation failed: %v", msg.Error)
		m.loading = false
		// Update UI with error message
		m.uiModel.UpdateContent(fmt.Sprintf("Error: %v", msg.Error))
		m.uiModel.UpdateStatus("error")
		
	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c", "q":
			// Graceful shutdown
			m.syncCookies()
			m.browser.Close()
			return m, tea.Quit
		case "r":
			// Reload current URL
			if m.currentURL != "" && !m.loading {
				return m, m.navigateToURL(m.currentURL)
			}
		case "1", "2", "3", "4", "5", "6", "7", "8", "9":
			if !m.loading {
				// Handle link navigation
				linkNumber := int(msg.String()[0] - '1')
				return m, m.handleLinkClick(linkNumber)
			}
		case "g":
			// Pass through to UI for URL input
			uiModel, uiCmd := m.uiModel.Update(msg)
			m.uiModel = uiModel.(ui.Model)
			cmd = uiCmd
			return m, cmd
		}
		
	case tea.WindowSizeMsg:
		// Handle window size changes by updating the UI model
		uiModel, uiCmd := m.uiModel.Update(msg)
		m.uiModel = uiModel.(ui.Model)
		cmd = uiCmd
		return m, cmd
	}
	
	// Update UI model for other messages
	uiModel, uiCmd := m.uiModel.Update(msg)
	m.uiModel = uiModel.(ui.Model)
	cmd = uiCmd
	
	return m, cmd
}

// View renders the UI
func (m AppModel) View() string {
	return m.uiModel.View()
}

func main() {
	// Set up signal handling for graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
	
	go func() {
		<-sigChan
		fmt.Println("\nReceived interrupt signal, shutting down gracefully...")
	}()
	
	// Initialize the cookie manager
	cookieManager, err := cookie.NewCookieManager()
	if err != nil {
		log.Fatalf("Failed to create cookie manager: %v", err)
	}
	
	// Initialize the browser
	browser, err := browser.NewBrowser()
	if err != nil {
		log.Fatalf("Failed to create browser: %v", err)
	}
	
	// Ensure browser is closed when main function exits
	defer func() {
		// Save cookies before closing
		if err := cookieManager.ExtractCookiesFromChromedp(browser.Context()); err != nil {
			log.Printf("Warning: Failed to extract cookies: %v", err)
		} else {
			if err := cookieManager.SaveCookiesToDisk(); err != nil {
				log.Printf("Warning: Failed to save cookies: %v", err)
			}
		}
		browser.Close()
	}()
	
	// Load cookies into the browser on startup
	fmt.Println("Loading cookies into browser...")
	if err := cookieManager.LoadCookiesIntoChromedp(browser.Context()); err != nil {
		log.Printf("Warning: Failed to load cookies into browser: %v", err)
	}
	
	// Create UI model
	uiModel := ui.NewModel()
	
	// Create app model
	model := AppModel{
		browser:       browser,
		cookieManager: cookieManager,
		uiModel:       uiModel,
		loading:       false,
		currentURL:    "",
	}
	
	// Start Bubble Tea program
	p := tea.NewProgram(model, tea.WithAltScreen())
	if _, err := p.Run(); err != nil {
		log.Fatalf("Bubble Tea program error: %v", err)
	}
	
	fmt.Println("Browser session completed successfully!")
}
