//go:build ignore

package main

import (
	"fmt"
	"os"
	"log"

	"github.com/jurikolo/go-browser/browser"
)

func main() {
	// Read the test HTML file
	content, err := os.ReadFile("test_headings.html")
	if err != nil {
		log.Fatal(err)
	}

	// Format the HTML
	formatted, links, err := browser.FormatHTML(string(content))
	if err != nil {
		log.Fatal(err)
	}

	// Print the result
	fmt.Println("Formatted output:")
	fmt.Println(formatted)
	
	// Print links if any
	if len(links) > 0 {
		fmt.Println("\nLinks:")
		for _, link := range links {
			fmt.Printf("[%d] %s\n", link.Number, link.URL)
		}
	}
}