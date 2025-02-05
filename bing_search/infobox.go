package bingsearch

import "github.com/PuerkitoBio/goquery"

type Answer struct {
	Title       string
	Description string
	Lines       []string
}

func ExtractAnswer(doc *goquery.Document) *BingAnswerBox {
	answer := &Answer{}

	// Extract all df_c d_ans elements
	doc.Find("div.df_c.d_ans").Each(func(i int, s *goquery.Selection) {
		// Extract title
		if title := s.Find("div.df_da .b_focusTextMedium"); title.Length() > 0 {
			answer.Title = title.Text()
		}

		// Extract description
		if description := s.Find("div.df_con .rwrl_sec"); description.Length() > 0 {
			answer.Description = description.Text()
		}

		// Extract lines
		s.Find("div.rch-cap-cntr .rch-cap-list div").Each(func(i int, s *goquery.Selection) {
			line := s.Text()
			answer.Lines = append(answer.Lines, line)
		})
	})

	answerbox := &BingAnswerBox{
		Type:    "infobox",
		Content: answer,
	}

	if answer.Title == "" {
		return nil
	}		

	return answerbox
}
