package search

import (
	"encoding/json"
	"fmt"
	"googlescrapper/config"
	"googlescrapper/standard_search"
	"googlescrapper/utils"
	"io"
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
	Links     []standard_search.SearchResult `json:"links,omitempty"`
	AnswerBox standard_search.AnswerBox      `json:"answer_box,omitempty"`
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

	if regionConfig, ok := config.RegionConfigs[s.config.Location]; ok {
		params.Add("gl", regionConfig.Gl)
		params.Add("lr", regionConfig.Lr)
		params.Add("hl", regionConfig.Hl)
	}

	// Add coordinates if both latitude and longitude are provided
	if s.config.Latitude != nil && s.config.Longitude != nil {
		// Add location bias parameter
		params.Add("geoloc", fmt.Sprintf("%f,%f", *s.config.Latitude, *s.config.Longitude))
		// Add additional location parameter used by Google
		params.Add("uule", utils.CreateUULE(*s.config.Latitude, *s.config.Longitude))
	}

	return "https://www.google.com/search?" + params.Encode()
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

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %v", err)
	}

	searchResponse := &SearchResponse{
		Links:     []standard_search.SearchResult{},
		AnswerBox: standard_search.AnswerBox{},
	}

	extractedResults := standard_search.ExtractSearchResults(doc, s.config.MaxResults)
	searchResponse.Links = extractedResults

	// Extract answer box
	answerBox := standard_search.ExtractAnswerbox(doc)
	if answerBox != nil {
		searchResponse.AnswerBox = *answerBox
	}

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

	// Validate region
	if _, ok := config.RegionConfigs[location]; !ok {
		http.Error(w, "Invalid region code", http.StatusBadRequest)
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
