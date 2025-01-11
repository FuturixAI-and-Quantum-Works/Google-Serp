package standard_search

import "github.com/PuerkitoBio/goquery"

type FeaturedSnippetContent struct {
	Title       string
	Description string
	Source      string
	SourceURL   string
}

func ExtractFeaturedSnippet(doc *goquery.Document) *AnswerBox {
	if featuredSnippet := doc.Find("div.g.wF4fFd.JnwWd.g-blk"); featuredSnippet.Length() > 0 {
		featuredSnippetContent := &FeaturedSnippetContent{}

		// Title
		if title := featuredSnippet.Find("h2.bNg8Rb"); title.Length() > 0 {
			featuredSnippetContent.Title = title.Text()
		}

		// Description
		if description := featuredSnippet.Find("span.hgKElc"); description.Length() > 0 {
			featuredSnippetContent.Description = description.Text()
		}

		// Source
		if source := featuredSnippet.Find("div.CA5RN div.VuuXrf"); source.Length() > 0 {
			featuredSnippetContent.Source = source.Text()
		}

		// Source URL
		if sourceURL := featuredSnippet.Find("div.CA5RN div.byvV5b cite.qLRx3b"); sourceURL.Length() > 0 {
			featuredSnippetContent.SourceURL = sourceURL.Text()
		}

		standard_search := &AnswerBox{
			Type:    "featured_snippet",
			Content: featuredSnippetContent,
		}

		return standard_search
	}
	return nil
}
