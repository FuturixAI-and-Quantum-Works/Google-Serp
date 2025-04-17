// Package bing_search provides extraction functionality for Bing weather boxes
package bingsearch

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

// ExtractWeatherBox extracts information from a weather box
func ExtractWeatherBox(doc *goquery.Document) *BingAnswerBox {
	weatherBox := doc.Find(".wtr_forecasts").First()
	if weatherBox.Length() == 0 {
		return nil
	}

	// Extract location
	location := weatherBox.Find(".wtr_loclink").Text()
	location = strings.TrimSpace(location)

	// Extract current temperature
	temperature := weatherBox.Find(".wtr_currtemp").Text()
	temperature = strings.TrimSpace(temperature)

	// Extract condition
	condition := weatherBox.Find(".wtr_condition").Text()
	condition = strings.TrimSpace(condition)

	// Get forecast days
	forecast := make([]map[string]string, 0)
	weatherBox.Find(".wtr_forecastday").Each(func(i int, s *goquery.Selection) {
		day := strings.TrimSpace(s.Find(".wtr_dayDate").Text())
		high := strings.TrimSpace(s.Find(".wtr_high").Text())
		low := strings.TrimSpace(s.Find(".wtr_low").Text())
		precip := strings.TrimSpace(s.Find(".wtr_precip").Text())

		forecastDay := map[string]string{
			"day":           day,
			"high":          high,
			"low":           low,
			"precipitation": precip,
		}

		forecast = append(forecast, forecastDay)
	})

	// Build attributes map
	attributes := map[string]interface{}{
		"location":    location,
		"temperature": temperature,
		"condition":   condition,
		"forecast":    forecast,
	}

	return &BingAnswerBox{
		Type:       "weather",
		Title:      "Weather for " + location,
		Content:    condition + ", " + temperature,
		Attributes: attributes,
	}
}
