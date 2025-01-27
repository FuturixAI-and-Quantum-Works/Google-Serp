package bingsearch

import "github.com/PuerkitoBio/goquery"

type BingAnswerBox struct {
	Type        string      `json:",omitempty"`
	Content     interface{} `json:",omitempty"`
	RelatedText string      `json:",omitempty"`
	Source      string      `json:",omitempty"`
	SourceURL   string      `json:",omitempty"`
}

func ExtractAnswerbox(doc *goquery.Document) *BingAnswerBox {
	if box := ExtractPersonBox(doc); box != nil {
		return box
	}

	if box := ExtractPersonBox2(doc); box != nil {
		return box
	}

	if box := ExtractTimeBox(doc); box != nil {
		return box
	}

	if box := ExtractStockBox(doc); box != nil {
		return box
	}

	if box := ExtractWeather(doc); box != nil {
		return box
	}

	if box := ExtractAnswer(doc); box != nil {
		return box
	}

	return &BingAnswerBox{}
}
