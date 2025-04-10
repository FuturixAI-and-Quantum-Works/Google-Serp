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

	bingsearch "googlescrapper/bing_search"
	"googlescrapper/cache"

	"github.com/PuerkitoBio/goquery"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"
	"github.com/gorilla/mux"
)

type BingLink struct {
	Title              string   `json:"title"`
	URL                string   `json:"url"`
	WebsiteName        string   `json:"websiteName"`
	WebsiteAttribution string   `json:"websiteAttribution"`
	Tags               []string `json:"tags"`
	Caption            string   `json:"caption"`
}
type BingInfo struct {
	Links     []BingLink               `json:"links"`
	AnswerBox bingsearch.BingAnswerBox `json:"answer_box"`
}

type BingConfig struct {
	Query string
}

// BingScraper handles the scraping functionality
type BingScraper struct {
	config BingConfig
}

// NewBingScraper creates a new scraper instance
func NewBingScraper(config BingConfig) *BingScraper {
	return &BingScraper{
		config: config,
	}
}

func (s *BingScraper) buildBingURL(query string) string {
	return fmt.Sprintf("https://www.bing.com/search?q=%s", url.QueryEscape(query))
}

// generateCacheKey creates a unique key for caching based on the query
func (s *BingScraper) generateCacheKey() string {
	// Create a hash of the query for a consistent cache key
	hash := md5.Sum([]byte(s.config.Query))
	return "bing_search:" + hex.EncodeToString(hash[:])
}

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

// fetchBingResults contains the original implementation of BingScrape
func (s *BingScraper) fetchBingResults() (BingInfo, error) {
	// Create optimized browser options
	opts := append(chromedp.DefaultExecAllocatorOptions[:],
		chromedp.Flag("headless", true),
		chromedp.Flag("disable-gpu", true),
		chromedp.Flag("disable-extensions", true),
		chromedp.Flag("disable-dev-shm-usage", true),
		chromedp.Flag("no-sandbox", true),
		chromedp.Flag("disable-setuid-sandbox", true),
		chromedp.Flag("disable-web-security", true),
		chromedp.Flag("disable-features", "IsolateOrigins,site-per-process"),
		chromedp.Flag("disable-site-isolation-trials", true),
		chromedp.Flag("ignore-certificate-errors", true),
		chromedp.WindowSize(1920, 1080),
		chromedp.UserAgent("Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/135.0.0.0 Safari/537.36"),
	)

	// Create browser context with timeout
	allocCtx, cancel := chromedp.NewExecAllocator(context.Background(), opts...)
	defer cancel()

	// Add timeout to context
	ctx, cancel := context.WithTimeout(allocCtx, 15*time.Second)
	defer cancel()

	// Create browser context
	ctx, cancel = chromedp.NewContext(ctx, chromedp.WithLogf(func(format string, args ...interface{}) {
		// Silent logging to improve performance
	}))
	defer cancel()

	searchURL := s.buildBingURL(s.config.Query)
	var htmlContent string

	// Navigate to the search URL and wait until key elements appear
	err := chromedp.Run(ctx,
		// Set custom headers
		chromedp.ActionFunc(func(ctx context.Context) error {
			// Set cookies and headers directly through CDP
			err := network.SetExtraHTTPHeaders(map[string]interface{}{
				"accept":             "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,image/apng,*/*;q=0.8,application/signed-exchange;v=b3;q=0.7",
				"accept-language":    "en-US,en;q=0.9",
				"cache-control":      "no-cache",
				"pragma":             "no-cache",
				"sec-ch-ua":          "\"Chromium\";v=\"135\", \"Not-A.Brand\";v=\"8\"",
				"sec-ch-ua-mobile":   "?0",
				"sec-ch-ua-platform": "\"Linux\"",
				"sec-fetch-dest":     "document",
				"sec-fetch-mode":     "navigate",
				"sec-fetch-site":     "same-origin",
			}).Do(ctx)
			return err
		}),
		chromedp.Navigate(searchURL),
		// Combination of wait conditions for faster results
		chromedp.WaitReady("body", chromedp.ByQuery),        // Wait for DOM to be ready
		chromedp.WaitVisible(`li.b_algo`, chromedp.ByQuery), // Wait for at least one search result
		// Reduced wait time - just enough for dynamic content loading
		// Extract the full HTML of the page
		chromedp.OuterHTML(`html`, &htmlContent, chromedp.ByQuery),
	)
	if err != nil {
		return BingInfo{}, fmt.Errorf("failed to retrieve page content: %v", err)
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
	BingInfos.AnswerBox = *<-answerBoxCh

	return BingInfos, nil
}

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
		http.Error(w, "Error scraping results", http.StatusInternalServerError)
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
