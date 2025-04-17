// Package scraper provides a modular system for scraping web content
package scraper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"

	"googlescrapper/browser"

	"github.com/PuerkitoBio/goquery"
)

// Service handles web content scraping operations
type Service struct {
	browserPool *browser.Pool
	registry    *Registry
}

// NewService creates a new scraper service
func NewService(pool *browser.Pool, registry *Registry) *Service {
	return &Service{
		browserPool: pool,
		registry:    registry,
	}
}

// DefaultService is the global scraper service
var DefaultService = NewService(browser.DefaultPool, DefaultRegistry)

// ScrapeURL fetches a URL and scrapes its content using the appropriate scraper
func (s *Service) ScrapeURL(urlStr string) (*ScrapedContent, error) {
	// Validate URL
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		urlStr = "https://" + urlStr
	}

	// Get HTML content using browser pool
	htmlContent, err := s.browserPool.FetchURL(urlStr, 15*time.Second)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL content: %v", err)
	}

	// Parse the HTML with goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %v", err)
	}

	// Find the appropriate scraper for this URL
	scraper := s.registry.FindScraper(urlStr)
	if scraper == nil {
		return nil, fmt.Errorf("no scraper available for URL: %s", urlStr)
	}

	// Use the scraper to extract structured content
	return scraper.Scrape(doc, urlStr)
}

// GetCleanHTML fetches a URL and returns the HTML with scripts, styles, and meta tags removed
func (s *Service) GetCleanHTML(urlStr string) (string, error) {
	// Validate URL
	if !strings.HasPrefix(urlStr, "http://") && !strings.HasPrefix(urlStr, "https://") {
		urlStr = "https://" + urlStr
	}
	
	// Get HTML content using browser pool
	htmlContent, err := s.browserPool.FetchURL(urlStr, 15*time.Second)
	if err != nil {
		return "", fmt.Errorf("failed to fetch URL content: %v", err)
	}
	
	// Parse the HTML with goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return "", fmt.Errorf("failed to parse HTML: %v", err)
	}
	
	// Check if a specialized scraper exists
	if scraper := s.registry.FindScraper(urlStr); scraper != nil && !IsGenericScraper(scraper) {
		// If a non-fallback scraper exists, return a message
		return "", fmt.Errorf("specialized scraper already exists for this URL")
	}
	
	// Remove unwanted elements
	CleanDocument(doc)
	
	// Get the HTML as a string
	cleanHTML, err := doc.Html()
	if err != nil {
		return "", fmt.Errorf("failed to get HTML: %v", err)
	}
	
	return cleanHTML, nil
}

// IsGenericScraper checks if a scraper is the fallback scraper
func IsGenericScraper(scraper Scraper) bool {
	_, isFallback := scraper.(*FallbackScraper)
	return isFallback
}

// CleanDocument removes scripts, styles, meta tags, and other unwanted elements from the document
func CleanDocument(doc *goquery.Document) {
	// Remove script tags
	doc.Find("script").Remove()
	
	// Remove style tags and inline styles
	doc.Find("style").Remove()
	doc.Find("[style]").RemoveAttr("style")
	
	// Remove meta tags
	doc.Find("meta").Remove()
	
	// Remove link tags (mainly for stylesheets)
	doc.Find("link").Remove()
	
	// Remove comments
	doc.Find("*").Contents().Each(func(i int, s *goquery.Selection) {
		if goquery.NodeName(s) == "#comment" {
			s.Remove()
		}
	})
	
	// Remove iframes
	doc.Find("iframe").Remove()
	
	// Remove noscript tags
	doc.Find("noscript").Remove()
	
	// Keep all class attributes for CSS selection
}

// ScrapeURLHandler is an HTTP handler for scraping URLs
func ScrapeURLHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse request body
	var requestBody struct {
		URL string `json:"url"`
	}

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if requestBody.URL == "" {
		http.Error(w, "URL parameter is required in the request body", http.StatusBadRequest)
		return
	}

	// Scrape the URL
	content, err := DefaultService.ScrapeURL(requestBody.URL)
	if err != nil {
		http.Error(w, "Error scraping URL: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// Return the scraped content as JSON with just the markdown field
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(content)
}

// GetCleanHTMLHandler is an HTTP handler for getting clean HTML from URLs
func GetCleanHTMLHandler(w http.ResponseWriter, r *http.Request) {
	// Only allow POST requests
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	
	// Parse request body
	var requestBody struct {
		URL string `json:"url"`
	}
	
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}
	
	if requestBody.URL == "" {
		http.Error(w, "URL parameter is required in the request body", http.StatusBadRequest)
		return
	}
	
	// Get clean HTML from the URL
	cleanHTML, err := DefaultService.GetCleanHTML(requestBody.URL)
	if err != nil {
		// If there's a specialized scraper, let the client know
		if strings.Contains(err.Error(), "specialized scraper already exists") {
			http.Error(w, err.Error(), http.StatusConflict) // 409 Conflict
			return
		}
		
		// Other errors
		http.Error(w, "Error getting HTML: "+err.Error(), http.StatusInternalServerError)
		return
	}
	
	// Return the clean HTML
	response := struct {
		HTML string `json:"html"`
		URL  string `json:"url"`
	}{
		HTML: cleanHTML,
		URL:  requestBody.URL,
	}
	
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(response)
}
