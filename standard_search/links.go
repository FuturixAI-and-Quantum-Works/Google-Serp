package standard_search

import (
	"googlescrapper/utils"
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type SearchResult struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	URL     string `json:"url"`
	Favicon string `json:"favicon"`
}

func ExtractSearchResults(doc *goquery.Document, maxResults int) []SearchResult {
	var results []SearchResult

	doc.Find("div.g").Each(func(i int, sel *goquery.Selection) {
		if len(results) >= maxResults {
			return
		}

		titleSel := sel.Find("h3")
		urlSel := sel.Find("a").First()
		snippetSel := sel.Find("div.VwiC3b")

		title := titleSel.Text()
		url, _ := urlSel.Attr("href")
		snippet := snippetSel.Text()

		if title != "" && url != "" {
			results = append(results, SearchResult{
				Title:   strings.TrimSpace(title),
				Content: strings.TrimSpace(snippet),
				URL:     url,
				Favicon: utils.GetFavicon(url),
			})
		}
	})

	return results
}
