package finance

import "github.com/PuerkitoBio/goquery"

func ExtractNewsArticles(doc *goquery.Document) []map[string]string {
	articles := make([]map[string]string, 0)

	doc.Find("div.yY3Lee").Each(func(i int, s *goquery.Selection) {
		article := make(map[string]string)

		article["source"] = s.AttrOr("data-article-source-name", "")
		article["title"] = s.Find("div.Yfwt5").Text()
		article["link"] = s.Find("a").AttrOr("href", "")
		article["image"] = s.Find("img.Z4idke").AttrOr("src", "")
		article["time"] = s.Find("div.Adak").Text()

		articles = append(articles, article)
	})

	return articles
}
