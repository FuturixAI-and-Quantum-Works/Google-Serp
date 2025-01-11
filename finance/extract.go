package finance

import "github.com/PuerkitoBio/goquery"

type FinanceData struct {
	Type           string
	FinanceSummary *FinanceSummary `json:"about"`
}

func ExtractFinanceData(doc *goquery.Document) *FinanceData {
	box := &FinanceData{
		Type: "google_finance",
	}

	box.FinanceSummary = ExtractFinanceSummary(doc)

	return box
}
