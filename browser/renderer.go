// Package browser provides utilities for rendering HTML content as formatted text
package browser

import (
	"bytes"
	"fmt"
	"golang.org/x/net/html"
	"regexp"
	"strings"
)

// RenderedLink represents a hyperlink with its number, text, and URL for easy navigation
// This is different from the Link struct in chromedp.go which only has Text and URL
type RenderedLink struct {
	Number int
	Text   string
	URL    string
}

// Renderer handles the conversion of HTML to formatted text
type Renderer struct {
	linkCounter int
	links       []RenderedLink
}

// NewRenderer creates a new Renderer instance
func NewRenderer() *Renderer {
	return &Renderer{
		linkCounter: 0,
		links:       make([]RenderedLink, 0),
	}
}

// FormatHTML takes an HTML string and converts it to formatted plain text
// It preserves structure and handles special elements like headings, lists, links, etc.
func (r *Renderer) FormatHTML(htmlContent string) (string, []RenderedLink, error) {
	// Reset the renderer state
	r.linkCounter = 0
	r.links = make([]RenderedLink, 0)

	doc, err := html.Parse(strings.NewReader(htmlContent))
	if err != nil {
		return "", nil, fmt.Errorf("failed to parse HTML: %w", err)
	}

	var buf bytes.Buffer
	r.renderNode(&buf, doc)

	// Add link footers if there are any links
	if len(r.links) > 0 {
		buf.WriteString("\n\nLinks:\n")
		for _, link := range r.links {
			buf.WriteString(fmt.Sprintf("[%d] %s\n", link.Number, link.URL))
		}
	}

	return buf.String(), r.links, nil
}

// renderNode recursively processes HTML nodes and converts them to formatted text
func (r *Renderer) renderNode(buf *bytes.Buffer, n *html.Node) {
	// Skip script and style elements entirely
	if n.Type == html.ElementNode && (n.Data == "script" || n.Data == "style") {
		return
	}

	// Handle element nodes first
	if n.Type == html.ElementNode {
		r.renderElement(buf, n)
	}

	// Process text nodes
	if n.Type == html.TextNode {
		// Process text nodes, trimming excess whitespace
		text := strings.TrimSpace(n.Data)
		if text != "" {
			// Normalize whitespace
			text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
			buf.WriteString(text)
		}
	}

	// Process child nodes
	for c := n.FirstChild; c != nil; c = c.NextSibling {
		r.renderNode(buf, c)
	}

	// Add post-element formatting
	if n.Type == html.ElementNode {
		r.addPostElementFormatting(buf, n)
	}
}

// renderElement handles the opening of HTML elements
func (r *Renderer) renderElement(buf *bytes.Buffer, n *html.Node) {
	switch n.Data {
	case "h1", "h2", "h3", "h4", "h5", "h6":
		// Add spacing before headings
		if buf.Len() > 0 && !strings.HasSuffix(buf.String(), "\n") {
			buf.WriteString("\n")
		}
	case "p":
		// Paragraphs need blank lines before them (if not the first element)
		if buf.Len() > 0 && !strings.HasSuffix(buf.String(), "\n\n") && !strings.HasSuffix(buf.String(), "\n") {
			buf.WriteString("\n")
		}
	case "br":
		buf.WriteString("\n")
	case "li":
		// Add indentation for list items
		buf.WriteString("  ")
	case "a":
		// Handle links by adding them to our collection
		for _, attr := range n.Attr {
			if attr.Key == "href" {
				r.linkCounter++
				link := RenderedLink{
					Number: r.linkCounter,
					Text:   r.extractText(n),
					URL:    attr.Val,
				}
				r.links = append(r.links, link)
				// Add link reference in text
				buf.WriteString(fmt.Sprintf(" [%d] ", r.linkCounter))
				break
			}
		}
	case "img":
		// Handle images
		altText := ""
		for _, attr := range n.Attr {
			if attr.Key == "alt" {
				altText = attr.Val
				break
			}
		}
		if altText == "" {
			altText = "image"
		}
		buf.WriteString(fmt.Sprintf(" [Image: %s] ", altText))
	case "pre":
		// Add indentation for code blocks
		if buf.Len() > 0 && !strings.HasSuffix(buf.String(), "\n") {
			buf.WriteString("\n")
		}
		buf.WriteString("  ")
	case "code":
		// For inline code, just add a space before if needed
		if buf.Len() > 0 && !strings.HasSuffix(buf.String(), " ") && !strings.HasSuffix(buf.String(), "\n") {
			buf.WriteString(" ")
		}
	}
}

// addPostElementFormatting handles formatting that comes after an element's content
func (r *Renderer) addPostElementFormatting(buf *bytes.Buffer, n *html.Node) {
	switch n.Data {
	case "h1":
		// For h1, we need to extract the text that was written to the buffer
		// and then add the underline
		content := buf.String()
		lines := strings.Split(content, "\n")
		if len(lines) > 0 {
			lastLine := strings.TrimSpace(lines[len(lines)-1])
			if lastLine != "" {
				underline := strings.Repeat("=", len(lastLine))
				buf.WriteString(fmt.Sprintf("\n%s\n", underline))
			}
		}
	case "h2":
		// For h2, we need to extract the text that was written to the buffer
		// and then add the underline
		content := buf.String()
		lines := strings.Split(content, "\n")
		if len(lines) > 0 {
			lastLine := strings.TrimSpace(lines[len(lines)-1])
			if lastLine != "" {
				underline := strings.Repeat("-", len(lastLine))
				buf.WriteString(fmt.Sprintf("\n%s\n", underline))
			}
		}
	case "h3", "h4", "h5", "h6":
		// For other headings, just add a newline
		if buf.Len() > 0 && !strings.HasSuffix(buf.String(), "\n") {
			buf.WriteString("\n")
		}
	case "p":
		// Add blank line after paragraphs
		if buf.Len() > 0 && !strings.HasSuffix(buf.String(), "\n\n") {
			buf.WriteString("\n\n")
		}
	case "ul", "ol":
		// Add blank line after lists
		if buf.Len() > 0 && !strings.HasSuffix(buf.String(), "\n\n") {
			buf.WriteString("\n")
		}
	case "li":
		// Add newline after list items
		buf.WriteString("\n")
	case "br":
		// Line break already handled in renderElement
		return
	case "div", "section", "article":
		// Add spacing around block-level elements
		if buf.Len() > 0 && !strings.HasSuffix(buf.String(), "\n") {
			buf.WriteString("\n")
		}
	case "pre":
		// Add newline after code blocks
		if buf.Len() > 0 && !strings.HasSuffix(buf.String(), "\n\n") {
			buf.WriteString("\n\n")
		}
	case "code":
		// For inline code, just add a space after if needed
		if buf.Len() > 0 && !strings.HasSuffix(buf.String(), " ") && !strings.HasSuffix(buf.String(), "\n") {
			buf.WriteString(" ")
		}
	}
}

// extractText extracts all text content from a node and its children
func (r *Renderer) extractText(n *html.Node) string {
	var buf bytes.Buffer
	var extract func(*html.Node)
	
	extract = func(node *html.Node) {
		if node.Type == html.TextNode {
			buf.WriteString(node.Data)
		}
		for c := node.FirstChild; c != nil; c = c.NextSibling {
			extract(c)
		}
	}
	
	extract(n)
	return strings.TrimSpace(buf.String())
}

// FormatHTML is a convenience function that creates a renderer and formats HTML content
func FormatHTML(htmlContent string) (string, []RenderedLink, error) {
	renderer := NewRenderer()
	return renderer.FormatHTML(htmlContent)
}

// ExtractLinks is a helper function that extracts all links from HTML content
func ExtractLinks(htmlContent string) ([]RenderedLink, error) {
	renderer := NewRenderer()
	_, links, err := renderer.FormatHTML(htmlContent)
	if err != nil {
		return nil, err
	}
	return links, nil
}