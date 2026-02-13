package main

import (
	"fmt"
	"log"

	"github.com/jurikolo/go-browser/browser"
)

func main() {
	html := `<h1>HELLO!</h1>
<h2>hello h2</h2>
<p><i>Hello there</i></p>`

	formatted, links, err := browser.FormatHTML(html)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Formatted output:")
	fmt.Println(formatted)

	if len(links) > 0 {
		fmt.Println("\nLinks:")
		for _, link := range links {
			fmt.Printf("[%d] %s\n", link.Number, link.URL)
		}
	}
}