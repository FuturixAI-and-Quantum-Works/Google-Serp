package bingsearch

import "github.com/PuerkitoBio/goquery"

type PersonBoxContent struct {
	Name        string
	Position    string
	TermStarted string
}

func ExtractPersonBox(doc *goquery.Document) *BingAnswerBox {
	if personBox := doc.Find("li.b_ans.b_top"); personBox.Length() > 0 {
		personContent := &PersonBoxContent{}

		// Name
		if name := personBox.Find("div.b_focusTextMedium a"); name.Length() > 0 {
			personContent.Name = name.Text()
		}

		// Position
		if position := personBox.Find("div.b_focusLabel"); position.Length() > 0 {
			personContent.Position = position.Text()
		}

		// Term Started
		if termStarted := personBox.Find("div.extra_infoLabel"); termStarted.Length() > 0 {
			personContent.TermStarted = termStarted.Text()
		}
		if personContent.Name == "" {
			return nil
		}

		standard_search := &BingAnswerBox{
			Type:    "person",
			Content: personContent,
		}

		return standard_search
	}
	return nil
}
