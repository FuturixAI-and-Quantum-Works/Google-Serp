package stock

import (
	"encoding/json"
	"fmt"
	"googlescrapper/cache"
	"net/http"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/mux"
)

// Structure for API Response
type StockData struct {
	PriceInfo         PriceInfo         `json:"priceInfo"`
	Performance       Performance       `json:"performance"`
	KeyMetrics        KeyMetrics        `json:"keyMetrics"`
	RelatedNews       []NewsArticle     `json:"relatedNews"`
	CompanyDetails    CompanyDetails    `json:"companyDetails"`
	Financials        Financials        `json:"financials"`
	Shareholding      Shareholding      `json:"shareholding"`
	TechnicalAnalysis TechnicalAnalysis `json:"technicalAnalysis"`
	Peers             []Peer            `json:"peers"`
	FAQs              []FAQ             `json:"faqs"`
}

type PriceInfo struct {
	CurrentPrice      string `json:"currentPrice"`
	ChangePercent     string `json:"changePercent"`
	TradingVolume     string `json:"tradingVolume"`
	DayRange          Range  `json:"dayRange"`
	FiftyTwoWeekRange Range  `json:"fiftyTwoWeekRange"`
}

type Range struct {
	Low  string `json:"low"`
	High string `json:"high"`
}

type Performance struct {
	YearToDate    string `json:"yearToDate"`
	FiveDayChange string `json:"fiveDayChange"`
	AnalystRating string `json:"analystRating"`
}

type KeyMetrics struct {
	DividendYield string `json:"dividendYield"`
	PB            string `json:"pb"`
	PE            string `json:"pe"`
	Beta          string `json:"beta"`
	DebtToEquity  string `json:"debtToEquity"`
}

type NewsArticle struct {
	Title    string `json:"title"`
	Link     string `json:"link"`
	ReadTime string `json:"readTime"`
	Time     string `json:"time"`
}

type CompanyDetails struct {
	Industry    string           `json:"industry"`
	ISIN        string           `json:"isin"`
	BSECode     string           `json:"bseCode"`
	NSECode     string           `json:"nseCode"`
	Description string           `json:"description"`
	Management  []ManagementInfo `json:"management"`
}

type ManagementInfo struct {
	Name  string `json:"name"`
	Title string `json:"title"`
}

type Financials struct {
	LatestQuarterProfit string   `json:"latestQuarterProfit"`
	RevenueGrowth       string   `json:"revenueGrowth"`
	ProfitGrowth        string   `json:"profitGrowth"`
	Insights            []string `json:"insights"`
}

type Shareholding struct {
	PromoterHolding string   `json:"promoterHolding"`
	FIIHolding      string   `json:"fiiHolding"`
	MFHolding       string   `json:"mfHolding"`
	Insights        []string `json:"insights"`
}

type TechnicalAnalysis struct {
	PivotLevels PivotLevels `json:"pivotLevels"`
}

type PivotLevels struct {
	R1    string `json:"r1"`
	R2    string `json:"r2"`
	R3    string `json:"r3"`
	Pivot string `json:"pivot"`
	S1    string `json:"s1"`
	S2    string `json:"s2"`
	S3    string `json:"s3"`
}

type Peer struct {
	Name            string `json:"name"`
	TechnicalRating string `json:"technicalRating"`
	Price           string `json:"price"`
	ChangePercent   string `json:"changePercent"`
	MarketCap       string `json:"marketCap"`
	PE              string `json:"pe"`
	PB              string `json:"pb"`
	DividendYield   string `json:"dividendYield"`
	DebtToEquity    string `json:"debtToEquity"`
}

type FAQ struct {
	Question string `json:"question"`
	Answer   string `json:"answer"`
}

func ScrapeStockData(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	stockIdentifier := vars["stockIdentifier"]

	cacheKey := fmt.Sprintf("stock-data:%s", stockIdentifier)

	result, err := cache.Memoize(cacheKey, 12*time.Hour, func() (interface{}, error) {
		liveMindTickerData, err := FetchStockTickerData(stockIdentifier)
		if err != nil {
			return nil, err
		}

		if len(liveMindTickerData) == 0 {
			return nil, fmt.Errorf("no data found for stock: %s", stockIdentifier)
		}

		livemintTicker := liveMindTickerData[0]
		identifier := fmt.Sprintf("stocks-%s-share-price-nse-bse-%s",
			strings.ReplaceAll(strings.ToLower(livemintTicker.CommonName), " ", "-"),
			livemintTicker.ID,
		)

		url := fmt.Sprintf("https://www.livemint.com/market/market-stats/%s", strings.ToLower(identifier))
		res, err := http.Get(url)
		if err != nil {
			return nil, fmt.Errorf("failed to fetch webpage: %v", err)
		}
		defer res.Body.Close()

		if res.StatusCode != 200 {
			return nil, fmt.Errorf("failed to fetch webpage, status code: %d", res.StatusCode)
		}

		doc, err := goquery.NewDocumentFromReader(res.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to parse HTML: %v", err)
		}

		// Price Info
		priceInfo := PriceInfo{
			CurrentPrice:  cleanText(doc.Find(".storyDetails_stockPriceInfo__b5EzL strong").Text()),
			ChangePercent: cleanText(doc.Find(".storyDetails_stockPriceInfo__b5EzL span").Text()),
			TradingVolume: cleanText(doc.Find(".storyDetails_stockVol__iqNEP").Text()),
			DayRange: Range{
				Low:  cleanText(doc.Find(".storyDetails_rangeBlock__BVF9Y:nth-child(1) .storyDetails_lowValue__ubfAY").First().Text()),
				High: cleanText(doc.Find(".storyDetails_rangeBlock__BVF9Y:nth-child(1) .storyDetails_lowValue__ubfAY").Last().Text()),
			},
			FiftyTwoWeekRange: Range{
				Low:  cleanText(doc.Find(".storyDetails_rangeBlock__BVF9Y:nth-child(2) .storyDetails_lowValue__ubfAY").First().Text()),
				High: cleanText(doc.Find(".storyDetails_rangeBlock__BVF9Y:nth-child(2) .storyDetails_lowValue__ubfAY").Last().Text()),
			},
		}

		// Performance
		perfText := doc.Find(".storyDetails_readMore__K4Tyd").Text()
		performance := Performance{
			YearToDate:    extractPerformance(perfText, "has given", "% in this year"),
			FiveDayChange: extractPerformance(perfText, "&", "% in the last 5 days"),
			AnalystRating: extractAnalystRating(doc),
		}

		// Key Metrics
		keyMetrics := KeyMetrics{
			DividendYield: cleanMetric(doc.Find(".storyDetails_info__i53sk li:nth-child(1) .storyDetails_databox__ejMc4 span").Eq(1).Text()),
			PB:            cleanMetric(doc.Find(".storyDetails_info__i53sk li:nth-child(2) .storyDetails_databox__ejMc4 span").Eq(1).Text()),
			PE:            cleanMetric(doc.Find(".storyDetails_info__i53sk li:nth-child(3) .storyDetails_databox__ejMc4 span").Eq(1).Text()),
			Beta:          cleanMetric(doc.Find(".storyDetails_info__i53sk li:nth-child(4) .storyDetails_databox__ejMc4 span").Eq(1).Text()),
			DebtToEquity:  cleanMetric(doc.Find(".storyDetails_info__i53sk li:nth-child(5) .storyDetails_databox__ejMc4 span").Eq(1).Text()),
		}

		// Related News
		var news []NewsArticle
		doc.Find(".storyDetails_newssecBlock__f2S5c li").Each(func(i int, s *goquery.Selection) {
			title := cleanText(s.Find(".storyDetails_newsTitle__dxnD5 a").Text())
			link, _ := s.Find(".storyDetails_newsTitle__dxnD5 a").Attr("href")
			readTime := cleanText(s.Find(".storyDetails_dateStock__lQbRz span").First().Text())
			time := cleanText(s.Find(".storyDetails_dateStock__lQbRz span").Last().Text())

			if title != "" && link != "" {
				news = append(news, NewsArticle{
					Title:    title,
					Link:     link,
					ReadTime: readTime,
					Time:     time,
				})
			}
		})

		// Company Details
		var management []ManagementInfo
		doc.Find(".storyDetails_audiName__uAfOP").Each(func(i int, s *goquery.Selection) {
			// Get the full text content before the span
			// Extract the name and title
			name := cleanText(s.Contents().Not("span").Text())
			title := cleanText(s.Find("span").Text())

			if name != "" && title != "" {
				management = append(management, ManagementInfo{
					Name:  name,
					Title: title,
				})
			}
		})

		company := CompanyDetails{
			Industry:    cleanText(doc.Find(".storyDetails_compInfo__wSZUv li:contains('Industry')").Contents().Last().Text()),
			ISIN:        cleanText(doc.Find(".storyDetails_compInfo__wSZUv li:contains('ISIN')").Contents().Last().Text()),
			BSECode:     cleanText(doc.Find(".storyDetails_compInfo__wSZUv li:contains('BSE Code')").Contents().Last().Text()),
			NSECode:     cleanText(doc.Find(".storyDetails_compInfo__wSZUv li:contains('NSE Code')").Contents().Last().Text()),
			Description: cleanText(doc.Find(".storyDetails_compDesc__TvfHP").Text()),
			Management:  management,
		}

		// Financials
		var finInsights []string
		doc.Find(".storyDetails_finInsight__a7T40 .storyDetails_insightListing__oQZ_I li").Each(func(i int, s *goquery.Selection) {
			finInsights = append(finInsights, cleanText(s.Text()))
		})
		financials := Financials{
			LatestQuarterProfit: extractFinancial(perfText, "net profit of", "Crores"),
			RevenueGrowth:       extractFinancial(perfText, "revenue grew in December quarter by", "%"),
			ProfitGrowth:        extractFinancial(perfText, "profit by", "%"),
			Insights:            finInsights,
		}

		// Shareholding
		var shareInsights []string
		doc.Find(".storyDetails_shareholding__IHGZ1 .storyDetails_insightListing__oQZ_I li").Each(func(i int, s *goquery.Selection) {
			shareInsights = append(shareInsights, cleanText(s.Text()))
		})
		shareholding := Shareholding{
			PromoterHolding: extractShareholding(perfText, "Promoter(s) holding is moderate at", "%"),
			FIIHolding:      cleanText(doc.Find("#stockFIIholding").Text()),
			MFHolding:       cleanText(doc.Find("#stockMFholding").Text()),
			Insights:        shareInsights,
		}

		// Technical Analysis
		pivot := TechnicalAnalysis{
			PivotLevels: PivotLevels{
				R1:    cleanText(doc.Find(".storyDetails_pBox__6Ap93:nth-child(1) li:nth-child(2)").Text()),
				R2:    cleanText(doc.Find(".storyDetails_pBox__6Ap93:nth-child(1) li:nth-child(4)").Text()),
				R3:    cleanText(doc.Find(".storyDetails_pBox__6Ap93:nth-child(1) li:nth-child(6)").Text()),
				Pivot: cleanText(doc.Find(".storyDetails_pBox__6Ap93:nth-child(2) strong").Text()),
				S1:    cleanText(doc.Find(".storyDetails_pBox__6Ap93:nth-child(3) li:nth-child(2)").Text()),
				S2:    cleanText(doc.Find(".storyDetails_pBox__6Ap93:nth-child(3) li:nth-child(4)").Text()),
				S3:    cleanText(doc.Find(".storyDetails_pBox__6Ap93:nth-child(3) li:nth-child(6)").Text()),
			},
		}

		// Peers
		var peers []Peer
		doc.Find(".storyDetails_tableData__Hsmic tbody tr").Each(func(i int, s *goquery.Selection) {
			name := cleanText(s.Find("td:nth-child(1) a").Text())
			if name != "" {
				peer := Peer{
					Name:            name,
					TechnicalRating: cleanText(s.Find("td:nth-child(2) .storyDetails_labelSec__nSEQM").Text()),
					Price:           cleanText(s.Find("td:nth-child(3)").Text()),
					ChangePercent:   cleanText(s.Find("td:nth-child(4)").Text()),
					MarketCap:       cleanText(s.Find("td:nth-child(5)").Text()),
					PE:              cleanText(s.Find("td:nth-child(6)").Text()),
					PB:              cleanText(s.Find("td:nth-child(7)").Text()),
					DividendYield:   cleanText(s.Find("td:nth-child(8)").Text()),
					DebtToEquity:    cleanText(s.Find("td:nth-child(9)").Text()),
				}
				peers = append(peers, peer)
			}
		})

		// FAQs
		var faqs []FAQ
		doc.Find(".storyDetails_accordionTab__1L_sn").Each(func(i int, s *goquery.Selection) {
			question := cleanText(s.Find(".storyDetails_accordionTab-label__Cb52o").Text())
			answer := cleanText(s.Find(".storyDetails_accordionTab-content__rCfVm").Text())
			if question != "" && answer != "" {
				faqs = append(faqs, FAQ{
					Question: question,
					Answer:   answer,
				})
			}
		})

		response := StockData{
			PriceInfo:         priceInfo,
			Performance:       performance,
			KeyMetrics:        keyMetrics,
			RelatedNews:       news,
			CompanyDetails:    company,
			Financials:        financials,
			Shareholding:      shareholding,
			TechnicalAnalysis: pivot,
			Peers:             peers,
			FAQs:              faqs,
		}

		return response, nil
	})

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(result)
}

// Helper functions
func cleanText(text string) string {
	text = strings.TrimSpace(text)
	text = strings.ReplaceAll(text, "\n", " ")
	text = strings.ReplaceAll(text, "\t", " ")
	return strings.TrimSpace(text)
}

func cleanMetric(text string) string {
	text = cleanText(text)
	if text == "-" || text == "" {
		return "N/A"
	}
	return text
}

func extractPerformance(text, start, end string) string {
	startIdx := strings.Index(text, start)
	if startIdx == -1 {
		return "N/A"
	}
	startIdx += len(start)
	endIdx := strings.Index(text[startIdx:], end)
	if endIdx == -1 {
		return "N/A"
	}
	return cleanText(text[startIdx : startIdx+endIdx])
}

func extractFinancial(text, start, end string) string {
	startIdx := strings.Index(text, start)
	if startIdx == -1 {
		return "N/A"
	}
	startIdx += len(start)
	endIdx := strings.Index(text[startIdx:], end)
	if endIdx == -1 {
		return "N/A"
	}
	return cleanText(text[startIdx : startIdx+endIdx])
}

func extractShareholding(text, start, end string) string {
	startIdx := strings.Index(text, start)
	if startIdx == -1 {
		return "N/A"
	}
	startIdx += len(start)
	endIdx := strings.Index(text[startIdx:], end)
	if endIdx == -1 {
		return "N/A"
	}
	return cleanText(text[startIdx : startIdx+endIdx])
}

func extractAnalystRating(doc *goquery.Document) string {
	var rating string
	doc.Find(".faqBullets li").Each(func(i int, s *goquery.Selection) {
		text := cleanText(s.Text())
		if strings.Contains(text, "strong buy") || strings.Contains(text, "buy") ||
			strings.Contains(text, "hold") || strings.Contains(text, "sell") {
			rating += text + "; "
		}
	})
	if rating == "" {
		return "N/A"
	}
	return strings.TrimSpace(rating)
}

// Note: FetchStockTickerData is assumed to be defined elsewhere
// You'll need to ensure this function exists and returns the expected data structure
