// Package scraper provides a modular system for scraping web content
package scraper

import (
	"net/url"
	"strings"
	"sync"

	"github.com/PuerkitoBio/goquery"
)

// ScrapedContent represents the extracted content in Markdown format
type ScrapedContent struct {
	Markdown string `json:"markdown"` // The main content in Markdown format
}

// Scraper interface defines the methods required for a webpage scraper
type Scraper interface {
	// CanHandle determines if this scraper can handle the given URL
	CanHandle(url string) bool

	// Scrape extracts content from the HTML document and converts to Markdown
	Scrape(doc *goquery.Document, url string) (*ScrapedContent, error)
}

// Registry manages the available scrapers
type Registry struct {
	scrapers []Scraper
	fallback Scraper
	mu       sync.RWMutex
}

// NewRegistry creates a new scraper registry
func NewRegistry() *Registry {
	return &Registry{
		scrapers: make([]Scraper, 0),
	}
}

// Register adds a new scraper to the registry
func (r *Registry) Register(scraper Scraper) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.scrapers = append(r.scrapers, scraper)
}

// SetFallback sets the fallback scraper to use when no other scraper matches
func (r *Registry) SetFallback(scraper Scraper) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.fallback = scraper
}

// FindScraper returns the appropriate scraper for the given URL
func (r *Registry) FindScraper(urlStr string) Scraper {
	r.mu.RLock()
	defer r.mu.RUnlock()

	for _, scraper := range r.scrapers {
		if scraper.CanHandle(urlStr) {
			return scraper
		}
	}

	return r.fallback
}

// ExtractDomain extracts the domain from a URL string
func ExtractDomain(urlStr string) string {
	parsedURL, err := url.Parse(urlStr)
	if err != nil {
		return ""
	}

	host := parsedURL.Hostname()
	parts := strings.Split(host, ".")

	// Handle cases like domain.co.uk
	if len(parts) > 2 && len(parts[len(parts)-2]) <= 3 && len(parts[len(parts)-1]) <= 3 {
		if len(parts) >= 3 {
			return parts[len(parts)-3] + "." + parts[len(parts)-2] + "." + parts[len(parts)-1]
		}
	}

	// Regular domain like example.com
	if len(parts) >= 2 {
		return parts[len(parts)-2] + "." + parts[len(parts)-1]
	}

	return host
}

// CleanText removes extra whitespace from text
func CleanText(text string) string {
	// Replace newlines and tabs with spaces
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\t", " ")

	// Replace multiple spaces with a single space
	for strings.Contains(text, "  ") {
		text = strings.ReplaceAll(text, "  ", " ")
	}

	return strings.TrimSpace(text)
}

// DefaultRegistry is the global scraper registry
var DefaultRegistry = NewRegistry()
