package standard_search

import (
	"github.com/PuerkitoBio/goquery"
)

type AnswerBox struct {
	Type        string      `json:",omitempty"`
	Content     interface{} `json:",omitempty"`
	RelatedText string      `json:",omitempty"`
	Source      string      `json:",omitempty"`
	SourceURL   string      `json:",omitempty"`
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
