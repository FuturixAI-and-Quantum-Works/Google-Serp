package standard_search

import (
	"github.com/PuerkitoBio/goquery"
)

type AnswerBox struct {
	Type        string
	Content     interface{}
	RelatedText string
	Source      string
	SourceURL   string
}

func ExtractAnswerbox(doc *goquery.Document) *AnswerBox {
	if box := ExtractWeatherBox(doc); box != nil {
		return box
	}
	if box := ExtractTimeBox(doc); box != nil {
		return box
	}
	if box := ExtractMathBox(doc); box != nil {
		return box
	}
	if box := ExtractFeaturedSnippet(doc); box != nil {
		return box
	}
	if box := ExtractStockBox(doc); box != nil {
		return box
	}

	return nil
}
