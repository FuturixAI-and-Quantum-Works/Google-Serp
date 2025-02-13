package stock

import (
	"compress/flate"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"googlescrapper/cache"
	"io"
	"net/http"
	"time"

	"github.com/andybalholm/brotli"
	"github.com/gorilla/mux"
	"github.com/klauspost/compress/zstd"
)

// ShareholdingTrend represents the shareholding trend response
type ShareholdingTrend struct {
	CategoryName string `json:"categoryName"`
	Categories   []struct {
		HoldingDate string `json:"holdingDate"`
		Percentage  string `json:"percentage"`
	} `json:"categories"`
}

// ShareholdingPieChart represents the pie chart response
type ShareholdingPieChart struct {
	Category   string `json:"category"`
	NoOfShares *int   `json:"noOfShares"`
	Percentage string `json:"percentage"`
}

// FetchShareholdings fetches shareholding details from the API
func FetchShareholdings(tickerId, shareType string) ([]ShareholdingTrend, error) {
	cacheKey := fmt.Sprintf("shareholdings:%s:%s", tickerId, shareType)

	return cache.Memoize(cacheKey, 12*time.Hour, func() ([]ShareholdingTrend, error) {

		url := fmt.Sprintf("https://api-mintgenie.livemint.com/api-gateway/fundamental/v2/getShareHoldingsDetailByTickerIdAndType?tickerId=%s&type=%s", tickerId, shareType)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:135.0) Gecko/20100101 Firefox/135.0")
		req.Header.Set("Accept", "*/*")
		req.Header.Set("Accept-Encoding", "gzip, deflate, br, zstd")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var body []byte
		switch resp.Header.Get("Content-Encoding") {
		case "gzip":
			gzipReader, err := gzip.NewReader(resp.Body)
			if err != nil {
				return nil, err
			}
			defer gzipReader.Close()
			body, err = io.ReadAll(gzipReader)
			if err != nil {
				return nil, err
			}
		case "deflate":
			flateReader := flate.NewReader(resp.Body)
			defer flateReader.Close()
			body, err = io.ReadAll(flateReader)
		case "br":
			brReader := brotli.NewReader(resp.Body)
			body, err = io.ReadAll(brReader)
		case "zstd":
			zstdReader, err := zstd.NewReader(resp.Body)
			if err != nil {
				return nil, err
			}
			defer zstdReader.Close()
			body, err = io.ReadAll(zstdReader)
			if err != nil {
				return nil, err
			}
		default:
			body, err = io.ReadAll(resp.Body)
		}
		if err != nil {
			return nil, err
		}

		var shareholdings []ShareholdingTrend
		err = json.Unmarshal(body, &shareholdings)
		if err != nil {
			return nil, err
		}

		return shareholdings, nil
	})
}

// GetShareholdingsHandler handles the API request for shareholding details
func GetShareholdingsHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	tickerId := params["tickerId"]
	shareType := params["type"]

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

	shareholdingsData, err := FetchShareholdings(livemintTicker.ID, shareType)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	println(shareholdingsData)
	json.NewEncoder(w).Encode(shareholdingsData)

}
