package main

import (
	"fmt"
	"googlescrapper/search"
	"googlescrapper/stock"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	search.ReadUserAgents()
	router := mux.NewRouter()
	router.HandleFunc("/search/{query}/{location}/{maxResults}/{latitude}/{longitude}/{useCoords}", search.StandardSearchHandler).Methods("GET")
	router.HandleFunc("/finance/{symbol}", search.StandardFinanceHandler)
	router.HandleFunc("/image/{query}", search.StandardImageHandler)
	router.HandleFunc("/shopping/{query}", search.StandardShoppingHandler)
	router.HandleFunc("/bing/{query}", search.StandardBingHandler)
	router.HandleFunc("/stock/charts", stock.GetCharts)
	router.HandleFunc("/stock/live/{tickerId}", stock.GetLivePricePred)
	// TrendDetails or PieChartDetails
	router.HandleFunc("/stock/shareholdings/{tickerId}/{type}", stock.GetShareholdingsHandler)
	router.HandleFunc("/stock/live-price/{tickerId}", stock.GetLivePriceV2Handler)
	router.HandleFunc("/stock/forecast/{tickerId}", stock.GetStockForecastHandler).Methods("GET")
	router.HandleFunc("/scrape/{stockIdentifier}", stock.ScrapeStockData).Methods("GET")

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	// cache := newCache(redisAddr)
	// router.Use(func(next http.Handler) http.Handler {
	// 	return cacheMiddleware(cache, next)
	// })

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000" // fallback for local development
	}

	fmt.Printf("Server is running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
