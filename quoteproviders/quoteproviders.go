package quoteproviders

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strings"
	"time"
)

// QuoteProvider - all quote providers must implement this interface
type QuoteProvider interface {
	FetchQuote(symbol string) float32
	SetAPIKey(apiKey string)
	Close()
}

// BaseQuoteProvider - abstract base for all quote providers
type BaseQuoteProvider struct {
	APIKey string
}

// FormattedDates - returns some data strings for today and tomorrow
func (provider BaseQuoteProvider) FormattedDates() (string, string) {
	today := time.Now()
	tomorrow := today.AddDate(0, 0, 1)

	sToday := fmt.Sprintf("%4d-%02d-%02d", today.Year(), today.Month(), today.Day())
	sTomorrow := fmt.Sprintf("%4d-%02d-%02d", tomorrow.Year(), tomorrow.Month(), tomorrow.Day())

	return sToday, sTomorrow
}

// PrepareURL - given the quote provider's Get URL for retrieving a stock quote, and a symbol,
// returns a valid URL which can be used in the REST GET call.
func (provider BaseQuoteProvider) PrepareURL(quoteURL string, symbol string) string {
	sToday, sTomorrow := provider.FormattedDates()

	url := strings.Replace(quoteURL, "{symbol}", symbol, 1)
	url = strings.Replace(url, "{today}", sToday, 2)
	url = strings.Replace(url, "{tomorrow}", sTomorrow, 2)
	url = strings.Replace(url, "{apiKey}", provider.APIKey, 1)

	//fmt.Println(url)

	return url
}

// FetchJSONResponse - calls the REST API fo fetch a quote and returns the JSON payload
func (provider BaseQuoteProvider) FetchJSONResponse(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		log.Fatalln(err)
	}

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Fatalln(err)
	}

	// fmt.Println(string(body))

	return body, nil
}

// SetAPIKey - sets the api key that is used to access the quote service's API
func (provider BaseQuoteProvider) SetAPIKey(apiKey string) {
	provider.APIKey = apiKey
}
