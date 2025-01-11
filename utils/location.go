package utils

import "fmt"

func CreateUULE(lat, lon float64) string {
	return fmt.Sprintf("w+CAIQICI%s", EncodeCoordinates(lat, lon))
}

// encodeCoordinates creates a simple encoding of the coordinates
func EncodeCoordinates(lat, lon float64) string {
	// Basic encoding - in practice, Google uses a more sophisticated algorithm
	return fmt.Sprintf("%f:%f", lat, lon)
}
