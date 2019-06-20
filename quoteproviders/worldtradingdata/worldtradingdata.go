package worldtradingdata

import (
	"encoding/json"
	"fmt"
	"strconv"

	qp ".."
)

// https://www.worldtradingdata.com/home

const quoteURL = "https://api.worldtradingdata.com/api/v1/stock?symbol={symbol}&api_token={apiKey}"

// WTDQuoteProvider - gets quotes from worldtradingdata.com
type WTDQuoteProvider struct {
	qp.BaseQuoteProvider
}

// CreateQuoteProvider - creates a new quote provider
func CreateQuoteProvider(apiKey string) qp.QuoteProvider {
	quoteProvider := new(WTDQuoteProvider)
	quoteProvider.APIKey = apiKey
	return quoteProvider
}

// Close - closes the provider
func (provider WTDQuoteProvider) Close() {
}

// FetchQuote - gets a quote
func (provider WTDQuoteProvider) FetchQuote(symbol string) float32 {

	url := provider.PrepareURL(quoteURL, symbol)
	payload, err := provider.FetchJSONResponse(url)

	if err == nil {
		data := new(quoteData)
		json.Unmarshal(payload, &data)
		fmt.Println(data)

		f, _ := strconv.ParseFloat(data.Data[0].Price, 32)
		return float32(f)
	}

	return 0
}

// QuoteData - contains the data for a symbol in WorldTradingData format
type quoteData struct {
	SymbolsRequested int `json:"symbols_requested"`
	SymbolsReturned  int `json:"symbols_returned"`
	Data             []struct {
		Symbol             string `json:"symbol"`
		Name               string `json:"name"`
		Price              string `json:"price"`
		CloseYesterday     string `json:"close_yesterday"`
		ReturnYtd          string `json:"return_ytd"`
		NetAssets          string `json:"net_assets"`
		ChangeAssetValue   string `json:"change_asset_value"`
		ChangePct          string `json:"change_pct"`
		YieldPct           string `json:"yield_pct"`
		ReturnDay          string `json:"return_day"`
		Return1Week        string `json:"return_1week"`
		Return4Week        string `json:"return_4week"`
		Return13Week       string `json:"return_13week"`
		Return52Week       string `json:"return_52week"`
		Return156Week      string `json:"return_156week"`
		Return260Week      string `json:"return_260week"`
		IncomeDividend     string `json:"income_dividend"`
		IncomeDividendDate string `json:"income_dividend_date"`
		CapitalGain        string `json:"capital_gain"`
		ExpenseRatio       string `json:"expense_ratio"`
	} `json:"data"`
}
