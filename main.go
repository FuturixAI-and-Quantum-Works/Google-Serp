package main

import (
	"fmt"
	"googlescrapper/search"
	"googlescrapper/stock"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/handlers"
	"github.com/gorilla/mux"
)

func main() {
	search.ReadUserAgents()
	router := mux.NewRouter()

	// Define routes
	router.HandleFunc("/search/{query}/{location}/{maxResults}/{latitude}/{longitude}/{useCoords}", search.StandardSearchHandler).Methods("GET")
	router.HandleFunc("/finance/{symbol}", search.StandardFinanceHandler)
	router.HandleFunc("/image/{query}", search.StandardImageHandler)
	router.HandleFunc("/shopping/{query}", search.StandardShoppingHandler)
	router.HandleFunc("/bing/{query}", search.StandardBingHandler)
	router.HandleFunc("/html", search.GetHTMLFromUrl)
	router.HandleFunc("/stock/charts", stock.GetCharts)
	router.HandleFunc("/stock/live/{tickerId}", stock.GetLivePricePred)
	router.HandleFunc("/stock/shareholdings/{tickerId}/{type}", stock.GetShareholdingsHandler)
	router.HandleFunc("/stock/live-price/{tickerId}", stock.GetLivePriceV2Handler)
	router.HandleFunc("/stock/forecast/{tickerId}", stock.GetStockForecastHandler).Methods("GET")
	router.HandleFunc("/scrape/{stockIdentifier}", stock.ScrapeStockData).Methods("GET")

	// Read environment variables
	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000" // fallback for local development
	}

	// Set up CORS middleware
	corsHandler := handlers.CORS(
		handlers.AllowedOrigins([]string{"*"}),
		handlers.AllowedMethods([]string{"GET", "POST", "PUT", "DELETE", "OPTIONS"}),
		handlers.AllowedHeaders([]string{"Content-Type", "Authorization"}),
	)(router)

	fmt.Printf("Server is running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, corsHandler))
}
