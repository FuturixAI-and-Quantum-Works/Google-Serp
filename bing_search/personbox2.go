package bingsearch

import "github.com/PuerkitoBio/goquery"

type PersonBoxContent2 struct {
	Title       string
	Description string
	Citations   []Citation
}

type Citation struct {
	URL         string
	Title       string
	DisplayName string
	Snippet     string
}

func ExtractPersonBox2(doc *goquery.Document) *BingAnswerBox {
	if personBox := doc.Find("b_ans"); personBox.Length() > 0 {
		personContent := &PersonBoxContent2{}

		// Title
		if title := doc.Find("div.gs_pre_cont_title"); title.Length() > 0 {
			personContent.Title = title.Text()
		}

		// Description
		if description := doc.Find("div.gs_text.gs_mdr div.gs_p"); description.Length() > 0 {
			personContent.Description = description.Text()
		}

		// Citations
		doc.Find("div.gs_cit").Each(func(i int, s *goquery.Selection) {
			citation := Citation{}
			if url, exists := s.Attr("data-url"); exists {
				citation.URL = url
			}
			if title, exists := s.Attr("data-title"); exists {
				citation.Title = title
			}
			if displayName, exists := s.Attr("data-displayname"); exists {
				citation.DisplayName = displayName
			}
			if snippet := s.Find("div.gs_cit_snip").Text(); snippet != "" {
				citation.Snippet = snippet
			}
			personContent.Citations = append(personContent.Citations, citation)
		})

		standard_search := &BingAnswerBox{
			Type:    "person-2",
			Content: personContent,
		}

		return standard_search
	}
	return nil
}
