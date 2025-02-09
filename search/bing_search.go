package search

import (
	"compress/flate"
	"compress/gzip"
	"encoding/json"
	"fmt"
	bingsearch "googlescrapper/bing_search"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/andybalholm/brotli"
	"github.com/gorilla/mux"
)

type BingLink struct {
	Title              string   `json:"title"`
	URL                string   `json:"url"`
	WebsiteName        string   `json:"websiteName"`
	WebsiteAttribution string   `json:"websiteAttribution"`
	Tags               []string `json:"tags"`
	Caption            string   `json:"caption"`
}
type BingInfo struct {
	Links     []BingLink               `json:"links"`
	AnswerBox bingsearch.BingAnswerBox `json:"answer_box"`
}

type BingConfig struct {
	Query string
}

// BingScraper handles the scraping functionality
type BingScraper struct {
	client *http.Client
	config BingConfig
}

// NewBingScraper creates a new scraper instance
func NewBingScraper(config BingConfig) *BingScraper {
	return &BingScraper{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		config: config,
	}
}

func (s *BingScraper) buildBingURL(query string) string {
	return fmt.Sprintf("https://www.bing.com/search?q=%s", url.QueryEscape(query))
}

func (s *BingScraper) BingScrape() (BingInfo, error) {
	req, err := http.NewRequest("GET", s.buildBingURL(s.config.Query), nil)
	if err != nil {
		return BingInfo{}, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:134.0) Gecko/20100101 Firefox/134.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br, zstd")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Cookie", "MUID=3CA33A574615637E3CB62F2A472862EC; MUIDB=3CA33A574615637E3CB62F2A472862EC; _EDGE_V=1; SRCHD=AF=NOFORM; SRCHUID=V=2&GUID=1E8E142F99E24014B924BE3ECCE2AC7E&dmnchg=1; SRCHUSR=DOB=20250123&T=1737985880000; SRCHHPGUSR=SRCHLANG=en&IG=529E3F1F871349569FE1658932A20C0A&DM=1&BRW=M&BRH=M&CW=1282&CH=962&SCW=1270&SCH=3394&DPR=1.0&UTC=330&HV=1737987411&WTS=63873582680&PRVCW=1920&PRVCH=962&EXLTT=10&AV=14&ADV=14&RB=1737658785111&MB=1737658785116; _HPVN=CS=eyJQbiI6eyJDbiI6MiwiU3QiOjAsIlFzIjowLCJQcm9kIjoiUCJ9LCJTYyI6eyJDbiI6MiwiU3QiOjAsIlFzIjowLCJQcm9kIjoiSCJ9LCJReiI6eyJDbiI6MiwiU3QiOjAsIlFzIjowLCJQcm9kIjoiVCJ9LCJBcCI6dHJ1ZSwiTXV0ZSI6dHJ1ZSwiTGFkIjoiMjAyNS0wMS0yN1QwMDowMDowMFoiLCJJb3RkIjowLCJHd2IiOjAsIlRucyI6MCwiRGZ0IjpudWxsLCJNdnMiOjAsIkZsdCI6MCwiSW1wIjo0LCJUb2JuIjowfQ==; _UR=QS=0&TQS=0&Pn=0; USRLOC=HS=1&ELOC=LAT=28.65484619140625|LON=77.1871109008789|N=New%20Delhi%2C%20Delhi|ELT=1|; _RwBf=r=0&ilt=12&ihpd=1&ispd=6&rc=30&rb=0&gb=0&rg=200&pc=27&mtu=0&rbb=0&g=0&cid=&clo=0&v=7&l=2025-01-27T08:00:00.0000000Z&lft=0001-01-01T00:00:00.0000000&aof=0&ard=0001-01-01T00:00:00.0000000&rwdbt=0&rwflt=0&rwaul2=0&o=2&p=&c=&t=0&s=0001-01-01T00:00:00.0000000+00:00&ts=2025-01-27T14:16:51.2816231+00:00&rwred=0&wls=&wlb=&wle=&ccp=&cpt=&lka=0&lkt=0&aad=0&TH=; _EDGE_S=SID=38B4F480F4DE6E911979E101F5C36F45&mkt=en-in; _SS=SID=38B4F480F4DE6E911979E101F5C36F45&R=30&RB=0&GB=0&RG=200&RP=27; ak_bmsc=9508F95A2E0F21E3E46831D0DDB9A0FB~000000000000000000000000000000~YAAQbwHVFy71A2mUAQAAm/EHqBrcueTIIoMhgcnnYpzdJ9NMurBZKZVwfMOOzDn+V+JQEBbAUghKJeNznFnYWgxN0nZqxFbUCvPTMjg4lN9Y61A4aioJr6fFgmVxj39AuhBAj26RqarrMUv6hjc8+WUEuMB4mzYXHuaoStCXi5IbSZViMT2lxRWMa4GSaHgQl8kKyqK+Orj8zgpSVBzdD80+zrzwJCne8U4QrTFuBmKLRoU0Mierl0YeC1IGhVH+iW63zRGUn7BTCq+b2TMLvQE0KZhgXFpKjIGWjMAfM2TQFkE39LJ29/ZD6CKaadYqzlktIeLgphyPjcOVQ4Gxt00SIEhtvURkGwhm5+DKvYUvCuPuJP9dWFf2XcIczfUYi2lnHeofqBcenuRiiRb8dHZX9jSj6u5fAwjKH8bsjA==; _Rwho=u=d&ts=2025-01-27; ipv6=hit=1737989484201&t=6; _C_ETH=1")
	req.Header.Set("Referer", "https://www.bing.com/")
	req.Header.Set("Host", "www.bing.com")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Priority", "u=0, i")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("TE", "trailers")

	resp, err := s.client.Do(req)
	if err != nil {
		return BingInfo{}, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return BingInfo{}, fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
	}

	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			return BingInfo{}, fmt.Errorf("failed to create gzip reader: %v", err)
		}
		defer reader.Close()
	case "deflate":
		reader = flate.NewReader(resp.Body)
		defer reader.Close()
	case "br":
		reader = io.NopCloser(brotli.NewReader(resp.Body))
	default:
		reader = resp.Body
	}

	body, err := io.ReadAll(reader) // write to a html file
	ioutil.WriteFile("bing.html", body, 0644)

	if err != nil {
		return BingInfo{}, fmt.Errorf("failed to read response body: %v", err)
	}

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return BingInfo{}, fmt.Errorf("failed to parse HTML: %v", err)
	}

	var BingLinks []BingLink
	var BingInfos BingInfo
	doc.Find("li.b_algo").Each(func(i int, s *goquery.Selection) {
		title := s.Find("h2").Text()
		url, exists := s.Find("a").Attr("href")
		if !exists {
			return
		}

		websiteName := s.Find("div.tptt").Text()
		websiteAttribution := s.Find("div.b_attribution cite").Text()
		caption := s.Find("div.b_caption p").Text()

		var tags []string
		s.Find(".tltg").Each(func(i int, tag *goquery.Selection) {
			tags = append(tags, tag.Text())
		})

		// extractedUrl, err := utils.GetRedirectedURL(url)

		BingLinks = append(BingLinks, BingLink{
			Title:              title,
			URL:                url,
			WebsiteName:        websiteName,
			WebsiteAttribution: websiteAttribution,
			Tags:               tags,
			Caption:            caption,
		})
	})
	AnswerBox := bingsearch.ExtractAnswerbox(doc)

	BingInfos.Links = BingLinks
	BingInfos.AnswerBox = *AnswerBox
	return BingInfos, nil
}

func StandardBingHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	query := vars["query"]
	println(query)
	if query == "" {
		http.Error(w, "Query parameter is required", http.StatusBadRequest)
		return
	}

	config := BingConfig{
		Query: query,
	}

	scraper := NewBingScraper(config)

	BingInfos, err := scraper.BingScrape()

	if err != nil {
		println(err.Error())
		http.Error(w, "Error scraping results", http.StatusInternalServerError)
		return
	}

	jsonData, err := json.MarshalIndent(BingInfos, "", "    ")
	if err != nil {
		http.Error(w, "Error marshaling to JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}
