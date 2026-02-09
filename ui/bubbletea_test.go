package ui

import "testing"

func TestSetUrlPrefix(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "Plain domain",
			input:    "asdf.xyz",
			expected: "https://asdf.xyz",
		},
		{
			name:     "HTTP URL",
			input:    "http://example.com",
			expected: "http://example.com",
		},
		{
			name:     "HTTPS URL",
			input:    "https://jurikolo.name",
			expected: "https://jurikolo.name",
		},
		{
			name:     "Local file path",
			input:    "/tmp/test.html",
			expected: "/tmp/test.html",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			model := NewModel()
			model.currentURL = tt.input
			model.SetUrlPrefix(tt.input)
			
			if model.currentURL != tt.expected {
				t.Errorf("SetUrlPrefix(%s) = %s; expected %s", tt.input, model.currentURL, tt.expected)
			}
		})
	}
}
