package stock

import (
	"encoding/json"
	"fmt"
	"googlescrapper/cache"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/mux"
)

// Structure for API Response
type StockData struct {
	RelatedNews      []NewsArticle `json:"relatedNews"`
	StockPerformance string        `json:"stockPerformance"`
	CompanyDetails   CompanyInfo   `json:"companyDetails"`
	FAQs             []FAQ         `json:"faqs"`
}

type NewsArticle struct {
	Title string `json:"title"`
	Link  string `json:"link"`
	Time  string `json:"time"`
}

type CompanyInfo struct {
	Industry string `json:"industry"`
	ISIN     string `json:"isin"`
	BSECode  string `json:"bseCode"`
	NSECode  string `json:"nseCode"`
	About    string `json:"about"`
	CEO      string `json:"ceo"`
	CFO      string `json:"cfo"`
}

type FAQ struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

// Scrapes stock data from Livemint dynamically
func ScrapeStockData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stockIdentifier := vars["stockIdentifier"]

	cacheKey := fmt.Sprintf("stock-data:%s", stockIdentifier)

	// Use cache to memoize response for 5 minutes
	result, err := cache.Memoize(cacheKey, 12*time.Hour, func() (interface{}, error) {

		// Fetch stock ticker data
		liveMindTickerData, err := FetchStockTickerData(stockIdentifier)
		if err != nil {
			return nil, err
		}

		if len(liveMindTickerData) == 0 {
			return nil, fmt.Errorf("no data found for stock: %s", stockIdentifier)
		}

		livemintTicker := liveMindTickerData[0]

		// Generate correct identifier
		identifier := fmt.Sprintf("stocks-%s-share-price-nse-bse-%s",
			strings.ReplaceAll(strings.ToLower(livemintTicker.CommonName), " ", "-"),
			livemintTicker.ID,
		)

		fmt.Println("Fetching URL:", identifier)

		// Construct Livemint URL
		url := fmt.Sprintf("https://www.livemint.com/market/market-stats/%s", strings.ToLower(identifier))

		// Fetch the HTML page
		res, err := http.Get(url)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch webpage: %v", err)
		}
		defer res.Body.Close()

		if res.StatusCode != 200 {
			return nil, fmt.Errorf("failed to fetch webpage, status code: %d", res.StatusCode)
		}

		// Load HTML document
		doc, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to parse HTML: %v", err)
		}

		// Extract Related News
		var news []NewsArticle
		doc.Find(".storyDetails_newssecBlock__f2S5c li").Each(func(i int, s *goquery.Selection) {
			title := s.Find(".storyDetails_newsTitle__dxnD5 a").Text()
			link, _ := s.Find(".storyDetails_newsTitle__dxnD5 a").Attr("href")
			time := s.Find(".storyDetails_dateStock__lQbRz span").First().Text()

			if title != "" && link != "" {
				news = append(news, NewsArticle{
					Title: title,
					Link:  link,
					Time:  strings.TrimSpace(time),
				})
			}
		})

		// Extract Stock Performance
		stockPerformance := doc.Find(".storyDetails_readMore__K4Tyd").Text()

		// Extract Company Details
		company := CompanyInfo{
			Industry: doc.Find("#about_the_company ul li:nth-child(1) span").Text(),
			ISIN:     doc.Find("#about_the_company ul li:nth-child(2) span").Text(),
			BSECode:  doc.Find("#about_the_company ul li:nth-child(3) span").Text(),
			NSECode:  doc.Find("#about_the_company ul li:nth-child(4) span").Text(),
			About:    doc.Find("#about_the_company .storyDetails_compDesc__TvfHP").Text(),
			CEO:      doc.Find(".storyDetails_audiName__uAfOP").First().Text(),
			CFO:      doc.Find(".storyDetails_audiName__uAfOP").Eq(1).Text(),
		}

		// Extract FAQs
		var faqs []FAQ
		doc.Find(".storyDetails_accordionTab__1L_sn").Each(func(i int, s *goquery.Selection) {
			question := s.Find(".storyDetails_accordionTab-label__Cb52o").Text()
			answer := s.Find(".storyDetails_accordionTab-content__rCfVm p").Text()

			if question != "" && answer != "" {
				faqs = append(faqs, FAQ{
					Question: question,
					Answer:   strings.TrimSpace(answer),
				})
			}
		})

		// Create response object
		response := StockData{
			RelatedNews:      news,
			StockPerformance: strings.TrimSpace(stockPerformance),
			CompanyDetails:   company,
			FAQs:             faqs,
		}

		return response, nil
	})

	// Handle errors
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// Convert result to JSON and send response
	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// Main function
// func main() {
// 	router := mux.NewRouter()
// 	router.HandleFunc("/scrape/{stockIdentifier}", ScrapeStockData).Methods("GET")

// 	fmt.Println("Server running on port 8000")
// 	log.Fatal(http.ListenAndServe(":8000", router))
// }
