package answerbox

import (
	"strings"

	"github.com/PuerkitoBio/goquery"
)

type StockBoxContent struct {
	CompanyName string
	Exchange    string
	TickerId    string
	Price       string
	PriceChange string
	Open        string
	High        string
	Low         string
	MktCap      string
	PERatio     string
	DivYield    string
	Week52High  string
	Week52Low   string
}

func ExtractStockBox(doc *goquery.Document) *AnswerBox {
	stockContent := &StockBoxContent{}
	if tickerId := doc.Find("div.iAIpCb.PZPZlf"); tickerId.Length() > 0 {
		tickerParts := strings.Split(tickerId.Text(), ": ")
		if len(tickerParts) == 2 {
			stockContent.TickerId = tickerParts[1]
			stockContent.Exchange = tickerParts[0]
		}
	}

	if stockBox := doc.Find("g-card-section.N9cLBc"); stockBox.Length() > 0 {
		println("Stock Box")

		// Company Name

		if companyName := stockBox.Find("span.aMEhee.PZPZlf"); companyName.Length() > 0 {
			stockContent.CompanyName = companyName.Text()
		}

		// Price
		if price := stockBox.Find("span.IsqQVc.NprOob.wT3VGc"); price.Length() > 0 {
			stockContent.Price = price.Text()
		}

		// Price Change
		if priceChange := stockBox.Find("span.WlRRw.IsqQVc.fw-price-dn"); priceChange.Length() > 0 {
			stockContent.PriceChange = priceChange.Text()
		}

		// Open
		if open := doc.Find("table.CYGKSb tr td.JgXcPd:contains('Open') + td.iyjjgb"); open.Length() > 0 {
			stockContent.Open = open.Text()
		}

		// High
		if high := doc.Find("table.CYGKSb tr td.JgXcPd:contains('High') + td.iyjjgb"); high.Length() > 0 {
			stockContent.High = high.Text()
		}

		// Low
		if low := doc.Find("table.CYGKSb tr td.JgXcPd:contains('Low') + td.iyjjgb"); low.Length() > 0 {
			stockContent.Low = low.Text()
		}

		// Mkt Cap
		if mktCap := doc.Find("table.CYGKSb tr td.JgXcPd:contains('Mkt cap') + td.iyjjgb"); mktCap.Length() > 0 {
			stockContent.MktCap = mktCap.Text()
		}

		// P/E Ratio
		if peRatio := doc.Find("table.CYGKSb tr td.JgXcPd:contains('P/E ratio') + td.iyjjgb"); peRatio.Length() > 0 {
			stockContent.PERatio = peRatio.Text()
		}

		// Div Yield
		if divYield := doc.Find("table.CYGKSb tr td.JgXcPd:contains('Div yield') + td.iyjjgb"); divYield.Length() > 0 {
			stockContent.DivYield = divYield.Text()
		}

		// 52-wk High
		if week52High := doc.Find("table.CYGKSb tr td.JgXcPd:contains('52-wk high') + td.iyjjgb"); week52High.Length() > 0 {
			stockContent.Week52High = week52High.Text()
		}

		// 52-wk Low
		if week52Low := doc.Find("table.CYGKSb tr td.JgXcPd:contains('52-wk low') + td.iyjjgb"); week52Low.Length() > 0 {
			stockContent.Week52Low = week52Low.Text()
		}

		answerBox := &AnswerBox{
			Type:    "stock",
			Content: stockContent,
		}

		return answerBox
	}
	return nil
}
