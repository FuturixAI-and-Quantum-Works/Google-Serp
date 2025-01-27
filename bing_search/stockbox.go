package bingsearch

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type NewsArticle struct {
	Title       string
	Source      string
	Thumbnail   string
	PublishDate string
}
type StockBoxContent struct {
	Name         string
	Ticker       string
	Exchange     string
	Rating       string
	AnalystCount string

	Analytics struct {
		Open           string
		Volume         string
		High           string
		AverageVol     string
		Low            string
		MarketCap      string
		FiftyTwoWkHigh string
		FiftyTwoWkLow  string
	}

	LastTrade struct {
		Price        string
		Currency     string
		Change       string
		Percentage   string
		LastUpdated  string
		MarketStatus string
	}

	News []*NewsArticle
}

func ExtractStockBox(doc *goquery.Document) *BingAnswerBox {
	if ansbox := doc.Find("div.b_slidesContainer"); ansbox.Length() > 0 {
		println("StockBox")
		companyContent := &StockBoxContent{}
		if companyBox := doc.Find("div.enti_c"); companyBox.Length() > 0 {
			println("CompanyBox")

			// Name
			if name := companyBox.Find("div.enti_ttl"); name.Length() > 0 {
				companyContent.Name = name.Text()
			}

			// Ticker and Exchange
			if ticker := companyBox.Find("div.enti_stxt"); ticker.Length() > 0 {
				tickerText := ticker.Text()
				split := strings.Split(tickerText, ":")
				if len(split) > 1 {
					companyContent.Exchange = strings.TrimSpace(split[0])
					companyContent.Ticker = strings.TrimSpace(split[1])
				} else {
					companyContent.Ticker = tickerText
				}
			}

		}
		if analystRatingBox := doc.Find("a#acFinAR"); analystRatingBox.Length() > 0 {

			// Rating
			if rating := analystRatingBox.Find("div.finar_st"); rating.Length() > 0 {
				companyContent.Rating = rating.Text()
			}

			// Analyst Count
			if analystCount := analystRatingBox.Find("div.finar_sb"); analystCount.Length() > 0 {
				companyContent.AnalystCount = analystCount.Text()
			}
		}

		if stockDetails := doc.Find("div#stockDetails"); stockDetails.Length() > 0 {

			// Open
			if open := stockDetails.Find("div.fin_dtcell[aria-label*='Open'] div.fin_dtval"); open.Length() > 0 {
				companyContent.Analytics.Open = open.Text()
			}

			// Volume
			if volume := stockDetails.Find("div.fin_dtcell[aria-label*='Vol'] div.fin_dtval"); volume.Length() > 0 {
				companyContent.Analytics.Volume = volume.Text()
			}

			// High
			if high := stockDetails.Find("div.fin_dtcell[aria-label*='High'] div.fin_dtval"); high.Length() > 0 {
				companyContent.Analytics.High = high.Text()
			}

			// Average Volume
			if averageVol := stockDetails.Find("div.fin_dtcell[aria-label*='Avg Vol'] div.fin_dtval"); averageVol.Length() > 0 {
				companyContent.Analytics.AverageVol = averageVol.Text()
			}

			// Low
			if low := stockDetails.Find("div.fin_dtcell[aria-label*='Low'] div.fin_dtval"); low.Length() > 0 {
				companyContent.Analytics.Low = low.Text()
			}

			// Market Cap
			if marketCap := stockDetails.Find("div.fin_dtcell[aria-label*='Mkt Cap'] div.fin_dtval"); marketCap.Length() > 0 {
				companyContent.Analytics.MarketCap = marketCap.Text()
			}

			// Fifty Two Week High
			if fiftyTwoWkHigh := stockDetails.Find("div.fin_dtcell[aria-label*='52wk High'] div.fin_dtval"); fiftyTwoWkHigh.Length() > 0 {
				companyContent.Analytics.FiftyTwoWkHigh = fiftyTwoWkHigh.Text()
			}

			// Fifty Two Week Low
			if fiftyTwoWkLow := stockDetails.Find("div.fin_dtcell[aria-label*='52wk Low'] div.fin_dtval"); fiftyTwoWkLow.Length() > 0 {
				companyContent.Analytics.FiftyTwoWkLow = fiftyTwoWkLow.Text()
			}
		}

		if stockQuote := doc.Find("div.q_head"); stockQuote.Length() > 0 {

			// Price
			if price := stockQuote.Find("div#Finance_Quote div.b_focusTextMedium"); price.Length() > 0 {
				companyContent.LastTrade.Price = price.Text()
			}

			// Currency
			if currency := stockQuote.Find("span.price_curr"); currency.Length() > 0 {
				companyContent.LastTrade.Currency = currency.Text()
			}

			// Change and Percentage
			if change := stockQuote.Find("span.fin_change"); change.Length() > 0 {
				changeText := change.Text()
				split := strings.Split(changeText, " ")
				if len(split) > 1 {
					companyContent.LastTrade.Change = split[0]
					companyContent.LastTrade.Percentage = split[1]
				} else {
					companyContent.LastTrade.Change = changeText
				}
			}

			// Last Updated
			if lastUpdated := stockQuote.Find("span.fin_lastUpdate"); lastUpdated.Length() > 0 {
				companyContent.LastTrade.LastUpdated = lastUpdated.Text()
			}

			// Market Status
			if marketStatus := stockQuote.Find("span.fin_marketStatusBadge"); marketStatus.Length() > 0 {
				companyContent.LastTrade.MarketStatus = marketStatus.Text()
			}
		}
		doc.Find("a.finnac_item").Each(func(i int, s *goquery.Selection) {
			article := &NewsArticle{}

			// Title
			if title := s.Find("div.finnac_t"); title.Length() > 0 {
				article.Title = title.Text()
			}

			// Source
			if source := s.Find("div.finnac_f span.finnac_pi span.attr_sep + span"); source.Length() > 0 {
				article.Source = source.Text()
			}

			// Thumbnail
			if thumbnail := s.Find("div.cico img"); thumbnail.Length() > 0 {
				article.Thumbnail = thumbnail.AttrOr("src", "")
			}

			// Publish Date
			if publishDate := s.Find("div.finnac_f span.finnac_pi span.art_time"); publishDate.Length() > 0 {
				article.PublishDate = publishDate.Text()
			}

			companyContent.News = append(companyContent.News, article)
		})
		if companyContent.Name == "" {
			return nil
		}
		// return nil if
		standard_search := &BingAnswerBox{
			Type:    "stock",
			Content: companyContent,
		}

		return standard_search
	}
	return nil
}
