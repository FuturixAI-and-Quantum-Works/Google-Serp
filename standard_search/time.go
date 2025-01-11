package standard_search

import (
	"github.com/PuerkitoBio/goquery"
)

type TimeBoxContent struct {
	Time     string
	Day      string
	Location string
}

func ExtractTimeBox(doc *goquery.Document) *AnswerBox {
	if timeBox := doc.Find("div.vk_gy.vk_sh.card-section.sL6Rbf"); timeBox.Length() > 0 {
		timeContent := &TimeBoxContent{}

		// Time
		if time := timeBox.Find("div.gsrt.vk_bk.FzvWSb"); time.Length() > 0 {
			timeContent.Time = time.Text()
		}

		// Day and date
		if day := timeBox.Find("div.vk_gy.vk_sh"); day.Length() > 0 {
			timeContent.Day = day.Text()
		}

		// Location
		if location := timeBox.Find("span.vk_gy.vk_sh"); location.Length() > 0 {
			timeContent.Location = location.Text()
		}

		standard_search := &AnswerBox{
			Type:    "time",
			Content: timeContent,
		}

		return standard_search
	}
	return nil
}
