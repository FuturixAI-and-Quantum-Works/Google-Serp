package finance

import "github.com/PuerkitoBio/goquery"

type FinanceData struct {
	Type        string
	CompanyInfo map[string]string
	CompanyData map[string]string
	News        []map[string]string
}

func ExtractFinanceData(doc *goquery.Document) *FinanceData {
	box := &FinanceData{
		Type: "google_finance",
	}

	box.CompanyInfo = ExtractCompanyInfo(doc)
	box.CompanyData = ExtractCompanyData(doc)
	box.News = ExtractNewsArticles(doc)
	return box
}
