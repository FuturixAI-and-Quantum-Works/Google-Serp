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
)

// StockInfo represents detailed stock information
type StockInfo struct {
	ID              string `json:"id"`
	CommonName      string `json:"commonName"`
	MgIndustry      string `json:"mgIndustry"`
	MgSector        string `json:"mgSector"`
	ExchangeCodeBse string `json:"exchangeCodeBse"`
	ExchangeCodeNsi string `json:"exchangeCodeNsi"`
	BseRic          string `json:"bseRic"`
	NseRic          string `json:"nseRic"`
	IsInId          string `json:"isInId"`
}

// FetchStockTickerData fetches stock data from MintGenie with caching
func FetchStockTickerData(query string) ([]StockInfo, error) {
	cacheKey := fmt.Sprintf("stock:%s", query)

	return cache.Memoize(cacheKey, 12*time.Hour, func() ([]StockInfo, error) {
		url := fmt.Sprintf("https://api-mintgenie.livemint.com/api-gateway/fundamental/v2/searchFromIndustryTickerMaster?query=%s", query)

		req, err := http.NewRequest("GET", url, nil)
		if err != nil {
			return nil, err
		}

		req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:135.0) Gecko/20100101 Firefox/135.0")
		req.Header.Set("Accept-Encoding", "gzip, deflate, br, zstd")

		client := &http.Client{}
		resp, err := client.Do(req)
		if err != nil {
			return nil, err
		}
		defer resp.Body.Close()

		var reader io.ReadCloser
		switch resp.Header.Get("Content-Encoding") {
		case "gzip":
			reader, err = gzip.NewReader(resp.Body)
			if err != nil {
				return nil, err
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

		body, err := io.ReadAll(reader)
		if err != nil {
			return nil, err
		}

		var stockInfo []StockInfo
		if err := json.Unmarshal(body, &stockInfo); err != nil {
			return nil, err
		}

		return stockInfo, nil
	})
}
