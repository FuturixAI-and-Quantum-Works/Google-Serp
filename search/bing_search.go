// Package search provides search engine scraping functionality
package search

import (
	"context"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"googlescrapper/bingsearch"
	"googlescrapper/browser"
	"googlescrapper/cache"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/gorilla/mux"
)

// BingLink represents a single search result link from Bing
type BingLink struct {
	Title              string   `json:"title"`
	URL                string   `json:"url"`
	WebsiteName        string   `json:"websiteName"`
	WebsiteAttribution string   `json:"websiteAttribution"`
	Tags               []string `json:"tags"`
	Caption            string   `json:"caption"`
}

// BingInfo represents the complete search results from Bing
type BingInfo struct {
	Links     []BingLink               `json:"links"`
	AnswerBox bingsearch.BingAnswerBox `json:"answer_box"`
}

// BingConfig holds configuration for Bing searches
type BingConfig struct {
	Query string
}

// BingScraper handles the scraping functionality for Bing search
type BingScraper struct {
	config BingConfig
}

// NewBingScraper creates a new Bing scraper instance
func NewBingScraper(config BingConfig) *BingScraper {
	return &BingScraper{
		config: config,
	}
}

// buildBingURL creates a Bing search URL for the given query
func (s *BingScraper) buildBingURL(query string) string {
	return fmt.Sprintf("https://www.bing.com/search?q=%s", url.QueryEscape(query))
}

// generateCacheKey creates a unique key for caching based on the query
func (s *BingScraper) generateCacheKey() string {
	// Create a hash of the query for a consistent cache key
	hash := md5.Sum([]byte(s.config.Query))
	return "bing_search:" + hex.EncodeToString(hash[:])
}

// BingScrape performs a Bing search and returns the results
func (s *BingScraper) BingScrape() (BingInfo, error) {
	// Check if the query is related to time or weather - these shouldn't be cached
	lowerQuery := strings.ToLower(s.config.Query)
	timeWeatherPatterns := []string{
		"time", "clock", "hour", "minute", "current time", "local time", "what time",
		"weather", "temperature", "forecast", "rain", "snow", "humidity", "climate",
		"how hot", "how cold", "degrees", "celsius", "fahrenheit",
	}

	// Check if query matches any of the exclusion patterns
	isTimeWeatherQuery := false
	for _, pattern := range timeWeatherPatterns {
		if strings.Contains(lowerQuery, pattern) {
			isTimeWeatherQuery = true
			break
		}
	}

	// For time/weather queries, bypass cache and fetch directly
	if isTimeWeatherQuery {
		return s.fetchBingResults()
	}

	// For other queries, use cache
	cacheKey := s.generateCacheKey()
	cacheTTL := 1 * time.Hour // Cache results for 1 hour

	result, err := cache.Memoize(cacheKey, cacheTTL, func() (BingInfo, error) {
		// This is the original function that will be called if cache misses
		return s.fetchBingResults()
	})

	return result, err
}

// fetchBingResults performs the actual scraping of Bing search results
func (s *BingScraper) fetchBingResults() (BingInfo, error) {
	// Get a browser context from the pool
	ctx, returnCtx, err := browser.DefaultPool.GetContext()
	if err != nil {
		return BingInfo{}, fmt.Errorf("failed to get browser context: %v", err)
	}
	defer returnCtx() // Return the context to the pool when done

	searchURL := s.buildBingURL(s.config.Query)
	var htmlContent string

	// Add a timeout for this specific operation
	timeoutCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
	defer cancel()

	// Navigate to the search URL and scrape the content
	err = chromedp.Run(timeoutCtx,
		// Set custom headers for this request
		chromedp.ActionFunc(func(ctx context.Context) error {
			return network.SetExtraHTTPHeaders(map[string]interface{}{
				"accept":                    "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
				"accept-language":           "en-US,en;q=0.9",
				"cache-control":             "no-cache",
				"pragma":                    "no-cache",
				"sec-ch-ua":                 "\"Chromium\";v=\"135\", \"Not-A.Brand\";v=\"8\"",
				"sec-ch-ua-mobile":          "?0",
				"sec-ch-ua-platform":        "\"Linux\"",
				"sec-fetch-dest":            "document",
				"sec-fetch-mode":            "navigate",
				"sec-fetch-site":            "same-origin",
				"sec-fetch-user":            "?1",
				"upgrade-insecure-requests": "1",
			}).Do(ctx)
		}),
		// Clear cookies to avoid personalization
		network.ClearBrowserCookies(),
		// Navigate to the search URL
		chromedp.Navigate(searchURL),
		// Wait for results to appear
		chromedp.WaitVisible(`li.b_algo`, chromedp.ByQuery),
		// Extract the full HTML of the page
		chromedp.OuterHTML(`html`, &htmlContent, chromedp.ByQuery),
	)

	if err != nil {
		return BingInfo{}, fmt.Errorf("failed to scrape content: %v", err)
	}

	// Optionally write the HTML to file for debugging
	if err := ioutil.WriteFile("bing.html", []byte(htmlContent), 0644); err != nil {
		// Just log error but continue with processing
		fmt.Printf("Warning: Failed to write debug file: %v\n", err)
	}

	// Parse the retrieved HTML with goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return BingInfo{}, fmt.Errorf("failed to parse HTML: %v", err)
	}

	var BingLinks []BingLink
	var BingInfos BingInfo
	var wg sync.WaitGroup
	var mu sync.Mutex

	// Use a worker pool for processing results
	const maxWorkers = 10
	semaphore := make(chan struct{}, maxWorkers)

	doc.Find("li.b_algo").Each(func(i int, s *goquery.Selection) {
		semaphore <- struct{}{} // Acquire token
		wg.Add(1)

		go func(i int, s *goquery.Selection) {
			defer func() {
				<-semaphore // Release token
				wg.Done()
			}()

			// Extract the title and link from the <h2> anchor.
			title := s.Find("h2 a").Text()
			link, exists := s.Find("h2 a").Attr("href")
			if !exists {
				return
			}

			// Extract website name and attribution from the "b_tpcn" section.
			websiteName := s.Find("div.b_tpcn .tptt").Text()
			websiteAttribution := s.Find("div.b_tpcn .b_attribution cite").Text()

			// Extract the caption from the <p> element with class "b_lineclamp2".
			caption := s.Find("p.b_lineclamp2").Text()

			// Optionally extract tags if available.
			var tags []string
			s.Find(".tltg").Each(func(i int, tag *goquery.Selection) {
				tags = append(tags, tag.Text())
			})

			mu.Lock()
			BingLinks = append(BingLinks, BingLink{
				Title:              title,
				URL:                link,
				WebsiteName:        websiteName,
				WebsiteAttribution: websiteAttribution,
				Tags:               tags,
				Caption:            caption,
			})
			mu.Unlock()
		}(i, s)
	})

	wg.Wait()

	// Process answer box concurrently
	answerBoxCh := make(chan *bingsearch.BingAnswerBox, 1)
	go func() {
		answerBoxCh <- bingsearch.ExtractAnswerbox(doc)
	}()

	BingInfos.Links = BingLinks

	// Check if the answer box type is "none"
	answerBox := <-answerBoxCh
	if answerBox != nil && answerBox.Type != "none" && answerBox.Type != "" {
		BingInfos.AnswerBox = *answerBox
	} else {
		// Use empty struct if type is none or empty
		BingInfos.AnswerBox = bingsearch.BingAnswerBox{}
	}

	return BingInfos, nil
}

// getHTML fetches the HTML content of a given URL
func getHTML(url string) (string, error) {
	return browser.DefaultPool.FetchURL(url, 15*time.Second)
}

// GetHTMLFromUrl handles HTTP requests to get HTML content from a URL
func GetHTMLFromUrl(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		URL string `json:"url"`
	}

	// Parse the JSON body
	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if requestBody.URL == "" {
		http.Error(w, "URL parameter is required in the request body", http.StatusBadRequest)
		return
	}

	htmlContent, err := getHTML(requestBody.URL)
	if err != nil {
		http.Error(w, "Error scraping results: "+err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(htmlContent))
}

// StandardBingHandler handles HTTP requests for Bing searches
func StandardBingHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	query := vars["query"]
	if query == "" {
		http.Error(w, "Query parameter is required", http.StatusBadRequest)
		return
	}

	config := BingConfig{
		Query: query,
	}

	scraper := NewBingScraper(config)
	BingInfos, err := scraper.BingScrape()
	if err != nil {
		http.Error(w, "Error scraping results: "+err.Error(), http.StatusInternalServerError)
		return
	}

	jsonData, err := json.MarshalIndent(BingInfos, "", "    ")
	if err != nil {
		http.Error(w, "Error marshaling to JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}
