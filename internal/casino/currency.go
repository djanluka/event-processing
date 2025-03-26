package casino

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
)

var Currencies = []string{
	"EUR",
	"USD",
	"GBP",
	"NZD",
	"BTC",
}

// Define the smallest unit for each currency
var SmallestUnit = map[string]float64{
	"EUR": 0.01,       // 1 cent
	"USD": 0.01,       // 1 cent
	"GBP": 0.01,       // 1 penny
	"NZD": 0.01,       // 1 cent
	"BTC": 0.00000001, // 1 satoshi
}

type ExchangeRateResponse struct {
	Success bool          `json:"success"`
	Query   QueryResponse `json:"query"`
	Info    InfoResponse  `json:"info"`
	Result  float64       `json:"result"`
}

type QueryResponse struct {
	From string `json:"from"`
	To   string `json:"to"`
}
type InfoResponse struct {
	Timestamp int64   `json:"timestamp"`
	Quote     float64 `json:"quote"`
}

func GetExchangedValueFromApi(from, to string, amount int) *ExchangeRateResponse {
	// Read API endpoint from .env
	apiEndpoint := os.Getenv("EXCHANGE_CONVERT_API_URL")
	if apiEndpoint == "" {
		log.Fatal("Api endpoint is not set in .env file")
	}
	apiEndpoint += fmt.Sprintf("&from=%s&to=%s&amount=%d&format=1", from, to, amount)

	// Make HTTP request to the API and unmarshal the response into a struct
	resp, err := http.Get(apiEndpoint)
	if err != nil {
		log.Fatalf("Error calling API: %v", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatalf("Error reading response body: %v", err)
	}

	var exchangeRateResponse ExchangeRateResponse
	err = json.Unmarshal(body, &exchangeRateResponse)
	if err != nil {
		log.Fatalf("Error unmarshalling JSON: %v", err)
	}

	return &exchangeRateResponse
}
