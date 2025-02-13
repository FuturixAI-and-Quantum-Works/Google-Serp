package stock

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/gorilla/mux"
)

// StockForecastResponse represents the response structure
// Modify this struct to include relevant fields from the response.json file

type StockForecastResponse struct {
	TickerMasterID         string      `json:"tickerMasterId"`
	ExchangeType           interface{} `json:"exchangeType"`
	StockForecastsResponse struct {
		GetEstimateSummariesResponse1 struct {
			EstimateSummaries struct {
				EstimateSummary []struct {
					Organization struct {
						Name          string `json:"Name"`
						DisplaySymbol struct {
							DisplayValue string `json:"DisplayValue"`
							Symbol       struct {
								Type  string `json:"Type"`
								Value string `json:"Value"`
							} `json:"Symbol"`
						} `json:"DisplaySymbol"`
						Industry struct {
							Code struct {
								Type  string `json:"Type"`
								Value string `json:"Value"`
							} `json:"Code"`
							Name string `json:"Name"`
						} `json:"Industry"`
						ValueFormats struct {
							ValueFormat []struct {
								Type         string `json:"Type"`
								Name         string `json:"Name"`
								Scale        int    `json:"Scale"`
								Unit         string `json:"Unit"`
								CurrencyCode string `json:"CurrencyCode"`
							} `json:"ValueFormat"`
						} `json:"ValueFormats"`
						PreferredMeasureCode      string `json:"PreferredMeasureCode"`
						PrimaryReportingBasis     string `json:"PrimaryReportingBasis"`
						HasMultipleReportingBases bool   `json:"HasMultipleReportingBases"`
						CountryCode               string `json:"CountryCode"`
						CountryName               string `json:"CountryName"`
					} `json:"Organization"`
					Measures struct {
						Measure []struct {
							Code           string `json:"Code"`
							Name           string `json:"Name"`
							ReportingBasis string `json:"ReportingBasis"`
							ValueFormat    struct {
								Type         string `json:"Type"`
								Name         string `json:"Name"`
								Scale        int    `json:"Scale"`
								Unit         string `json:"Unit"`
								CurrencyCode string `json:"CurrencyCode"`
							} `json:"ValueFormat"`
							Section                         string `json:"Section"`
							PrecisionDetail                 int    `json:"PrecisionDetail"`
							PrecisionSummary                int    `json:"PrecisionSummary"`
							LongTermGrowthEstimateSnapshots struct {
								LongTermGrowthEstimateSnapshot []struct {
									Mean              interface{} `json:"Mean"`
									High              interface{} `json:"High"`
									Low               interface{} `json:"Low"`
									Median            interface{} `json:"Median"`
									NumberOfEstimates interface{} `json:"NumberOfEstimates"`
									StandardDeviation interface{} `json:"StandardDeviation"`
									Age               string      `json:"Age"`
								} `json:"LongTermGrowthEstimateSnapshot"`
							} `json:"LongTermGrowthEstimateSnapshots,omitempty"`
							Periods struct {
								Period []struct {
									RelativePeriod struct {
										Type   string `json:"Type"`
										Number int    `json:"Number"`
									} `json:"RelativePeriod"`
									FiscalPeriod struct {
										Type string `json:"Type"`
										Year int    `json:"Year"`
									} `json:"FiscalPeriod"`
									CalendarYear     int    `json:"CalendarYear"`
									CalendarMonth    int    `json:"CalendarMonth"`
									ActualReportDate string `json:"ActualReportDate"`
									Actuals          struct {
										Actual []struct {
											CurrencyCode                      string      `json:"CurrencyCode"`
											Reported                          float64     `json:"Reported"`
											ReportedDate                      string      `json:"ReportedDate"`
											RestatedDate                      interface{} `json:"RestatedDate"`
											Restated                          interface{} `json:"Restated"`
											PostReport30DayPriceChangePercent interface{} `json:"PostReport30DayPriceChangePercent"`
											SurprisePercent                   float64     `json:"SurprisePercent"`
											Surprise60DayPercent              interface{} `json:"Surprise60DayPercent"`
											SurpriseMean                      float64     `json:"SurpriseMean"`
											StandardizedUnexpectedEarnings    float64     `json:"StandardizedUnexpectedEarnings"`
											NumberOfEstimates                 int         `json:"NumberOfEstimates"`
										} `json:"Actual"`
									} `json:"Actuals"`
									Estimates         interface{} `json:"Estimates"`
									EstimateSnapshots interface{} `json:"EstimateSnapshots"`
								} `json:"Period"`
							} `json:"Periods"`
						} `json:"Measure"`
					} `json:"Measures"`
					PriceTarget struct {
						CurrencyCode      string `json:"CurrencyCode"`
						UnverifiedMean    int    `json:"UnverifiedMean"`
						PreliminaryMean   int    `json:"PreliminaryMean"`
						High              int    `json:"High"`
						Low               int    `json:"Low"`
						Mean              int    `json:"Mean"`
						Median            int    `json:"Median"`
						NumberOfEstimates int    `json:"NumberOfEstimates"`
						StandardDeviation int    `json:"StandardDeviation"`
					} `json:"PriceTarget"`
					PriceTargetSnapshots struct {
						PriceTargetSnapshot []struct {
							CurrencyCode      string `json:"CurrencyCode"`
							High              int    `json:"High"`
							Low               int    `json:"Low"`
							Mean              int    `json:"Mean"`
							Median            int    `json:"Median"`
							NumberOfEstimates int    `json:"NumberOfEstimates"`
							StandardDeviation int    `json:"StandardDeviation"`
							Age               string `json:"Age"`
						} `json:"PriceTargetSnapshot"`
					} `json:"PriceTargetSnapshots"`
					Recommendation struct {
						UnverifiedMean          float64 `json:"UnverifiedMean"`
						PreliminaryMean         float64 `json:"PreliminaryMean"`
						Mean                    float64 `json:"Mean"`
						High                    int     `json:"High"`
						Low                     int     `json:"Low"`
						NumberOfRecommendations int     `json:"NumberOfRecommendations"`
						Statistics              struct {
							Statistic []struct {
								Recommendation   int `json:"Recommendation"`
								NumberOfAnalysts int `json:"NumberOfAnalysts"`
							} `json:"Statistic"`
						} `json:"Statistics"`
					} `json:"Recommendation"`
					RecommendationSnapshots struct {
						RecommendationSnapshot []struct {
							Mean                    float64 `json:"Mean"`
							High                    int     `json:"High"`
							Low                     int     `json:"Low"`
							NumberOfRecommendations int     `json:"NumberOfRecommendations"`
							Statistics              struct {
								Statistic []struct {
									Recommendation   int `json:"Recommendation"`
									NumberOfAnalysts int `json:"NumberOfAnalysts"`
								} `json:"Statistic"`
							} `json:"Statistics"`
							Age string `json:"Age"`
						} `json:"RecommendationSnapshot"`
					} `json:"RecommendationSnapshots"`
				} `json:"EstimateSummary"`
			} `json:"EstimateSummaries"`
			CalculationBasis string `json:"CalculationBasis"`
		} `json:"GetEstimateSummaries_Response_1"`
	} `json:"stockForecastsResponse"`
	Created []int `json:"created"`
}

// FetchStockForecast fetches stock forecast data from the Mint Genie API
func FetchStockForecast(tickerId, exchangeCode string) (StockForecastResponse, error) {
	url := fmt.Sprintf("https://api-mintgenie.livemint.com/api-gateway/fundamental/v2/getStockFore/%s/%s", tickerId, exchangeCode)

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return StockForecastResponse{}, err
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:135.0) Gecko/20100101 Firefox/135.0")
	req.Header.Set("Accept", "*/*")
	req.Header.Set("Cache-Control", "no-cache")

	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return StockForecastResponse{}, err
	}
	defer resp.Body.Close()

	var forecast StockForecastResponse
	if err := json.NewDecoder(resp.Body).Decode(&forecast); err != nil {
		return StockForecastResponse{}, err
	}
	// print the complete body

	return forecast, nil
}

// GetStockForecastHandler handles API requests for stock forecasts
func GetStockForecastHandler(w http.ResponseWriter, r *http.Request) {
	params := mux.Vars(r)
	tickerId := params["tickerId"]

	liveMindTickerData, err := FetchStockTickerData(tickerId)

	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return

	}

	if len(liveMindTickerData) == 0 {
		http.Error(w, "No data found", http.StatusNotFound)
		return
	}

	livemintTicker := liveMindTickerData[0]

	forecast, err := FetchStockForecast(livemintTicker.ID, "bse")
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	json.NewEncoder(w).Encode(forecast)
}

// func main() {
// 	router := mux.NewRouter()
// 	router.HandleFunc("/stock/forecast/{tickerId}/{exchangeCode}", GetStockForecastHandler).Methods("GET")
//
// 	port := "8000"
// 	fmt.Printf("Server is running on port %s\n", port)
// 	http.ListenAndServe(":"+port, router)
// }
