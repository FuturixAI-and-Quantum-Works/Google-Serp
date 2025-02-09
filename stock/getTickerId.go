package stock

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
)

// Define the response struct
type TickerStockData struct {
	ID               string `json:"id"`
	Modified         int64  `json:"modified"`
	Created          int64  `json:"created"`
	MgIndustryId     string `json:"mgIndustryId"`
	MgSectorId       string `json:"mgSectorId"`
	StockType        string `json:"stockType"`
	Margins          string `json:"margins"`
	Trading          bool   `json:"trading"`
	Intraday         bool   `json:"intraday"`
	ListingStatusBse string `json:"listingStatusBse"`
	ListingStatusNsi string `json:"listingStatusNsi"`
	ExchangeCodeBse  string `json:"exchangeCodeBse"`
	ExchangeCodeNsi  string `json:"exchangeCodeNsi"`
	BseRic           string `json:"bseRic"`
	NseRic           string `json:"nseRic"`
	IsInId           string `json:"isInId"`
	CommonName       string `json:"commonName"`
	MgIndustry       string `json:"mgIndustry"`
	MgSector         string `json:"mgSector"`
	ReutersIndustry  string `json:"reutersIndustryClassification"`
	Gics             string `json:"gics"`

	ActiveStockTrends struct {
		TickerId       string `json:"tickerId"`
		ShortTermTrend string `json:"shortTermTrends"`
		LongTermTrend  string `json:"longTermTrends"`
		OverallRating  string `json:"overallRating"`
		Description    string `json:"description"`
		TickerType     string `json:"tickerType"`
	} `json:"activeStockTrends"`
}

// FetchStockData fetches stock data using the given query
func FetchStockData(query string) (*TickerStockData, error) {
	println("Fetching stock data for", query)
	url := fmt.Sprintf("https://api-mintgenie.livemint.com/api-gateway/fundamental/v2/searchFromIndustryTickerMaster?query=%s", query)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, err
	}

	// Set headers
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:135.0) Gecko/20100101 Firefox/135.0")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Cache-Control", "no-cache")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	// Read response body
	body, err := ioutil.ReadAll(resp.Body)
	print(string(body))
	if err != nil {
		return nil, err
	}

	// Parse JSON
	var stockData []TickerStockData
	err = json.Unmarshal(body, &stockData)
	if err != nil {
		return nil, err
	}

	if len(stockData) == 0 {
		return nil, nil
	}

	return &stockData[0], nil
}
