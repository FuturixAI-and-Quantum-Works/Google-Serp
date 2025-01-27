package search

import (
	"compress/flate"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/andybalholm/brotli"
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

	cookie := GetRandomCookie()
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:134.0) Gecko/20100101 Firefox/134.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br, zstd")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Priority", "u=0, i")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Cookie", cookie)
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("TE", "trailers")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
	}

	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %v", err)
		}
		defer reader.Close()
	case "deflate":
		reader = flate.NewReader(resp.Body)
		defer reader.Close()
	case "br":
		reader = io.NopCloser(brotli.NewReader(resp.Body))
	default:
		reader = resp.Body
	}

	body, err := io.ReadAll(reader) // write to a html file
	ioutil.WriteFile("images.html", body, 0644)

	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %v", err)
	}

	var imageInfos []ImageInfo
	doc.Find("img").Each(func(i int, s *goquery.Selection) {
		title := s.AttrOr("alt", "")

		url, exists := s.Attr("src")
		if !exists {
			url, exists = s.Attr("data-src")
			if !exists {
				return
			}
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
