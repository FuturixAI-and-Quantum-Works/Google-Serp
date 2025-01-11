package search

import (
	"encoding/json"
	"fmt"
	"googlescrapper/finance"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/mux"
)

type StockInfo struct {
	Symbol    string `json:"symbol"`
	Name      string `json:"name"`
	Price     string `json:"price"`
	Change    string `json:"change"`
	MarketCap string `json:"marketCap"`
}

type NewsItem struct {
	Title string `json:"title"`
	URL   string `json:"url"`
}

type FinanceConfig struct {
	Symbol string
	Window string
}

// FinanceScraper handles the scraping functionality
type FinanceScraper struct {
	client *http.Client
	config FinanceConfig
}

// NewFinanceScraper creates a new scraper instance
func NewFinanceScraper(config FinanceConfig) *FinanceScraper {
	return &FinanceScraper{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		config: config,
	}
}

// buildFinanceURL creates the finance URL with parameters
func (s *FinanceScraper) buildFinanceURL(symbol, window string) string {
	return fmt.Sprintf("https://finance.google.com/finance?q=%s&window=%s", symbol, window)
}

func (s *FinanceScraper) FinanceScrape() (*finance.FinanceData, error) {
	req, err := http.NewRequest("GET", s.buildFinanceURL(s.config.Symbol, s.config.Window), nil)
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

	// write to file
	ioutil.WriteFile("finance.html", body, 0644)

	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %v", err)
	}

	financeResponse := finance.ExtractFinanceData(doc)

	return financeResponse, nil
}

// StandardFinanceHandler handles finance queries
func StandardFinanceHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	symbol := vars["symbol"]

	if symbol == "" {
		http.Error(w, "Symbol parameter is required", http.StatusBadRequest)
		return
	}

	window := r.URL.Query().Get("window")
	if window == "" {
		window = "1d" // default window if not provided
	}

	config := FinanceConfig{
		Symbol: symbol,
		Window: window,
	}

	scraper := NewFinanceScraper(config)

	financeResponse, err := scraper.FinanceScrape()
	if err != nil {
		http.Error(w, "Error scraping results", http.StatusInternalServerError)
		return
	}

	jsonData, err := json.MarshalIndent(financeResponse, "", "    ")
	if err != nil {
		http.Error(w, "Error marshaling to JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}
