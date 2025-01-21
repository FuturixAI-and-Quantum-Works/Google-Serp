package main

import (
	"fmt"
	"googlescrapper/search"
	"log"
	"net/http"
	"os"

	"github.com/gorilla/mux"
)

func main() {
	router := mux.NewRouter()
	router.HandleFunc("/search/{query}/{location}/{maxResults}/{latitude}/{longitude}/{useCoords}", search.StandardSearchHandler).Methods("GET")
	router.HandleFunc("/finance/{symbol}", search.StandardFinanceHandler)
	router.HandleFunc("/image/{query}", search.StandardImageHandler)
	router.HandleFunc("/shopping/{query}", search.StandardShoppingHandler)

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000" // fallback for local development
	}

	fmt.Printf("Server is running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
