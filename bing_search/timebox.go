package bingsearch

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type TimeBoxContent struct {
	Location string
	Time     string
	Date     string
	TimeZone string
}

func ExtractTimeBox(doc *goquery.Document) *BingAnswerBox {
	if timeBox := doc.Find("li.b_ans.b_top.b_topborder"); timeBox.Length() > 0 {
		timeContent := &TimeBoxContent{}

		// Location
		if location := timeBox.Find("div.b_focusLabel"); location.Length() > 0 {
			timeContent.Location = location.Text()
		}

		// Time
		if time := timeBox.Find("div.b_focusTextLarge"); time.Length() > 0 {
			timeContent.Time = time.Text()
		}

		// Date
		if date := timeBox.Find("div.b_secondaryFocus"); date.Length() > 0 {
			timeContent.Date = date.Text()
		}

		// Time Zone
		if location := timeBox.Find("div.b_focusLabel"); location.Length() > 0 {
			timeContent.TimeZone = location.Text()
			timeContent.TimeZone = timeContent.TimeZone[strings.Index(timeContent.TimeZone, "(")+1 : strings.Index(timeContent.TimeZone, ")")]
		}

		if timeContent.Location == "" {
			return nil
		}

		standard_search := &BingAnswerBox{
			Type:    "time",
			Content: timeContent,
		}

		return standard_search
	}
	return nil
}
