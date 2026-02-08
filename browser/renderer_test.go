package browser

import (
	"strings"
	"testing"
)

func TestFormatHTML(t *testing.T) {
	tests := []struct {
		name     string
		html     string
		expected string
	}{
		{
			name:     "Simple text",
			html:     "<p>Hello world</p>",
			expected: "Hello world\n\n",
		},
		{
			name: "Headings",
			html: `<h1>Main Title</h1>
                   <h2>Subtitle</h2>
                   <h3>Section</h3>`,
			expected: "Main Title\n==========\nSubtitle\n--------\nSection\n\n",
		},
		{
			name: "Links",
			html: `<p>Visit <a href="https://example.com">Example</a> for more info.</p>`,
			expected: "Visit [1] Examplefor more info.\n\n\n\nLinks:\n[1] https://example.com\n",
		},
		{
			name: "Lists",
			html: `<ul>
                   <li>First item</li>
                   <li>Second item</li>
                   </ul>
                   <ol>
                   <li>Numbered item</li>
                   </ol>`,
			expected: "  First item\n  Second item\n\n  Numbered item\n\n",
		},
		{
			name: "Images",
			html: `<p>Check out this <img src="image.jpg" alt="sample image">.</p>`,
			expected: "Check out this [Image: sample image] .\n\n",
		},
		{
			name: "Code blocks",
			html: `<p>Here's some code:</p>
                   <pre><code>func main() {
                       fmt.Println("Hello")
                   }</code></pre>`,
			expected: "Here's some code:\n\n  func main() { fmt.Println(\"Hello\") }\n\n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, _, err := FormatHTML(tt.html)
			if err != nil {
				t.Fatalf("FormatHTML failed: %v", err)
			}

			// Normalize whitespace for comparison
			expected := strings.TrimSpace(tt.expected)
			actual := strings.TrimSpace(result)

			if actual != expected {
				t.Errorf("Expected:\n%q\n\nGot:\n%q", expected, actual)
			}
		})
	}
}

func TestExtractLinks(t *testing.T) {
	html := `<p>Visit <a href="https://example.com">Example</a> or <a href="https://test.com">Test</a>.</p>`
	
	links, err := ExtractLinks(html)
	if err != nil {
		t.Fatalf("ExtractLinks failed: %v", err)
	}

	if len(links) != 2 {
		t.Fatalf("Expected 2 links, got %d", len(links))
	}

	if links[0].Number != 1 || links[0].Text != "Example" || links[0].URL != "https://example.com" {
		t.Errorf("First link incorrect: %+v", links[0])
	}

	if links[1].Number != 2 || links[1].Text != "Test" || links[1].URL != "https://test.com" {
		t.Errorf("Second link incorrect: %+v", links[1])
	}
}

func TestStripScriptAndStyle(t *testing.T) {
	html := `<html>
                   <head>
                       <script>alert('test');</script>
                       <style>body { color: red; }</style>
                   </head>
                   <body>
                       <p>Visible content</p>
                   </body>
               </html>`
	
	result, _, err := FormatHTML(html)
	if err != nil {
		t.Fatalf("FormatHTML failed: %v", err)
	}

	if strings.Contains(result, "alert") {
		t.Error("Script content was not stripped")
	}

	if strings.Contains(result, "color: red") {
		t.Error("Style content was not stripped")
	}

	if !strings.Contains(result, "Visible content") {
		t.Error("Visible content was incorrectly stripped")
	}
}