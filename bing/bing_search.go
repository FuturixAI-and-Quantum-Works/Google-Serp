package bing_scrapper

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/mux"
)

// SearchResult represents a single search result
type SearchResult struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	URL     string `json:"url"`
	Favicon string `json:"favicon"`
}

// SearchResponse represents the complete search response
type SearchResponse struct {
	Links             []SearchResult     `json:"links,omitempty"`
	AnswerBox         *AnswerBox         `json:"answer_box,omitempty"`
	SuggestedProducts []SuggestedProduct `json:"suggested_products,omitempty"`
}

// SuggestedProduct represents a suggested product
type SuggestedProduct struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

// AnswerBox represents an answer box
type AnswerBox struct {
	Type        string      `json:",omitempty"`
	Content     interface{} `json:",omitempty"`
	RelatedText string      `json:",omitempty"`
	Source      string      `json:",omitempty"`
	SourceURL   string      `json:",omitempty"`
}

// SearchConfig holds the search parameters
type SearchConfig struct {
	Query      string
	Location   string
	Language   string
	MaxResults int
	Latitude   *float64 // Optional latitude
	Longitude  *float64 // Optional longitude
}

// SearchScraper handles the scraping functionality
type SearchScraper struct {
	client *http.Client
	config SearchConfig
}

// NewSearchScraper creates a new scraper instance
func NewSearchScraper(config SearchConfig) *SearchScraper {
	return &SearchScraper{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		config: config,
	}
}

// buildSearchURL creates the search URL with parameters
func (s *SearchScraper) buildSearchURL() string {
	params := url.Values{}
	params.Add("q", s.config.Query)

	if s.config.Location != "" {
		params.Add("setmkt", s.config.Language)
	}

	if s.config.Latitude != nil && s.config.Longitude != nil {
		params.Add("setlang", fmt.Sprintf("%f,%f", *s.config.Latitude, *s.config.Longitude))
	}

	return "https://www.bing.com/search?" + params.Encode()
}

func (s *SearchScraper) Scrape() (*SearchResponse, error) {
	req, err := http.NewRequest("GET", s.buildSearchURL(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
	}

	body, err := ioutil.ReadAll(resp.Body)

	// write to a file
	ioutil.WriteFile("bing.html", body, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %v", err)
	}

	searchResponse := &SearchResponse{
		Links:             []SearchResult{},
		AnswerBox:         nil,
		SuggestedProducts: []SuggestedProduct{},
	}

	extractedResults := ExtractSearchResults(doc, s.config.MaxResults)
	searchResponse.Links = extractedResults

	// Extract answer box
	answerBox := ExtractAnswerbox(doc)
	if answerBox != nil {
		searchResponse.AnswerBox = answerBox
	}

	suggestedProducts := ExtractSuggestedProducts(doc)
	searchResponse.SuggestedProducts = suggestedProducts

	return searchResponse, nil
}

func StandardSearchHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	query := vars["query"]
	location := vars["location"]
	maxResults, err := strconv.Atoi(vars["maxResults"])
	if err != nil {
		http.Error(w, "Invalid maxResults parameter", http.StatusBadRequest)
		return
	}

	lat, err := strconv.ParseFloat(vars["latitude"], 64)
	if err != nil {
		http.Error(w, "Invalid latitude parameter", http.StatusBadRequest)
		return
	}

	lon, err := strconv.ParseFloat(vars["longitude"], 64)
	if err != nil {
		http.Error(w, "Invalid longitude parameter", http.StatusBadRequest)
		return
	}

	useCoords := vars["useCoords"] == "true"

	if query == "" {
		http.Error(w, "Query parameter is required", http.StatusBadRequest)
		return
	}

	config := SearchConfig{
		Query:      query,
		Location:   location,
		MaxResults: maxResults,
	}

	if useCoords {
		config.Latitude = &lat
		config.Longitude = &lon
	}

	scraper := NewSearchScraper(config)

	searchResponse, err := scraper.Scrape()
	if err != nil {
		fmt.Println(err.Error())
		http.Error(w, "Error scraping results", http.StatusInternalServerError)
		return
	}

	jsonData, err := json.MarshalIndent(searchResponse, "", "    ")
	if err != nil {
		http.Error(w, "Error marshaling to JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}

func ExtractSearchResults(doc *goquery.Document, maxResults int) []SearchResult {
	results := []SearchResult{}

	doc.Find(".b_algo").Each(func(i int, s *goquery.Selection) {
		if i >= maxResults {
			return
		}

		title := s.Find(".pr").Text()
		content := s.Find(".sn").Text()
		url := s.Find("a").AttrOr("href", "")

		results = append(results, SearchResult{
			Title:   title,
			Content: content,
			URL:     url,
			Favicon: "",
		})
	})

	return results
}

func ExtractAnswerbox(doc *goquery.Document) *AnswerBox {
	// Implement answer box extraction logic for Bing
	return nil
}

func ExtractSuggestedProducts(doc *goquery.Document) []SuggestedProduct {
	// Implement suggested products extraction logic for Bing
	return []SuggestedProduct{}
}
