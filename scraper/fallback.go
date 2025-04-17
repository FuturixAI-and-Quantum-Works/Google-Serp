// Package scraper provides implementations of specific website scrapers
package scraper

import (
	"fmt"
	"strings"
	"unicode"

	"github.com/PuerkitoBio/goquery"
)

// FallbackScraper is a generic scraper that attempts to extract meaningful content from any webpage
type FallbackScraper struct{}

// CanHandle always returns true since this is a fallback scraper
func (s *FallbackScraper) CanHandle(url string) bool {
	return true
}

// Scrape attempts to extract the most relevant content from a generic webpage and format as Markdown
func (s *FallbackScraper) Scrape(doc *goquery.Document, url string) (*ScrapedContent, error) {
	var markdownBuilder strings.Builder

	// Add the title as H1
	title := extractTitle(doc)
	if title != "" {
		markdownBuilder.WriteString(fmt.Sprintf("# %s\n\n", title))
	}

	// Add URL as a reference
	markdownBuilder.WriteString(fmt.Sprintf("*Source: %s*\n\n", url))

	// Try to find the main content container
	mainContent := extractMainContentNode(doc)
	if mainContent != nil {
		// Extract all headings and their content
		processContentAsMarkdown(mainContent, &markdownBuilder)
	} else {
		// Fallback: Process the entire document
		processContentAsMarkdown(doc.Selection, &markdownBuilder)
	}

	// Create the content object
	content := &ScrapedContent{
		Markdown: strings.TrimSpace(markdownBuilder.String()),
	}

	return content, nil
}

// extractTitle gets the most likely title of the page
func extractTitle(doc *goquery.Document) string {
	// First try the title tag
	title := CleanText(doc.Find("title").First().Text())

	// If title is empty or too long, try h1
	if title == "" || len(title) > 150 {
		title = CleanText(doc.Find("h1").First().Text())
	}

	// Try open graph title as a last resort
	if title == "" {
		title, _ = doc.Find("meta[property='og:title']").Attr("content")
		title = CleanText(title)
	}

	return title
}

// extractMainContentNode attempts to find the main content container
func extractMainContentNode(doc *goquery.Document) *goquery.Selection {
	// Common content containers
	contentSelectors := []string{
		"article", "main", ".content", "#content", ".post", ".article",
		".entry-content", ".post-content", ".article-content", "#main-content",
	}

	// Try each selector and return the first one with substantial content
	for _, selector := range contentSelectors {
		container := doc.Find(selector).First()
		if container.Length() > 0 {
			// See if this container has a substantial amount of text
			text := container.Text()
			wordCount := countWords(text)

			if wordCount > 100 { // Arbitrary threshold for "enough" content
				return container
			}
		}
	}

	return nil
}

// processContentAsMarkdown converts HTML structure to markdown format
func processContentAsMarkdown(selection *goquery.Selection, sb *strings.Builder) {
	// Keep track of heading stack to maintain document structure
	headings := make(map[string]bool)

	// First process headings to create document structure
	selection.Find("h1, h2, h3, h4, h5, h6").Each(func(i int, s *goquery.Selection) {
		headingText := CleanText(s.Text())
		if headingText == "" {
			return
		}

		// Determine heading level
		level := 0
		switch goquery.NodeName(s) {
		case "h1":
			level = 1
		case "h2":
			level = 2
		case "h3":
			level = 3
		case "h4":
			level = 4
		case "h5":
			level = 5
		case "h6":
			level = 6
		}

		// Add to tracking map
		headings[fmt.Sprintf("%s-%d", headingText, level)] = true
	})

	// Now process all content in a linear fashion
	selection.Contents().Each(func(i int, s *goquery.Selection) {
		nodeName := goquery.NodeName(s)

		// Handle different node types
		switch nodeName {
		case "h1", "h2", "h3", "h4", "h5", "h6":
			headingText := CleanText(s.Text())
			if headingText == "" {
				return
			}

			// Determine heading level
			level := 1
			switch nodeName {
			case "h1":
				level = 1
			case "h2":
				level = 2
			case "h3":
				level = 3
			case "h4":
				level = 4
			case "h5":
				level = 5
			case "h6":
				level = 6
			}

			// Only process if we're tracking this heading
			headingKey := fmt.Sprintf("%s-%d", headingText, level)
			if headings[headingKey] {
				// Add heading to markdown
				sb.WriteString(fmt.Sprintf("\n%s %s\n\n", strings.Repeat("#", level), headingText))
				headings[headingKey] = false // Mark as processed
			}

		case "p":
			paraText := CleanText(s.Text())
			if paraText != "" {
				sb.WriteString(paraText)
				sb.WriteString("\n\n")
			}

		case "ul", "ol":
			isList := nodeName == "ol"
			s.Find("li").Each(func(i int, li *goquery.Selection) {
				liText := CleanText(li.Text())
				if liText != "" {
					if isList {
						sb.WriteString(fmt.Sprintf("%d. %s\n", i+1, liText))
					} else {
						sb.WriteString(fmt.Sprintf("* %s\n", liText))
					}
				}
			})
			sb.WriteString("\n")

		case "table":
			// Handle tables - simplified conversion to markdown
			tableContent := extractTableAsMarkdown(s)
			if tableContent != "" {
				sb.WriteString(tableContent)
				sb.WriteString("\n\n")
			}

		case "blockquote":
			quoteText := CleanText(s.Text())
			if quoteText != "" {
				// Format as markdown blockquote
				lines := strings.Split(quoteText, "\n")
				for _, line := range lines {
					if trimmed := strings.TrimSpace(line); trimmed != "" {
						sb.WriteString("> " + trimmed + "\n")
					}
				}
				sb.WriteString("\n")
			}

		case "pre", "code":
			codeText := s.Text() // Preserve whitespace
			if codeText != "" {
				sb.WriteString("```\n")
				sb.WriteString(codeText)
				sb.WriteString("\n```\n\n")
			}

		case "div", "section", "article":
			// Recursively process div contents
			processContentAsMarkdown(s, sb)
		}
	})
}

// countWords counts the number of words in a string
func countWords(s string) int {
	words := 0
	inWord := false

	for _, r := range s {
		if unicode.IsSpace(r) {
			inWord = false
		} else {
			if !inWord {
				words++
				inWord = true
			}
		}
	}

	return words
}

// Register the fallback scraper
func init() {
	DefaultRegistry.SetFallback(&FallbackScraper{})
}
