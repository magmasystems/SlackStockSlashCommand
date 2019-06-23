package alphavantageprovider

import (
	"encoding/json"
	"strconv"

	qp "github.com/magmasystems/SlackStockSlashCommand/quoteproviders"
)

// https://www.alphavantage.co
const quoteURL = "https://www.alphavantage.co/query?function=GLOBAL_QUOTE&symbol={symbol}&apikey={apiKey}"

// AVQuoteProvider - gets quotes from the provider
type AVQuoteProvider struct {
	qp.BaseQuoteProvider
}

// CreateQuoteProvider - creates a new quote provider
func CreateQuoteProvider(apiKey string) qp.QuoteProvider {
	quoteProvider := new(AVQuoteProvider)
	quoteProvider.APIKey = apiKey
	return quoteProvider
}

// Close - closes the provider
func (provider AVQuoteProvider) Close() {
}

// FetchQuote - gets a quote
func (provider AVQuoteProvider) FetchQuote(symbol string) float32 {
	url := provider.PrepareURL(quoteURL, symbol)
	payload, err := provider.FetchJSONResponse(url)

	if err == nil {
		data := new(quoteData)
		err = json.Unmarshal(payload, &data)
		if err != nil {
			return 0
		}
		//fmt.Println(data)

		f, _ := strconv.ParseFloat(data.GlobalQuote.Price, 32)
		return float32(f)
	}

	return 0
}

// QuoteData - contains the data for a symbol in WorldTradingData format
type quoteData struct {
	GlobalQuote struct {
		Symbol           string `json:"01. symbol"`
		Open             string `json:"02. open"`
		High             string `json:"03. high"`
		Low              string `json:"04. low"`
		Price            string `json:"05. price"`
		Volume           string `json:"06. volume"`
		LatestTradingDay string `json:"07. latest trading day"`
		PreviousClose    string `json:"08. previous close"`
		Change           string `json:"09. change"`
		ChangePercent    string `json:"10. change percent"`
	} `json:"Global Quote"`
}
