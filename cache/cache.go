package cache

import (
	"context"
	"encoding/json"
	"time"

	"github.com/go-redis/redis/v8"
)

// RedisClient is a global Redis client instance
var RedisClient = redis.NewClient(&redis.Options{
	Addr:     "localhost:6379", // Change if needed
	Password: "",               // No password by default
	DB:       0,                // Default DB
})

// Memoize function for caching any function result in Redis
func Memoize[T any](key string, ttl time.Duration, fn func() (T, error)) (T, error) {
	var result T
	ctx := context.Background()

	// Try fetching from cache
	cachedData, err := RedisClient.Get(ctx, key).Bytes()
	if err == nil {
		if jsonErr := json.Unmarshal(cachedData, &result); jsonErr == nil {
			return result, nil
		}
	}

	// Call the actual function
	result, err = fn()
	if err != nil {
		return result, err
	}

	// Store result in cache
	cacheData, _ := json.Marshal(result)
	RedisClient.Set(ctx, key, cacheData, ttl)

	return result, nil
}
