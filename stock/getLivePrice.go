package stock

import (
	"encoding/json"
	"fmt"
	"googlescrapper/cache"
	"io/ioutil"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// LivePriceV2Response represents the live stock price response
type LivePriceV2Response struct {
	TickerId              string `json:"tickerId"`
	Ric                   string `json:"ric"`
	Price                 string `json:"price"`
	PercentChange         string `json:"percentChange"`
	NetChange             string `json:"netChange"`
	Bid                   string `json:"bid"`
	Ask                   string `json:"ask"`
	High                  string `json:"high"`
	Low                   string `json:"low"`
	Open                  string `json:"open"`
	LowCircuitLimit       string `json:"lowCircuitLimit"`
	UpCircuitLimit        string `json:"upCircuitLimit"`
	Volume                string `json:"volume"`
	DisplayName           string `json:"displayName"`
	Date                  string `json:"date"`
	Time                  string `json:"time"`
	PriceArrow            string `json:"priceArrow"`
	Close                 string `json:"close"`
	BidSize               string `json:"bidSize"`
	AskSize               string `json:"askSize"`
	ExchangeType          string `json:"exchangeType"`
	LotSize               string `json:"lotSize"`
	TotalShareOutstanding string `json:"totalShareOutstanding"`
	MarketCap             string `json:"marketCap"`
	ShortTermTrends       string `json:"shortTermTrends"`
	LongTermTrends        string `json:"longTermTrends"`
	OverallRating         string `json:"overallRating"`
	Description           string `json:"description"`
	ImageUrl              string `json:"imageUrl"`
	YLow                  string `json:"ylow"`
	YHigh                 string `json:"yhigh"`
}

// FetchLivePriceV2 fetches live stock price from the API
func FetchLivePriceV2(tickerId, exchangeCode string) (LivePriceV2Response, error) {
	cacheKey := fmt.Sprintf("live-price-v2:%s:%s", tickerId, exchangeCode)

	return cache.Memoize(cacheKey, 5*time.Minute, func() (LivePriceV2Response, error) {

		url := fmt.Sprintf("https://api-mintgenie.livemint.com/api-gateway/fundamental/markets-data/live-price/v2?exchangeCode=%s&tickerId=%s", exchangeCode, tickerId)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return LivePriceV2Response{}, err
		}

		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:135.0) Gecko/20100101 Firefox/135.0")
		req.Header.Set("Accept", "*/*")
		req.Header.Set("Cache-Control", "no-cache")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return LivePriceV2Response{}, err
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return LivePriceV2Response{}, err
		}

		var livePrice LivePriceV2Response
		if err := json.Unmarshal(body, &livePrice); err != nil {
			return LivePriceV2Response{}, err
		}

		return livePrice, nil
	})
}

// GetLivePriceV2Handler handles the API request for live stock price
func GetLivePriceV2Handler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	tickerId := params["tickerId"]

	liveMindTickerData, err := FetchStockTickerData(tickerId)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return

	}

	if len(liveMindTickerData) == 0 {
		http.Error(w, "No data found", http.StatusNotFound)
		return
	}

	livemintTicker := liveMindTickerData[0]

	livePrice, err := FetchLivePriceV2(livemintTicker.ID, "bse")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(livePrice)
}

// func main() {
// 	router := mux.NewRouter()
// 	router.HandleFunc("/stock/live-price/v2/{tickerId}/{exchangeCode}", GetLivePriceV2Handler).Methods("GET")

// 	port := "8000"
// 	fmt.Printf("Server is running on port %s\n", port)
// 	http.ListenAndServe(":"+port, router)
// }
