package stock

import (
	"bytes"
	"encoding/json"
	"fmt"
	"googlescrapper/cache"
	"io/ioutil"
	"net/http"
	"time"
)

type StockValue struct {
	Close     float64 `json:"close"`
	Volume    int     `json:"volume"`
	TimeStamp string  `json:"timeStamp"`
}

type StockChartResponse struct {
	TickerId    string       `json:"tickerId"`
	ReturnValue float64      `json:"returnValue"`
	Values      []StockValue `json:"values"`
}

type StockRequest struct {
	Days     string `json:"days"`
	TickerId string `json:"tickerId"`
}

func FetchStockChart(days, tickerId, tickerType string) ([]StockChartResponse, error) {
	cacheKey := fmt.Sprintf("stock-chart:%s:%s:%s", days, tickerId, tickerType)

	return cache.Memoize(cacheKey, 5*time.Minute, func() ([]StockChartResponse, error) {

		url := "https://api-mintgenie.livemint.com/api-gateway/fundamental/api/v2/charts"

		requestBody, err := json.Marshal(map[string]interface{}{
			"stockFilters": []map[string]string{
				{
					"days":       days,
					"tickerId":   tickerId,
					"tickerType": tickerType,
				},
			},
		})
		if err != nil {
			return nil, err
		}

		req, err := http.NewRequest("POST", url, bytes.NewBuffer(requestBody))
		if err != nil {
			return nil, err
		}

		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:135.0) Gecko/20100101 Firefox/135.0")
		req.Header.Set("Accept", "*/*")
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("mintgenie-client", "LM-WEB")
		req.Header.Set("Cache-Control", "no-cache")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, err
		}

		var stockData []StockChartResponse
		if err := json.Unmarshal(body, &stockData); err != nil {
			return nil, err
		}

		return stockData, nil
	})
}

func GetCharts(w http.ResponseWriter, r *http.Request) {
	var reqBody StockRequest
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(&reqBody); err != nil {
		http.Error(w, "Invalid request body", http.StatusBadRequest)
		return
	}

	TickerId := reqBody.TickerId

	liveMindTickerData, err := FetchStockTickerData(TickerId)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return

	}

	if len(liveMindTickerData) == 0 {
		http.Error(w, "No data found", http.StatusNotFound)
		return
	}

	livemintTicker := liveMindTickerData[0]

	livemintTickerJSON, err := json.Marshal(livemintTicker)
	if err != nil {
		http.Error(w, "Error converting ticker data to JSON", http.StatusInternalServerError)
		return
	}
	println(string(livemintTickerJSON))

	stockData, err := FetchStockChart(reqBody.Days, livemintTicker.ID, "bse")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(stockData)
}
