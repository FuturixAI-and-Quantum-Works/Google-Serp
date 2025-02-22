package search

import (
	"context"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	bingsearch "googlescrapper/bing_search"

	"github.com/PuerkitoBio/goquery"
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

func (s *BingScraper) BingScrape() (BingInfo, error) {
	// Create a new headless browser context using chromedp
	ctx, cancel := chromedp.NewContext(context.Background())
	defer cancel()

	searchURL := s.buildBingURL(s.config.Query)
	var htmlContent string

	// Navigate to the search URL and wait until a key element appears
	err := chromedp.Run(ctx,
		chromedp.Navigate(searchURL),
		// Wait for at least one search result to be visible
		chromedp.WaitVisible(`li.b_algo`, chromedp.ByQuery),
		// Wait for 3 seconds
		chromedp.Sleep(2*time.Second),
		// Extract the full HTML of the page
		chromedp.OuterHTML(`html`, &htmlContent, chromedp.ByQuery),
	)
	if err != nil {
		return BingInfo{}, fmt.Errorf("failed to retrieve page content: %v", err)
	}

	// Optionally write the HTML to file for debugging
	ioutil.WriteFile("bing.html", []byte(htmlContent), 0644)

	// Parse the retrieved HTML with goquery
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(htmlContent))
	if err != nil {
		return BingInfo{}, fmt.Errorf("failed to parse HTML: %v", err)
	}

	var BingLinks []BingLink
	var BingInfos BingInfo
	var wg sync.WaitGroup
	var mu sync.Mutex
	doc.Find("li.b_algo").Each(func(i int, s *goquery.Selection) {
		// Extract the title and link from the <h2> anchor.
		title := s.Find("h2 a").Text()
		link, exists := s.Find("h2 a").Attr("href")
		if !exists {
			println("Link does not exist")
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

		wg.Add(1)
		go func(title, link, websiteName, websiteAttribution, caption string, tags []string) {
			defer wg.Done()

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
		}(title, link, websiteName, websiteAttribution, caption, tags)
	})

	wg.Wait()
	AnswerBox := bingsearch.ExtractAnswerbox(doc)

	BingInfos.Links = BingLinks
	BingInfos.AnswerBox = *AnswerBox
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
