package main

import (
	"fmt"
	"googlescrapper/search"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

// SearchResult represents a single search result

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/search/{query}/{location}/{maxResults}/{latitude}/{longitude}/{useCoords}", search.StandardSearchHandler).Methods("GET")
	router.HandleFunc("/finance/{symbol}", search.StandardFinanceHandler)
	router.HandleFunc("/image/{query}", search.StandardImageHandler)
	fmt.Println("Server is running on port 8000")
	log.Fatal(http.ListenAndServe(":8000", router))
}
