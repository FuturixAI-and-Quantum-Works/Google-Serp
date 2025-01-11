package finance

import "github.com/PuerkitoBio/goquery"

func ExtractCompanyData(doc *goquery.Document) map[string]string {
	data := make(map[string]string)

	// Extract company description
	description := doc.Find(".bLLb2d").Text()
	data["Description"] = description

	// Extract company data
	doc.Find(".gyFHrc").Each(func(i int, s *goquery.Selection) {
		label := s.Find(".mfs7Fc").Text()

		if label == "CEO" {
			value := s.Find(".tBHE4e").Text()
			data[label] = value
		} else if label == "Website" {
			value := s.Find(".tBHE4e").AttrOr("href", "")
			data[label] = value
		} else {
			value := s.Find(".P6K39c").Text()
			data[label] = value
		}
	})

	return data
}
