package main

import (
	"context"
	"encoding/json"
	"fmt"
	"googlescrapper/search"
	"log"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/gorilla/mux"
)

type cache struct {
	client *redis.Client
}

func newCache(addr string) *cache {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: "", // no password set
		DB:       0,  // use default DB
	})

	return &cache{client: client}
}

func (c *cache) get(ctx context.Context, key string) ([]byte, error) {
	return c.client.Get(ctx, key).Bytes()
}

func (c *cache) set(ctx context.Context, key string, value []byte, expiration time.Duration) error {
	return c.client.Set(ctx, key, value, expiration).Err()
}
func cacheMiddleware(c *cache, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		ctx := context.Background()
		key := r.URL.Path

		// Check if the query contains the word "time"
		if r.URL.Query().Get("query") != "" && containsTimeWord(r.URL.Query().Get("query")) {
			next.ServeHTTP(w, r)
			return
		}

		if cached, err := c.get(ctx, key); err == nil {
			w.Header().Set("Content-Type", "application/json")
			w.Write(cached)
			return
		}
		recorder := &responseRecorder{ResponseWriter: w}
		next.ServeHTTP(recorder, r)
		if recorder.status == 0 {
			var jsonBody map[string]interface{}
			if err := json.Unmarshal(recorder.body, &jsonBody); err == nil {
				if _, ok := jsonBody["error"]; ok {
					w.WriteHeader(http.StatusBadRequest)
				}
			}
			c.set(ctx, key, recorder.body, 1*time.Hour)
		}
	})
}

func containsTimeWord(query string) bool {
	return strings.Contains(query, "time")
}

type responseRecorder struct {
	http.ResponseWriter
	status int
	body   []byte
}

func (r *responseRecorder) Write(b []byte) (int, error) {
	r.body = b
	return r.ResponseWriter.Write(b)
}

func (r *responseRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func main() {
	search.ReadUserAgents()
	router := mux.NewRouter()
	router.HandleFunc("/search/{query}/{location}/{maxResults}/{latitude}/{longitude}/{useCoords}", search.StandardSearchHandler).Methods("GET")
	router.HandleFunc("/finance/{symbol}", search.StandardFinanceHandler)
	router.HandleFunc("/image/{query}", search.StandardImageHandler)
	router.HandleFunc("/shopping/{query}", search.StandardShoppingHandler)

	redisAddr := os.Getenv("REDIS_ADDR")
	if redisAddr == "" {
		redisAddr = "localhost:6379"
	}

	cache := newCache(redisAddr)
	router.Use(func(next http.Handler) http.Handler {
		return cacheMiddleware(cache, next)
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "8000" // fallback for local development
	}

	fmt.Printf("Server is running on port %s\n", port)
	log.Fatal(http.ListenAndServe(":"+port, router))
}
