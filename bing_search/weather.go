package bingsearch

import "github.com/PuerkitoBio/goquery"

type Weather struct {
	Location    string
	UpdateTime  string
	CurrentTime string
	Condition   string
	Temperature string
	High        string
	Low         string
	Wind        string
	Humidity    string
	Forecast    []Forecast
}

type Forecast struct {
	Day  string
	High string
	Low  string
}

func ExtractWeather(doc *goquery.Document) *BingAnswerBox {
	weather := &Weather{}

	// Location
	if location := doc.Find("div.wtr_locTitleWrap div.wtr_locTitle span.wtr_foreGround"); location.Length() > 0 {
		weather.Location = location.Text()
	}

	// Update Time
	if updateTime := doc.Find("div.wtr_locTitleWrap div.wtr_lastUpdate div.b_meta"); updateTime.Length() > 0 {
		weather.UpdateTime = updateTime.Text()
	}

	// Condition
	if condition := doc.Find("div.wtr_condiSecondary div.wtr_caption"); condition.Length() > 0 {
		weather.Condition = condition.Text()
	}

	// Temperature
	if temperature := doc.Find("div.wtr_condiTemp div.wtr_currTemp"); temperature.Length() > 0 {
		weather.Temperature = temperature.Text()
	}

	// High
	if high := doc.Find("div.wtr_condiHighLow div.wtr_high span"); high.Length() > 0 {
		weather.High = high.Text()
	}

	// Low
	if low := doc.Find("div.wtr_condiHighLow div.wtr_low"); low.Length() > 0 {
		weather.Low = low.Text()
	}

	// Wind
	if wind := doc.Find("div.wtr_condiAttribs div.wtr_currWind"); wind.Length() > 0 {
		weather.Wind = wind.Text()
	}

	// Humidity
	if humidity := doc.Find("div.wtr_condiAttribs div.wtr_currHumi"); humidity.Length() > 0 {
		weather.Humidity = humidity.Text()
	}

	if dateTime := doc.Find("div.wtr_dayTime_id.wtr_dayTime"); dateTime.Length() > 0 {
		weather.CurrentTime = dateTime.Text()
	}

	// Forecast
	forecast := []Forecast{}
	doc.Find("div.vpc").Each(func(i int, s *goquery.Selection) {
		day := s.Find("div.wtr_weekday span").Text()
		high := s.Find("div.wtr_high span").Text()
		low := s.Find("div.wtr_low span").Text()
		forecast = append(forecast, Forecast{Day: day, High: high, Low: low})
	})
	weather.Forecast = forecast
	answerbox := &BingAnswerBox{
		Type:    "weather",
		Content: weather,
	}

	if weather.Location == "" {
		return nil
	}

	return answerbox
}
