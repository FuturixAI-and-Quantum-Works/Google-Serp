package finance

import "github.com/PuerkitoBio/goquery"

func ExtractCompanyInfo(doc *goquery.Document) map[string]string {
	data := make(map[string]string)

	doc.Find(".gyFHrc").Each(func(i int, s *goquery.Selection) {
		label := s.Find(".mfs7Fc").Text()
		value := s.Find(".P6K39c").Text()

		data[label] = value
	})

	return data
}
