package search

import (
	"compress/flate"
	"compress/gzip"
	"encoding/json"
	"fmt"
	"googlescrapper/config"
	"googlescrapper/standard_search"
	"googlescrapper/utils"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/andybalholm/brotli"

	"github.com/PuerkitoBio/goquery"
	"github.com/gorilla/mux"
)

// SearchResult represents a single search result
type SearchResult struct {
	Title   string `json:"title"`
	Content string `json:"content"`
	URL     string `json:"url"`
	Favicon string `json:"favicon"`
}

// SearchResponse represents the complete search response
type SearchResponse struct {
	Links             []standard_search.SearchResult     `json:"links,omitempty"`
	AnswerBox         standard_search.AnswerBox          `json:"answer_box,omitempty"`
	SuggestedProducts []standard_search.SuggestedProduct `json:"suggested_products,omitempty"`
}

// SearchConfig holds the search parameters
type SearchConfig struct {
	Query      string
	Location   string
	Language   string
	MaxResults int
	Latitude   *float64 // Optional latitude
	Longitude  *float64 // Optional longitude
}

// SearchScraper handles the scraping functionality
type SearchScraper struct {
	client *http.Client
	config SearchConfig
}

// NewSearchScraper creates a new scraper instance
func NewSearchScraper(config SearchConfig) *SearchScraper {
	return &SearchScraper{
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
		config: config,
	}
}

// buildSearchURL creates the search URL with parameters
func (s *SearchScraper) buildSearchURL() string {
	params := url.Values{}
	params.Add("q", s.config.Query)

	if regionConfig, ok := config.RegionConfigs[s.config.Location]; ok {
		params.Add("gl", regionConfig.Gl)
		params.Add("lr", regionConfig.Lr)
		params.Add("hl", regionConfig.Hl)
	}

	// Add coordinates if both latitude and longitude are provided
	if s.config.Latitude != nil && s.config.Longitude != nil {
		// Add location bias parameter
		params.Add("geoloc", fmt.Sprintf("%f,%f", *s.config.Latitude, *s.config.Longitude))
		// Add additional location parameter used by Google
		params.Add("uule", utils.CreateUULE(*s.config.Latitude, *s.config.Longitude))
	}

	return "https://www.google.com/search?" + params.Encode()
}

func (s *SearchScraper) Scrape() (*SearchResponse, error) {
	req, err := http.NewRequest("GET", s.buildSearchURL(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %v", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:134.0) Gecko/20100101 Firefox/134.0")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	req.Header.Set("Accept-Encoding", "gzip, deflate, br, zstd")
	req.Header.Set("Connection", "keep-alive")
	req.Header.Set("Upgrade-Insecure-Requests", "1")
	req.Header.Set("Sec-Fetch-Dest", "document")
	req.Header.Set("Sec-Fetch-Mode", "navigate")
	req.Header.Set("Sec-Fetch-Site", "none")
	req.Header.Set("Sec-Fetch-User", "?1")
	req.Header.Set("Priority", "u=0, i")
	req.Header.Set("Pragma", "no-cache")
	req.Header.Set("Cookie", "AEC=AZ6Zc-WK4w0I45M285ZfXoJ1fRgAkOy8qVo_BpkVRWOIMWpMfuMH9T38aA; NID=521=VLg5bs_FNg3sHkrCHoEM8hLLt3P7rseIb4Zk35R_e3XysJ0k2xtV9H0mJp1SiGCP-zTx9ThpCqTCl2xkhJeBbGzZZYzvxzIzZR7EFvuxsKAPbl6lK170Zx_c1RWNJa-bXKtZTvdgZ_175Yq_aTimCIktRhM8OgoT_JPsc9_NQob9IBhwIKIgi-1ocOo4mIwENjEsJWVMSRJ8zZEJepUPRzuEntvZ04ET8JNrxmCt679lfR3VWBpRcUBoby6G5OAVT1-ZhRBQU9EAiFz6VRBc7_og_RgD-A9-kXL9kT8M8DQC-t61L9tk9KlUR3X8IeqLFOvYNdErxkKaxHjbjJ00pjYRh43vpW6BN5_GyWnBgCpOQ_ikiWmNDz6vIRZ84S95cyroPhWlfDrvLoFMEMpQH9sp4dd2cE3d0JZmNe0Qk8u6b5zQMFr15aEFG26Dth99RsWuEkjLtxr7U8LPtQWkTxJM9eLBOMIJTrrkw8TPQ7_x8VrXiYx7laCTm3zGh0Z_ow_Ntb4Qtv4e56Z8pnPKVirxkX8K9yHW3K8ntAw64EBsiD86PfxHu5GTw-Mk_oD0DDmLh27BYicM43j-yNBjY7KaJrLiEOjhUb9dVFLFZpSsOCkk13z6DxrYPciVaTIQDrAXIfYQ3ySIg22UzxHvAQAWE_hieUB6pk6Sfj9zsE7BXJCq3dGg075DF8E06JQ8N9Iy_kttgkfEAjnTjZ2WDQgouR9A6k_InQMwP7RDjQWNt3F4ftw9P-dsn37b6UifCmrOv2Y5TLSzEVfRT-aqwVpTbnA-zxnV4a0cbQD_FUskuFsAlbJH2EOM2u1vWS6UtvauReo4r63qBl-OvNYwrtF4koo-y1fYgKLR7IfRV52hEvtOAusEABg0ZZyWcxzoWvZHuf5E62IUFJWs-NXDM3Otija7dwsK7wD5bpdRyaqrOF9zN-1rBF7ROlWXg3__iNqGLl5qn17FhEvOCSVzOZRqYExmFYVgWynxlmSqIr1NnhykX3N1ruxvrxQFnxnpCL8Fzte2lDeT2I8l4B8lf8ByVbADs2WmqOGTasmsISZdFM1J5Ikq2tJ4s64e1wwOkIA1dz5ab2QriUZxhp9wyjv3IBmWqxokzCmdtb-sIG_pp8i2bNFpXuOB4sZ3F6efzl1AApHyTndyEAltbMc8hFsF4sPs3TPlu-vpP1JNVaLH0IBJDWeabGXCnJxL7FGPDHUrxuFLDMAUwFltadqSLFzaaH547eLumUoQxLhq88wnekos6Hg5rR3BjsihvUdVuBPZ_hMfIJighpJYarC3L_gySBIgqYTK7zei6VNkcJoEvWi2jKMWbRgY0ZN8xHkfkQE_NpH67KQrSwDbTuaBEFxJ-7eGRazyPkjn4ApkTZCMW-i0MIByosRQqIEQOtCJd6ggW7Nig3GOkw6opJEtQ5K5xyljXEifpiBRlAJFbO-LjO01cGVo5s5BsgOLLaUklsXwb6j2YmRPRTVgwFRwogtkFp37WhR_tiY8eAEZoqLRUau5tXHL36Ry1wMzTsR3KRhCDooPmDO41KYMFhkOqxb2PrSSHZRTPpcngLdJ3eoDBr9kEZV9XWWgYqawyJ_4MWSJNFU1n4nEx3iKPFAO7lgdCtbL-E88jBDgDw07wpJafB4mhQc1WXLettEMNXdra0d0HKoDKaTJU9f1NclGOrENdRQI4T_cXEuFv3UVuqDhIF6C_7DL9y5eaKs5GlwRtG_toI9ibj5qHm270rxMLdBnxu2OThklLhiM2wYkgWFXlJ9gQMLLeEtbtmIpEiT22uj7taFnaZY_eAuaiRm5OJjypUWqWqggoZrMOlBbTcel9SItpl9eLfgy6vjEMNPneopJQ-LXOMz0psr16XIvNeaMAZ3jx1xk-xN7pJmWntkYkzRG45ygR91epklynkUyPz-S3lbAKHILH0Z1bzCBDAGT4ELUkZgdRzSvlt1_72Q8yOZE3LfCyJx6y4h3gos-6d874r8nvctKIilcEsfOaqq-h9vEXP7ATlppTHWJRoj7rZjeUB0wOH6WIJlPAP9WWErVHcv5gu8t1IMjTZeEcxHQZrlAZVBh9KQgwu5hsexDUV_xCSqR4EZWOn_eq5sIh8e-56JcDt-3f64DhzGm8WWgL8hxezTXLojlE_6jAyWrLWY12vNx5sQNR-3kFF_wD-FQEeOmMQXx9B2IzFTL9QtweJIq_Nrrj9bl0zHvhQVK1dxOS9aW_8yz-7uDzWS7p6BCCUitr6tqZtQ8RXUSj8hkORcyd2WjDr_PFvPjM_ZROx61Lr2xxphd; SID=g.a000swiYNjmVt85bR1pNw41eoFmh0CfFHcddr2igE34BR7EjkhetjYEwQczjcKWZWXRC7cTaRwACgYKAQwSAQASFQHGX2MitB87j8HwfiA9J34AS-pUKxoVAUF8yKoPA3kgVhR9nvo11anaeK5V0076; __Secure-1PSID=g.a000swiYNjmVt85bR1pNw41eoFmh0CfFHcddr2igE34BR7EjkhetUokeljSRDhbEu-6AUHadsgACgYKAbsSAQASFQHGX2MiqeGCcr2VsfhISugWEv_2jRoVAUF8yKq0qt-jB1hAEDTqm7LkVHB70076; __Secure-3PSID=g.a000swiYNjmVt85bR1pNw41eoFmh0CfFHcddr2igE34BR7EjkhetY0VR5Vmff1oq1DkgjwYPyAACgYKAXESAQASFQHGX2Mi3mPTrIz5FOwJ-N0QbHkB2xoVAUF8yKr2Stn39c7eQMNDsvDpIej-0076; HSID=A_P9EBOqNvh7xKEF1; SSID=ACpNUBvUNjkHtNjGP; APISID=Ro0iWdqXzjfcJ4yh/APUUT9PYsBBGDUEsG; SAPISID=xdeH_DGDWkcj5OqV/ALx3l7EaMbTPWExK8; __Secure-1PAPISID=xdeH_DGDWkcj5OqV/ALx3l7EaMbTPWExK8; __Secure-3PAPISID=xdeH_DGDWkcj5OqV/ALx3l7EaMbTPWExK8; SIDCC=AKEyXzW_KMiw3LlqHME0yuV5TBqRXGaiLUgeWPh_qtt-EqKZYFJoHh4esRvTGRrJzAITFKQCpxwk; __Secure-1PSIDCC=AKEyXzWA_Qey41HbSCWmjEdIKnwvTeU3AeWlAL26z7C7imq-0EjBIqa1QJn04rQapJKt8Cy2bBg; __Secure-3PSIDCC=AKEyXzX_7Lrs8njr5QJlOEy2qAl6Wszj211h1JGBJLMikOmGbIHDnDPUSAmRD0PIyYqvxP1lNfPP; SEARCH_SAMESITE=CgQIkp0B; __Secure-1PSIDTS=sidts-CjEBmiPuTZH4SWHV9vDr62vqLG8lcUuUHZO1wXrGP2QLg9SffxqI-ScVCj9AbLISeEq1EAA; __Secure-3PSIDTS=sidts-CjEBmiPuTZH4SWHV9vDr62vqLG8lcUuUHZO1wXrGP2QLg9SffxqI-ScVCj9AbLISeEq1EAA; OTZ=7904857_34_34__34_; S=billing-ui-v3=V8On6v3WAmtIJheJWRXUX2rTlLN99Ehf:billing-ui-v3-efe=V8On6v3WAmtIJheJWRXUX2rTlLN99Ehf; DV=o0UuqGbBFV9W8JiFc16AWAY-qK97ShktVbcxKQal2AEAAIBHfY3ypUEuxAAAAPCSFA-qpoSSOgAAAF0qnu00FCAWDwAAAA")
	req.Header.Set("Cache-Control", "no-cache")
	req.Header.Set("TE", "trailers")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %v", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("received non-200 status code: %d", resp.StatusCode)
	}

	var reader io.ReadCloser
	switch resp.Header.Get("Content-Encoding") {
	case "gzip":
		reader, err = gzip.NewReader(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("failed to create gzip reader: %v", err)
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

	body, err := io.ReadAll(reader)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %v", err)
	}

	// write to file
	ioutil.WriteFile("response.html", body, 0644)

	doc, err := goquery.NewDocumentFromReader(strings.NewReader(string(body)))
	if err != nil {
		return nil, fmt.Errorf("failed to parse HTML: %v", err)
	}

	searchResponse := &SearchResponse{
		Links:             []standard_search.SearchResult{},
		AnswerBox:         standard_search.AnswerBox{},
		SuggestedProducts: []standard_search.SuggestedProduct{},
	}

	extractedResults := standard_search.ExtractSearchResults(doc, s.config.MaxResults)
	searchResponse.Links = extractedResults

	// Extract answer box
	answerBox := standard_search.ExtractAnswerbox(doc)
	if answerBox != nil {
		searchResponse.AnswerBox = *answerBox
	}
	suggestedProducts := standard_search.ExtractSuggestedProducts(doc)
	searchResponse.SuggestedProducts = suggestedProducts

	return searchResponse, nil
}

func StandardSearchHandler(w http.ResponseWriter, r *http.Request) {
	vars := mux.Vars(r)
	query := vars["query"]
	location := vars["location"]
	maxResults, err := strconv.Atoi(vars["maxResults"])
	if err != nil {
		http.Error(w, "Invalid maxResults parameter", http.StatusBadRequest)
		return
	}

	lat, err := strconv.ParseFloat(vars["latitude"], 64)
	if err != nil {
		http.Error(w, "Invalid latitude parameter", http.StatusBadRequest)
		return
	}

	lon, err := strconv.ParseFloat(vars["longitude"], 64)
	if err != nil {
		http.Error(w, "Invalid longitude parameter", http.StatusBadRequest)
		return
	}

	useCoords := vars["useCoords"] == "true"

	if query == "" {
		http.Error(w, "Query parameter is required", http.StatusBadRequest)
		return
	}

	// Validate region
	if _, ok := config.RegionConfigs[location]; !ok {
		http.Error(w, "Invalid region code", http.StatusBadRequest)
		return
	}

	config := SearchConfig{
		Query:      query,
		Location:   location,
		MaxResults: maxResults,
	}

	if useCoords {
		config.Latitude = &lat
		config.Longitude = &lon
	}

	scraper := NewSearchScraper(config)

	searchResponse, err := scraper.Scrape()
	if err != nil {
		println(err.Error())
		http.Error(w, "Error scraping results", http.StatusInternalServerError)
		return
	}

	jsonData, err := json.MarshalIndent(searchResponse, "", "    ")
	if err != nil {
		http.Error(w, "Error marshaling to JSON", http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.Write(jsonData)
}
