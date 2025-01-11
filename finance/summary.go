package finance

import (
	"github.com/PuerkitoBio/goquery"
)

type FinanceSummary struct {
	PreviousClose   string `json:"previousClose"`
	DayRange        string `json:"dayRange"`
	YearRange       string `json:"yearRange"`
	MarketCap       string `json:"marketCap"`
	AvgVolume       string `json:"avgVolume"`
	PERatio         string `json:"peRatio"`
	DividendYield   string `json:"dividendYield"`
	PrimaryExchange string `json:"primaryExchange"`
}

func ExtractFinanceSummary(doc *goquery.Document) *FinanceSummary {
	data := &FinanceSummary{}

	doc.Find(".gyFHrc").Each(func(i int, s *goquery.Selection) {
		label := s.Find(".mfs7Fc").Text()
		value := s.Find(".P6K39c").Text()

		switch label {
		case "Previous close":
			data.PreviousClose = value
		case "Day range":
			data.DayRange = value
		case "Year range":
			data.YearRange = value
		case "Market cap":
			data.MarketCap = value
		case "Avg Volume":
			data.AvgVolume = value
		case "P/E ratio":
			data.PERatio = value
		case "Dividend yield":
			data.DividendYield = value
		case "Primary exchange":
			data.PrimaryExchange = value
		}
	})

	return data
}
