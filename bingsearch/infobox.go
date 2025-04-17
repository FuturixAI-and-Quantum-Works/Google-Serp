// Package bing_search provides extraction functionality for Bing info boxes
package bingsearch

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type Answer struct {
	Title       string
	Description string
	Lines       []string
}

// ExtractInfoBox extracts information from an entity info box
func ExtractInfoBox(doc *goquery.Document) *BingAnswerBox {
	infoBox := doc.Find(".b_infocardTop").First()
	if infoBox.Length() == 0 {
		return nil
	}

	// Extract the title/entity name
	title := infoBox.Find("h2").Text()
	title = strings.TrimSpace(title)

	// Extract the description
	description := infoBox.Find(".b_snippet").Text()
	description = strings.TrimSpace(description)

	// Extract attributes
	attributes := make(map[string]interface{})
	infoBox.Find(".b_factrow").Each(func(i int, s *goquery.Selection) {
		label := strings.TrimSpace(s.Find(".b_label").Text())
		value := strings.TrimSpace(s.Find(".b_factvalue").Text())

		if label != "" && value != "" {
			attributes[label] = value
		}
	})

	return &BingAnswerBox{
		Type:       "info_box",
		Title:      title,
		Content:    description,
		Attributes: attributes,
	}
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
