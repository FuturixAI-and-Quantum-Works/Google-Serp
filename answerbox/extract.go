package answerbox

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

func ExtractAnswerBox(doc *goquery.Document) *AnswerBox {
	// Try each type of answer box in sequence
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
