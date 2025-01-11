package search

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/mux"
)

type ImageInfo struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type ImageConfig struct {
	Query string
}

// ImageScraper handles the scraping functionality
type ImageScraper struct {
	client *http.Client
	config ImageConfig
}

// NewImageScraper creates a new scraper instance
func NewImageScraper(config ImageConfig) *ImageScraper {
	return &ImageScraper{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		config: config,
	}
}

// buildImageURL creates the image URL with parameters
func (s *ImageScraper) buildImageURL(query string) string {
	return fmt.Sprintf("https://www.google.com/search?q=%s&tbm=isch", query)
}

func (s *ImageScraper) ImageScrape() ([]ImageInfo, error) {
	req, err := http.NewRequest("GET", s.buildImageURL(s.config.Query), nil)
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

	// write to a html file
	ioutil.WriteFile("images.html", body, 0644)

	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %v", err)
	}

	var imageInfos []ImageInfo
	doc.Find("img.rg_i").Each(func(i int, s *goquery.Selection) {
		println(s)
		title := s.AttrOr("alt", "")

		url, exists := s.Attr("src")
		if !exists {
			return
		}

		imageInfos = append(imageInfos, ImageInfo{
			Title: title,
			URL:   url,
		})
	})

	println(len(imageInfos))
	return imageInfos, nil
}

// StandardImageHandler handles image queries
func StandardImageHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	query := vars["query"]
	println(query)
	if query == "" {
		http.Error(w, "Query parameter is required", http.StatusBadRequest)
		return
	}

	config := ImageConfig{
		Query: query,
	}

	scraper := NewImageScraper(config)

	imageInfos, err := scraper.ImageScrape()

	if err != nil {
		http.Error(w, "Error scraping results", http.StatusInternalServerError)
		return
	}

	jsonData, err := json.MarshalIndent(imageInfos, "", "    ")
	if err != nil {
		http.Error(w, "Error marshaling to JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}
