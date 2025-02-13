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

type LivePriceResponse struct {
	ShortTermTrends string `json:"shortTermTrends"`
	LongTermTrends  string `json:"longTermTrends"`
	OverallRating   string `json:"overallRating"`
	Description     string `json:"description"`
}

func FetchLivePrice(tickerId, exchangeCode string) (LivePriceResponse, error) {
	cacheKey := fmt.Sprintf("live-price:%s:%s", tickerId, exchangeCode)

	return cache.Memoize(cacheKey, 5*time.Minute, func() (LivePriceResponse, error) {
		url := fmt.Sprintf("https://api-mintgenie.livemint.com/api-gateway/fundamental/markets-data/live-price/v4?tickerId=%s&exchangeCode=%s", tickerId, exchangeCode)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return LivePriceResponse{}, err
		}

		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:135.0) Gecko/20100101 Firefox/135.0")
		req.Header.Set("Accept", "*/*")
		req.Header.Set("Cache-Control", "no-cache")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return LivePriceResponse{}, err
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return LivePriceResponse{}, err
		}

		var livePrice LivePriceResponse
		if err := json.Unmarshal(body, &livePrice); err != nil {
			return LivePriceResponse{}, err
		}

		return livePrice, nil
	})
}

func GetLivePricePred(w http.ResponseWriter, r *http.Request) {
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

	livePrice, err := FetchLivePrice(livemintTicker.ID, "bse")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(livePrice)
}
