package quandlprovider

import (
	"encoding/json"
	"fmt"
	"strconv"
	"time"

	qp "github.com/magmasystems/SlackStockSlashCommand/quoteproviders"
)

const quoteURL = "https://www.quandl.com/api/v3/datasets/WIKI/{symbol}.json?start_date={today}&end_date={tomorrow}&api_key={apiKey}"

// QuandlQuoteProvider - gets quotes from worldtradingdata.com
type QuandlQuoteProvider struct {
	qp.BaseQuoteProvider
}

// CreateQuoteProvider - creates a new quote provider
func CreateQuoteProvider(apiKey string) qp.QuoteProvider {
	quoteProvider := new(QuandlQuoteProvider)
	quoteProvider.APIKey = apiKey
	return quoteProvider
}

// Close - closes the provider
func (provider QuandlQuoteProvider) Close() {
}

// FetchQuote - gets a quote
func (provider QuandlQuoteProvider) FetchQuote(symbol string) float32 {
	url := provider.PrepareURL(quoteURL, symbol)
	payload, err := provider.FetchJSONResponse(url)

	if err == nil {
		data := new(quoteData)
		json.Unmarshal(payload, &data)
		fmt.Println(data)

		f, _ := strconv.ParseFloat("156.0", 32)
		return float32(f)
	}

	return 0
}

type quoteData struct {
	Dataset struct {
		ID                  int           `json:"id"`
		DatasetCode         string        `json:"dataset_code"`
		DatabaseCode        string        `json:"database_code"`
		Name                string        `json:"name"`
		Description         string        `json:"description"`
		RefreshedAt         time.Time     `json:"refreshed_at"`
		NewestAvailableDate string        `json:"newest_available_date"`
		OldestAvailableDate string        `json:"oldest_available_date"`
		ColumnNames         []string      `json:"column_names"`
		Frequency           string        `json:"frequency"`
		Type                string        `json:"type"`
		Premium             bool          `json:"premium"`
		Limit               interface{}   `json:"limit"`
		Transform           interface{}   `json:"transform"`
		ColumnIndex         interface{}   `json:"column_index"`
		StartDate           string        `json:"start_date"`
		EndDate             string        `json:"end_date"`
		Data                []interface{} `json:"data"`
		Collapse            interface{}   `json:"collapse"`
		Order               interface{}   `json:"order"`
		DatabaseID          int           `json:"database_id"`
	} `json:"dataset"`
}
