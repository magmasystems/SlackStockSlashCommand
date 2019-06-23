package quandlprovider

import (
	"encoding/json"
	"fmt"
	"math"
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
		err = json.Unmarshal(payload, &data)
		if err != nil {
			return 0
		}
		fmt.Println(data)

		value := data.Dataset.Data[0][4]

		switch i := value.(type) {
		case float64:
			return float32(i)
		case float32:
			return float32(i)
		case int64:
			return float32(i)
		default:
			return float32(math.NaN())
		}
	}

	return 0
}

type quoteData struct {
	Dataset struct {
		ID                  int             `json:"id"`
		DatasetCode         string          `json:"dataset_code"`
		DatabaseCode        string          `json:"database_code"`
		Name                string          `json:"name"`
		Description         string          `json:"description"`
		RefreshedAt         time.Time       `json:"refreshed_at"`
		NewestAvailableDate string          `json:"newest_available_date"`
		OldestAvailableDate string          `json:"oldest_available_date"`
		ColumnNames         []string        `json:"column_names"`
		Frequency           string          `json:"frequency"`
		Type                string          `json:"type"`
		Premium             bool            `json:"premium"`
		Limit               interface{}     `json:"limit"`
		Transform           interface{}     `json:"transform"`
		ColumnIndex         interface{}     `json:"column_index"`
		StartDate           string          `json:"start_date"`
		EndDate             string          `json:"end_date"`
		Data                [][]interface{} `json:"data"`
		Collapse            interface{}     `json:"collapse"`
		Order               interface{}     `json:"order"`
		DatabaseID          int             `json:"database_id"`
	} `json:"dataset"`
}

/*

  "dataset": {
    "id": 9775827,
    "dataset_code": "MSFT",
    "database_code": "WIKI",
    "name": "Microsoft Corporation (MSFT) Prices, Dividends, Splits and Trading Volume",
    "description": "End of day open, high, low, close and volume, dividends and splits, and split/dividend adjusted open, high, low close and volume for Microsoft Corporation (MSFT). Ex-Dividend is non-zero on ex-dividend dates. Split Ratio is 1 on non-split dates. Adjusted prices are calculated per CRSP (www.crsp.com/products/documentation/crsp-calculations)\n\nThis data is in the public domain. You may copy, distribute, disseminate or include the data in other products for commercial and/or noncommercial purposes.\n\nThis data is part of Quandl's Wiki initiative to get financial data permanently into the public domain. Quandl relies on users like you to flag errors and provide data where data is wrong or missing. Get involved: connect@quandl.com\n",
    "refreshed_at": "2018-03-27T21:46:11.788Z",
    "newest_available_date": "2018-03-27",
    "oldest_available_date": "1986-03-13",
    "column_names": [
      "Date",
      "Open",
      "High",
      "Low",
      "Close",
      "Volume",
      "Ex-Dividend",
      "Split Ratio",
      "Adj. Open",
      "Adj. High",
      "Adj. Low",
      "Adj. Close",
      "Adj. Volume"
    ],
    "frequency": "daily",
    "type": "Time Series",
    "premium": false,
    "limit": null,
    "transform": null,
    "column_index": null,
    "start_date": "2017-06-20",
    "end_date": "2017-06-20",
    "data": [
      [
        "2017-06-20",
        70.82,
        70.87,
        69.87,
        69.91,
        20775590.0,
        0.0,
        1.0,
        70.090024064216,
        70.139508690073,
        69.149816172928,
        69.189403873614,
        20775590.0
      ]
    ],
    "collapse": null,
    "order": null,
    "database_id": 4922
  }
}
*/
