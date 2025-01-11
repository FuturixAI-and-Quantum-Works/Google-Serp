package standard_search

import (
	"github.com/PuerkitoBio/goquery"
)

type WeatherBoxContent struct {
	Temperature   string
	Condition     string
	Location      string
	Time          string
	Precipitation string
	Humidity      string
	Wind          string
	Forecast      []Forecast
}

type Forecast struct {
	Day       string
	HighTemp  string
	LowTemp   string
	Condition string
}

func ExtractWeatherBox(doc *goquery.Document) *AnswerBox {
	if weatherBox := doc.Find("div.nawv0d#wob_wc"); weatherBox.Length() > 0 {
		weatherContent := &WeatherBoxContent{}

		// Temperature
		if temp := weatherBox.Find("span#wob_tm"); temp.Length() > 0 {
			weatherContent.Temperature = temp.Text()
		}

		// Weather condition
		if condition := weatherBox.Find("span#wob_dc"); condition.Length() > 0 {
			weatherContent.Condition = condition.Text()
		}

		// Location and time
		if loc := weatherBox.Find("div.wob_loc"); loc.Length() > 0 {
			weatherContent.Location = loc.Text()
		}
		if time := weatherBox.Find("div.wob_dts"); time.Length() > 0 {
			weatherContent.Time = time.Text()
		}

		// Current conditions
		if info := weatherBox.Find("div.wtsRwe"); info.Length() > 0 {
			weatherContent.Precipitation = info.Find("span#wob_pp").Text()
			weatherContent.Humidity = info.Find("span#wob_hm").Text()
			weatherContent.Wind = info.Find("span#wob_ws").Text()
		}

		// Extract forecast data
		weatherBox.Find("div.wob_df").Each(func(i int, s *goquery.Selection) {
			day := s.Find("div.Z1VzSb").AttrOr("aria-label", "")
			highTemp := s.Find("div.gNCp2e span.wob_t").First().Text()
			lowTemp := s.Find("div.QrNVmd.ZXCv8e span.wob_t").First().Text()
			condition := s.Find("img.YQ4gaf").AttrOr("alt", "")

			forecast := Forecast{
				Day:       day,
				HighTemp:  highTemp,
				LowTemp:   lowTemp,
				Condition: condition,
			}
			weatherContent.Forecast = append(weatherContent.Forecast, forecast)
		})

		standard_search := &AnswerBox{
			Type:    "weather",
			Content: weatherContent,
		}

		return standard_search
	}
	return nil
}
