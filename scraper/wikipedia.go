// Package scraper provides implementations of specific website scrapers
package scraper

import (
	"fmt"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// WikipediaScraper handles scraping content from Wikipedia
type WikipediaScraper struct{}

// CanHandle determines if this scraper can handle the given URL
func (s *WikipediaScraper) CanHandle(url string) bool {
	return strings.Contains(url, "wikipedia.org")
}

// Scrape extracts content from a Wikipedia page and formats it as Markdown
func (s *WikipediaScraper) Scrape(doc *goquery.Document, url string) (*ScrapedContent, error) {
	var markdownBuilder strings.Builder
	
	// Extract the title and add as H1
	title := CleanText(doc.Find("#firstHeading").Text())
	if title != "" {
		markdownBuilder.WriteString(fmt.Sprintf("# %s\n\n", title))
	}
	
	// Add URL as a reference
	markdownBuilder.WriteString(fmt.Sprintf("*Source: %s*\n\n", url))
	
	// Extract the introduction paragraphs
	var introAdded bool
	doc.Find("#mw-content-text .mw-parser-output > p").Each(func(i int, s *goquery.Selection) {
		paraText := CleanText(s.Text())
		if paraText != "" {
			markdownBuilder.WriteString(paraText)
			markdownBuilder.WriteString("\n\n")
			introAdded = true
		}
	})
	
	if introAdded {
		markdownBuilder.WriteString("---\n\n")
	}
	
	// Extract sections with their content
	doc.Find("#mw-content-text .mw-parser-output > h2, #mw-content-text .mw-parser-output > h3").Each(func(i int, s *goquery.Selection) {
		// Get heading level and text
		headingLevel := 2 // h2
		if goquery.NodeName(s) == "h3" {
			headingLevel = 3 // h3
		}
		
		headingText := CleanText(s.Find(".mw-headline").Text())
		if headingText == "" {
			return
		}
		
		// Skip certain sections that usually don't contain useful content
		skipSections := []string{
			"References", "External links", "See also", "Further reading", 
			"Notes", "Bibliography", "Citations", "Sources", "Footnotes",
		}
		
		for _, skip := range skipSections {
			if strings.Contains(headingText, skip) {
				return
			}
		}
		
		// Add heading to markdown
		markdownBuilder.WriteString(fmt.Sprintf("%s %s\n\n", strings.Repeat("#", headingLevel), headingText))
		
		// Collect paragraphs for this section
		var sectionContent bool
		currentElement := s.Next()
		
		for currentElement.Length() > 0 && 
			goquery.NodeName(currentElement) != "h2" && 
			(headingLevel == 3 || goquery.NodeName(currentElement) != "h3") {
			
			if goquery.NodeName(currentElement) == "p" {
				paraText := CleanText(currentElement.Text())
				if paraText != "" {
					markdownBuilder.WriteString(paraText)
					markdownBuilder.WriteString("\n\n")
					sectionContent = true
				}
			} else if goquery.NodeName(currentElement) == "ul" || goquery.NodeName(currentElement) == "ol" {
				isList := goquery.NodeName(currentElement) == "ol"
				currentElement.Find("li").Each(func(i int, li *goquery.Selection) {
					liText := CleanText(li.Text())
					if liText != "" {
						if isList {
							markdownBuilder.WriteString(fmt.Sprintf("%d. %s\n", i+1, liText))
						} else {
							markdownBuilder.WriteString(fmt.Sprintf("* %s\n", liText))
						}
						sectionContent = true
					}
				})
				if sectionContent {
					markdownBuilder.WriteString("\n")
				}
			} else if goquery.NodeName(currentElement) == "table" {
				// Handle tables - simplified conversion to markdown
				tableContent := extractTableAsMarkdown(currentElement)
				if tableContent != "" {
					markdownBuilder.WriteString(tableContent)
					markdownBuilder.WriteString("\n\n")
					sectionContent = true
				}
			}
			
			currentElement = currentElement.Next()
		}
		
		// Add a separator if we added content
		if sectionContent {
			markdownBuilder.WriteString("---\n\n")
		}
	})
	
	// Create the content object
	content := &ScrapedContent{
		Markdown: strings.TrimSpace(markdownBuilder.String()),
	}
	
	return content, nil
}

// extractTableAsMarkdown converts a table to markdown format
func extractTableAsMarkdown(table *goquery.Selection) string {
	var sb strings.Builder
	
	// Extract headers
	var headers []string
	table.Find("tr th").Each(func(i int, th *goquery.Selection) {
		headers = append(headers, CleanText(th.Text()))
	})
	
	// If no headers found, try to get them from the first row
	if len(headers) == 0 {
		table.Find("tr:first-child td").Each(func(i int, td *goquery.Selection) {
			headers = append(headers, CleanText(td.Text()))
		})
	}
	
	// Skip empty tables
	if len(headers) == 0 {
		return ""
	}
	
	// Write headers
	for i, header := range headers {
		if i > 0 {
			sb.WriteString(" | ")
		}
		sb.WriteString(header)
	}
	sb.WriteString("\n")
	
	// Write separator row
	for i := 0; i < len(headers); i++ {
		if i > 0 {
			sb.WriteString(" | ")
		}
		sb.WriteString("---")
	}
	sb.WriteString("\n")
	
	// Write data rows
	var rowsProcessed bool
	table.Find("tr").Each(func(rowIdx int, tr *goquery.Selection) {
		// Skip header row if we already processed it
		if rowIdx == 0 && len(headers) > 0 && tr.Find("th").Length() > 0 {
			return
		}
		
		var rowData []string
		tr.Find("td").Each(func(cellIdx int, td *goquery.Selection) {
			rowData = append(rowData, CleanText(td.Text()))
		})
		
		// Skip empty rows
		if len(rowData) == 0 {
			return
		}
		
		// Write row data
		for i, cell := range rowData {
			if i > 0 {
				sb.WriteString(" | ")
			}
			sb.WriteString(cell)
		}
		sb.WriteString("\n")
		rowsProcessed = true
	})
	
	// Only return if we processed actual data rows
	if rowsProcessed {
		return sb.String()
	}
	return ""
}

// Register the Wikipedia scraper
func init() {
	DefaultRegistry.Register(&WikipediaScraper{})
}