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
	"github.com/chromedp/cdproto/page"
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

// BrowserPool manages a pool of browser contexts for reuse
type BrowserPool struct {
	contexts      chan context.Context
	cancelFuncs   map[context.Context]context.CancelFunc
	initOnce      sync.Once
	minSize       int
	maxSize       int
	currentSize   int
	mu            sync.Mutex
	allocCtx      context.Context
	allocCancel   context.CancelFunc
	initialized   bool
	scaleUpCount  int       // Tracks consecutive scale up events
	scaleDownTime time.Time // Last time we scaled down
	waitQueue     int       // Count of waiting requests
}

var (
	// Global browser pool with auto-scaling
	globalBrowserPool = &BrowserPool{
		minSize:     3,
		maxSize:     20,
		currentSize: 0,
		contexts:    make(chan context.Context, 20), // Buffer up to max size
		cancelFuncs: make(map[context.Context]context.CancelFunc),
	}
)

// Initialize creates the browser pool
func (pool *BrowserPool) Initialize() {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	if pool.initialized {
		return
	}

	// Create a shared allocator context with options
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

	pool.allocCtx, pool.allocCancel = chromedp.NewExecAllocator(context.Background(), opts...)

	// Create initial set of browser contexts
	pool.scaleUp(pool.minSize)

	// Start the auto-scaler
	go pool.autoScaler()

	pool.initialized = true
	fmt.Printf("Browser pool initialized with %d browsers (min: %d, max: %d)\n",
		pool.currentSize, pool.minSize, pool.maxSize)
}

// scaleUp adds n browser instances to the pool
func (pool *BrowserPool) scaleUp(n int) {
	for i := 0; i < n; i++ {
		ctx, cancel := chromedp.NewContext(pool.allocCtx, chromedp.WithLogf(func(format string, args ...interface{}) {
			// Silent logging
		}))

		// Add event listeners to handle the events that need handling
		chromedp.ListenTarget(ctx, func(ev interface{}) {
			// Silently handle the EventFrameStartedNavigating event
			switch ev.(type) {
			case *page.EventFrameStartedNavigating:
				// Just log or silently ignore, depending on your needs
				// fmt.Printf("Frame started navigating\n")
			}
		})

		// Initialize the browser in advance
		if err := chromedp.Run(ctx, chromedp.Navigate("about:blank")); err != nil {
			fmt.Printf("Error initializing browser: %v\n", err)
			cancel()
			continue
		}

		pool.contexts <- ctx
		pool.cancelFuncs[ctx] = cancel
		pool.currentSize++
	}

	if n > 0 {
		fmt.Printf("Scaled up pool by %d browsers to %d total\n", n, pool.currentSize)
	}
}

// scaleDown removes n browser instances from the pool
func (pool *BrowserPool) scaleDown(n int) {
	if pool.currentSize <= pool.minSize {
		return // Don't scale below minimum
	}

	// Limit scale down to not go below minSize
	if pool.currentSize-n < pool.minSize {
		n = pool.currentSize - pool.minSize
	}

	if n <= 0 {
		return
	}

	for i := 0; i < n; i++ {
		select {
		case ctx := <-pool.contexts:
			// Get the cancel function
			if cancel, exists := pool.cancelFuncs[ctx]; exists {
				cancel()
				delete(pool.cancelFuncs, ctx)
				pool.currentSize--
			}
		default:
			// No contexts available to scale down
			return
		}
	}

	pool.scaleDownTime = time.Now()
	fmt.Printf("Scaled down pool by %d browsers to %d total\n", n, pool.currentSize)
}

// autoScaler periodically checks if the pool needs to be resized
func (pool *BrowserPool) autoScaler() {
	ticker := time.NewTicker(10 * time.Second)
	defer ticker.Stop()

	for {
		<-ticker.C

		pool.mu.Lock()

		// Get metrics
		poolSize := pool.currentSize
		availableBrowsers := len(pool.contexts)
		waitingRequests := pool.waitQueue

		// Calculate utilization percentage (0-100)
		utilization := 0
		if poolSize > 0 {
			utilization = 100 * (poolSize - availableBrowsers) / poolSize
		}

		// Scale up logic
		if (utilization > 80 || waitingRequests > 0) && poolSize < pool.maxSize {
			// Calculate how many browsers to add
			toAdd := 1
			if waitingRequests > 1 {
				// Add more browsers if there are multiple waiting requests
				toAdd = min(waitingRequests, pool.maxSize-poolSize)
			}

			pool.scaleUpCount++
			// If we've needed to scale up multiple times in succession, add more browsers
			if pool.scaleUpCount > 3 {
				toAdd = min(toAdd*2, pool.maxSize-poolSize)
			}

			pool.scaleUp(toAdd)
		} else {
			pool.scaleUpCount = 0
		}

		// Scale down logic - only if utilization is low and we haven't scaled down recently
		cooldownPeriod := 2 * time.Minute
		if utilization < 30 && poolSize > pool.minSize && time.Since(pool.scaleDownTime) > cooldownPeriod {
			// Calculate how many browsers to remove
			excessBrowsers := availableBrowsers - max(1, poolSize/5) // Keep at least 20% capacity as buffer
			toRemove := min(excessBrowsers, poolSize-pool.minSize)

			if toRemove > 0 {
				pool.scaleDown(toRemove)
			}
		}

		pool.mu.Unlock()
	}
}

// min returns the smaller of two integers
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}

// max returns the larger of two integers
func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

// GetContext gets a browser context from the pool
func (pool *BrowserPool) GetContext() (context.Context, context.CancelFunc, error) {
	pool.initOnce.Do(func() {
		pool.Initialize()
	})

	pool.mu.Lock()
	pool.waitQueue++
	pool.mu.Unlock()

	// Try to get a context immediately or wait up to 500ms
	select {
	case ctx := <-pool.contexts:
		pool.mu.Lock()
		pool.waitQueue--
		pool.mu.Unlock()

		// Create a return function that puts the context back in the pool
		returnCtx := func() {
			// Refresh the browser before returning to pool
			refreshCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
			defer cancel()

			// Navigate to blank page to clear state and reduce memory
			_ = chromedp.Run(refreshCtx,
				network.ClearBrowserCookies(),
				chromedp.Navigate("about:blank"),
			)

			select {
			case pool.contexts <- ctx:
			default:
				// Pool channel is full, context will be GC'd
				if cancel, exists := pool.cancelFuncs[ctx]; exists {
					pool.mu.Lock()
					cancel()
					delete(pool.cancelFuncs, ctx)
					pool.currentSize--
					pool.mu.Unlock()
				}
			}
		}

		return ctx, returnCtx, nil

	case <-time.After(500 * time.Millisecond):
		// If we waited more than 500ms, try to create a new browser instance
		pool.mu.Lock()

		// Only create a new instance if we're below max capacity
		if pool.currentSize < pool.maxSize {
			ctx, cancel := chromedp.NewContext(pool.allocCtx, chromedp.WithLogf(func(format string, args ...interface{}) {
				// Silent logging
			}))

			// Try to initialize the browser quickly
			initCtx, initCancel := context.WithTimeout(context.Background(), 5*time.Second)
			err := chromedp.Run(initCtx, chromedp.Navigate("about:blank"))
			initCancel()

			if err != nil {
				cancel()
				pool.waitQueue--
				pool.mu.Unlock()
				return nil, nil, fmt.Errorf("failed to create new browser instance: %v", err)
			}

			// Add to pool management
			pool.cancelFuncs[ctx] = cancel
			pool.currentSize++
			pool.waitQueue--
			pool.mu.Unlock()

			// Create return function
			returnCtx := func() {
				// Refresh before returning to pool
				refreshCtx, cancel := context.WithTimeout(ctx, 3*time.Second)
				defer cancel()

				_ = chromedp.Run(refreshCtx,
					network.ClearBrowserCookies(),
					chromedp.Navigate("about:blank"),
				)

				// Put context back in the pool
				select {
				case pool.contexts <- ctx:
					// Successfully returned to pool
				default:
					// Pool is full, this extra instance will be closed
					if cancel, exists := pool.cancelFuncs[ctx]; exists {
						pool.mu.Lock()
						cancel()
						delete(pool.cancelFuncs, ctx)
						pool.currentSize--
						pool.mu.Unlock()
					}
				}
			}

			return ctx, returnCtx, nil
		}

		// If we couldn't create a new instance, wait longer for an existing one
		pool.mu.Unlock()

		// Wait up to 3 more seconds for a context to become available
		select {
		case ctx := <-pool.contexts:
			pool.mu.Lock()
			pool.waitQueue--
			pool.mu.Unlock()

			returnCtx := func() {
				select {
				case pool.contexts <- ctx:
					// Successfully returned to pool
				default:
					// Pool is full
					fmt.Println("Browser pool is full, cannot return context")
				}
			}

			return ctx, returnCtx, nil
		case <-time.After(3 * time.Second):
			pool.mu.Lock()
			pool.waitQueue--
			pool.mu.Unlock()
			return nil, nil, fmt.Errorf("timeout getting browser context from pool")
		}
	}
}

// Shutdown closes all browser instances
func (pool *BrowserPool) Shutdown() {
	pool.mu.Lock()
	defer pool.mu.Unlock()

	if !pool.initialized {
		return
	}

	// Cancel all contexts
	for ctx, cancel := range pool.cancelFuncs {
		cancel()
		delete(pool.cancelFuncs, ctx)
	}

	// Cancel the allocator
	if pool.allocCancel != nil {
		pool.allocCancel()
	}

	// Clear the channel
	for len(pool.contexts) > 0 {
		<-pool.contexts
	}

	pool.currentSize = 0
	pool.initialized = false
	fmt.Println("Browser pool shut down")
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

// fetchBingResults contains the browser-based implementation
func (s *BingScraper) fetchBingResults() (BingInfo, error) {
	// Get a browser context from the pool
	ctx, returnCtx, err := globalBrowserPool.GetContext()
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
	BingInfos.AnswerBox = *<-answerBoxCh

	return BingInfos, nil
}

// function to get html of a page
func getHTML(url string) (string, error) {
	ctx, returnCtx, err := globalBrowserPool.GetContext()

	if err != nil {
		return "", fmt.Errorf("failed to get browser context: %v", err)
	}

	defer returnCtx() // Return the context to the pool when done

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
		chromedp.Navigate(url),
		// Wait for p elements to appear with a deadline of 200ms
		chromedp.ActionFunc(func(ctx context.Context) error {
			deadlineCtx, cancel := context.WithTimeout(ctx, 800*time.Millisecond)
			defer cancel()
			return chromedp.Run(deadlineCtx, chromedp.WaitVisible(`p`, chromedp.ByQuery))
		}),
		// Extract the full HTML of the page
		chromedp.OuterHTML(`html`, &htmlContent, chromedp.ByQuery),
	)
	if err != nil {
		return "", fmt.Errorf("failed to scrape content: %v", err)
	}

	// return the HTML content
	return htmlContent, nil
}

func GetHTMLFromUrl(w http.ResponseWriter, r *http.Request) {
	var requestBody struct {
		URL string `json:"url"`
	}

	// Parse the JSON body
	println("Parsing JSON body")

	if err := json.NewDecoder(r.Body).Decode(&requestBody); err != nil {
		println("Error decoding JSON body:", err)
		http.Error(w, "Invalid JSON body", http.StatusBadRequest)
		return
	}

	if requestBody.URL == "" {
		println("Error: URL parameter is required")
		http.Error(w, "URL parameter is required in the request body", http.StatusBadRequest)
		return
	}

	htmlContent, err := getHTML(requestBody.URL)
	if err != nil {
		println("Error scraping results:", err.Error())
		http.Error(w, "Error scraping results", http.StatusInternalServerError)
		return
	}

	println(htmlContent)
	w.Header().Set("Content-Type", "text/plain")
	w.Write([]byte(htmlContent))
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

// Initialize browser pool at package level
func init() {
	// Initialize the browser pool in a background goroutine
	go globalBrowserPool.Initialize()
}
