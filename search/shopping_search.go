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

type ProductInfo struct {
	Title    string `json:"title,omitempty"`
	Price    string `json:"price,omitempty"`
	Link     string `json:"link,omitempty"`
	ImageURL string `json:"imageURL,omitempty"`
	Stars    string `json:"stars,omitempty"`
	Reviews  string `json:"reviews,omitempty"`
}

type ShoppingConfig struct {
	Query string
}

// ShoppingScraper handles the scraping functionality
type ShoppingScraper struct {
	client *http.Client
	config ShoppingConfig
}

// NewShoppingScraper creates a new scraper instance
func NewShoppingScraper(config ShoppingConfig) *ShoppingScraper {
	return &ShoppingScraper{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		config: config,
	}
}

// buildShoppingURL creates the shopping URL with parameters
func (s *ShoppingScraper) buildShoppingURL(query string) string {
	return fmt.Sprintf("https://www.google.com/search?q=%s&tbm=shop", query)
}

func (s *ShoppingScraper) ShoppingScrape() ([]ProductInfo, error) {
	req, err := http.NewRequest("GET", s.buildShoppingURL(s.config.Query), nil)
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
	ioutil.WriteFile("shopping.html", body, 0644)

	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %v", err)
	}

	var products []ProductInfo
	doc.Find(".sh-dgr__content").Each(func(i int, s *goquery.Selection) {
		title := s.Find(".tAxDx").Text()
		price := s.Find(".a8Pemb").Text()
		linkElement := s.Find(".mnIHsc a").First()
		rawLink, _ := linkElement.Attr("href")
		// Parse the actual URL from Google's redirect URL
		var link string
		if strings.HasPrefix(rawLink, "/url?url=") {
			link = strings.Split(strings.TrimPrefix(rawLink, "/url?url="), "&")[0]
		} else {
			link = rawLink
		}
		imageURL, _ := s.Find(".FM6uVc img").Attr("src")
		stars := s.Find(".Rsc7Yb").Text()
		// reviews := s.Find(".NzUzee div span").First().Text()

		products = append(products, ProductInfo{
			Title:    title,
			Price:    price,
			Link:     link,
			ImageURL: imageURL,
			Stars:    stars,
			// Reviews:  reviews,
		})
	})

	return products, nil
}

// StandardShoppingHandler handles shopping queries
func StandardShoppingHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	query := vars["query"]

	if query == "" {
		http.Error(w, "Query parameter is required", http.StatusBadRequest)
		return
	}

	config := ShoppingConfig{
		Query: query,
	}

	scraper := NewShoppingScraper(config)

	products, err := scraper.ShoppingScrape()
	if err != nil {
		http.Error(w, "Error scraping results", http.StatusInternalServerError)
		return
	}

	jsonData, err := json.MarshalIndent(products, "", "    ")
	if err != nil {
		http.Error(w, "Error marshaling to JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}
