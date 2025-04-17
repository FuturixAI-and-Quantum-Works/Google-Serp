package bingsearch

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// BingAnswerBox represents the structured data from Bing's answer box
type BingAnswerBox struct {
	Type       string                 `json:"type"`
	Content    interface{}            `json:"content,omitempty"`
	Title      string                 `json:"title,omitempty"`
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// ExtractAnswerbox extracts any answer box from Bing search results
func ExtractAnswerbox(doc *goquery.Document) *BingAnswerBox {
	// Try to extract different types of answer boxes in order of priority

	// First check for info box (entity)
	if infoBox := ExtractInfoBox(doc); infoBox != nil {
		return infoBox
	}

	// Check for Weather box
	if weatherBox := ExtractWeatherBox(doc); weatherBox != nil {
		return weatherBox
	}

	// Check for Time box
	if timeBox := ExtractTimeBox(doc); timeBox != nil {
		return timeBox
	}

	// Check for Stock box
	if stockBox := ExtractStockBox(doc); stockBox != nil {
		return stockBox
	}

	// Check for Person box
	if personBox := ExtractPersonBox(doc); personBox != nil {
		return personBox
	}

	// Check for featured snippet
	if featuredSnippet := ExtractFeaturedSnippet(doc); featuredSnippet != nil {
		return featuredSnippet
	}

	// If no specific box found, return a default empty box
	return &BingAnswerBox{
		Type:    "none",
		Content: "",
	}
}

// ExtractFeaturedSnippet extracts a featured snippet from Bing search results
func ExtractFeaturedSnippet(doc *goquery.Document) *BingAnswerBox {
	featuredSnippet := doc.Find(".b_featuredSnippet").First()
	if featuredSnippet.Length() == 0 {
		return nil
	}

	title := featuredSnippet.Find("h2").Text()
	content := featuredSnippet.Find(".b_caption").Text()

	// Clean and trim the content
	content = strings.TrimSpace(content)

	return &BingAnswerBox{
		Type:    "featured_snippet",
		Title:   title,
		Content: content,
	}
}

// GetTextContent extracts all text from a selection and formats it
func GetTextContent(s *goquery.Selection) string {
	var content strings.Builder

	// Process each text node and add appropriate spacing
	var prev *goquery.Selection
	s.Contents().Each(func(i int, sel *goquery.Selection) {
		if goquery.NodeName(sel) == "#text" {
			text := strings.TrimSpace(sel.Text())
			if text != "" {
				if i > 0 && prev != nil && goquery.NodeName(prev) != "br" {
					content.WriteString(" ")
				}
				content.WriteString(text)
			}
		} else if goquery.NodeName(sel) == "br" {
			content.WriteString("\n")
		} else {
			// Recursively get text from child elements
			childText := GetTextContent(sel)
			if childText != "" {
				if content.Len() > 0 && !strings.HasSuffix(content.String(), "\n") {
					content.WriteString(" ")
				}
				content.WriteString(childText)
			}
		}
		prev = sel
	})

	return strings.TrimSpace(content.String())
}
