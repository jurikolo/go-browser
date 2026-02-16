// Package browser provides utilities for rendering HTML content as formatted text
package browser

import (
	"bytes"
	"fmt"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

type RenderedLink struct {
	Number int
	Text   string
	URL    string
}

type Renderer struct {
	linkCounter int
	links       []RenderedLink
}

// Create a new Renderer instance
func NewRenderer() *Renderer {
	return &Renderer{
		linkCounter: 0,
		links:       make([]RenderedLink, 0),
	}
}

// Convert HTML string to formatted plain text
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

	// Add link footers
	if len(r.links) > 0 {
		buf.WriteString("\n\nLinks:\n")
		for _, link := range r.links {
			buf.WriteString(fmt.Sprintf("[%d] %s\n", link.Number, link.URL))
		}
	}

	return buf.String(), r.links, nil
}

// Process HTML nodes and convert to formatted text
func (r *Renderer) renderNode(buf *bytes.Buffer, n *html.Node) {
	// Skip script and style elements entirely
	if n.Type == html.ElementNode && (n.Data == "script" || n.Data == "style") {
		return
	}

	if n.Type == html.ElementNode {
		r.renderElement(buf, n)
	}

	if n.Type == html.TextNode {
		text := strings.TrimSpace(n.Data)
		if text != "" {
			text = regexp.MustCompile(`\s+`).ReplaceAllString(text, " ")
			buf.WriteString(text)
		}
	}

	for c := n.FirstChild; c != nil; c = c.NextSibling {
		r.renderNode(buf, c)
	}

	if n.Type == html.ElementNode {
		r.addPostElementFormatting(buf, n)
	}
}

// Handle the opening of HTML elements
func (r *Renderer) renderElement(buf *bytes.Buffer, n *html.Node) {
	switch n.Data {
	case "h1":
		if buf.Len() > 0 && !strings.HasSuffix(buf.String(), "\n") {
			buf.WriteString("\n")
		}
		buf.WriteString("# ")
	case "h2":
		if buf.Len() > 0 && !strings.HasSuffix(buf.String(), "\n") {
			buf.WriteString("\n")
		}
		buf.WriteString("## ")
	case "h3":
		if buf.Len() > 0 && !strings.HasSuffix(buf.String(), "\n") {
			buf.WriteString("\n")
		}
		buf.WriteString("### ")
	case "h4":
		if buf.Len() > 0 && !strings.HasSuffix(buf.String(), "\n") {
			buf.WriteString("\n")
		}
		buf.WriteString("#### ")
	case "h5":
		if buf.Len() > 0 && !strings.HasSuffix(buf.String(), "\n") {
			buf.WriteString("\n")
		}
		buf.WriteString("##### ")
	case "h6":
		if buf.Len() > 0 && !strings.HasSuffix(buf.String(), "\n") {
			buf.WriteString("\n")
		}
		buf.WriteString("###### ")
	case "p":
		if buf.Len() > 0 && !strings.HasSuffix(buf.String(), "\n\n") && !strings.HasSuffix(buf.String(), "\n") {
			buf.WriteString("\n")
		}
	case "br":
		buf.WriteString("\n")
	case "li":
		buf.WriteString("  ")
	case "a":
		for _, attr := range n.Attr {
			if attr.Key == "href" {
				r.linkCounter++
				link := RenderedLink{
					Number: r.linkCounter,
					Text:   r.extractText(n),
					URL:    attr.Val,
				}
				r.links = append(r.links, link)
				buf.WriteString(fmt.Sprintf(" [%d] ", r.linkCounter))
				break
			}
		}
	case "img":
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
		if buf.Len() > 0 && !strings.HasSuffix(buf.String(), "\n") {
			buf.WriteString("\n")
		}
		buf.WriteString("  ")
	case "code":
		if buf.Len() > 0 && !strings.HasSuffix(buf.String(), " ") && !strings.HasSuffix(buf.String(), "\n") {
			buf.WriteString(" ")
		}
	case "i", "em":
		buf.WriteString("*")
	}
}

// addPostElementFormatting handles formatting that comes after an element's content
func (r *Renderer) addPostElementFormatting(buf *bytes.Buffer, n *html.Node) {
	switch n.Data {
	case "h1", "h2", "h3", "h4", "h5", "h6":
		if buf.Len() > 0 && !strings.HasSuffix(buf.String(), "\n") {
			buf.WriteString("\n")
		}
	case "p":
		if buf.Len() > 0 && !strings.HasSuffix(buf.String(), "\n\n") {
			buf.WriteString("\n\n")
		}
	case "ul", "ol":
		if buf.Len() > 0 && !strings.HasSuffix(buf.String(), "\n\n") {
			buf.WriteString("\n")
		}
	case "li":
		buf.WriteString("\n")
	case "br":
		return
	case "div", "section", "article":
		if buf.Len() > 0 && !strings.HasSuffix(buf.String(), "\n") {
			buf.WriteString("\n")
		}
	case "pre":
		if buf.Len() > 0 && !strings.HasSuffix(buf.String(), "\n\n") {
			buf.WriteString("\n\n")
		}
	case "code":
		if buf.Len() > 0 && !strings.HasSuffix(buf.String(), " ") && !strings.HasSuffix(buf.String(), "\n") {
			buf.WriteString(" ")
		}
	case "i", "em":
		buf.WriteString("*")
	}
}

// Extract all text content from a node and its children
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

// Create a renderer and format HTML content
func FormatHTML(htmlContent string) (string, []RenderedLink, error) {
	renderer := NewRenderer()
	return renderer.FormatHTML(htmlContent)
}

// Extract all links from HTML content
func ExtractLinks(htmlContent string) ([]RenderedLink, error) {
	renderer := NewRenderer()
	_, links, err := renderer.FormatHTML(htmlContent)
	if err != nil {
		return nil, err
	}
	return links, nil
}
