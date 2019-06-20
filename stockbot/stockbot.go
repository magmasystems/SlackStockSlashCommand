package stockbot

import (
	"bufio"
	"errors"
	"fmt"
	"os"
	"strings"
	"sync"

	config "../configuration"

	av "../alphavantageprovider"
	quandl "../quandlprovider"
	q "../quoteproviders"
	wtd "../worldtradingdata"
)

// QuoteInfo - contains info about a quote
type QuoteInfo struct {
	Symbol    string
	LastPrice float32
}

// BotOps - interface defining all operations the Stockbot can do
type BotOps interface {
	Close()
	QuoteSingle(symbol string) []QuoteInfo
	Quote(symbols []string) []QuoteInfo
	QuoteSingleAsync(symbol string)
	QuoteAsync(symbols []string)
	Config() config.AppSettings
}

// Stockbot - the bot that retrieves stock quotes fro a provider
type Stockbot struct {
	quoteProvider q.QuoteProvider
	QuoteReceived chan []QuoteInfo
}

func main() {
	bot := CreateStockbot()
	fmt.Println(bot.QuoteSingle("MSFT"))

	scanner := bufio.NewScanner(os.Stdin)
	print("Enter the symbol: ")
	for scanner.Scan() {
		symbol := scanner.Text()
		if len(symbol) == 0 {
			break
		}
		price := bot.QuoteSingle(symbol)
		fmt.Println(price)
		print("Enter the symbol: ")
	}
}

// CreateStockbot - creates a new instance of the StockBot
func CreateStockbot() *Stockbot {
	configMgr := new(config.ConfigManager)
	appSettings := configMgr.Config()

	driver := appSettings.Driver
	apiKey := appSettings.APIKeys[driver]

	bot := new(Stockbot)
	bot.quoteProvider, _ = quoteProviderFactory(driver, apiKey)
	bot.QuoteReceived = make(chan []QuoteInfo, 20)

	return bot
}

// Close - disposes of the resources of a stock bot
func (bot *Stockbot) Close() {
	if bot.quoteProvider != nil {
		bot.quoteProvider.Close()
	}
}

// QuoteAsync - gets the price for one or more stocks, and sends a message into the channel when the quotes are ready
func (bot *Stockbot) QuoteAsync(symbols []string) {
	quoteInfo := bot.Quote(symbols)
	bot.QuoteReceived <- quoteInfo
}

// QuoteSingle - gets the price for a stock
func (bot *Stockbot) QuoteSingle(symbol string) []QuoteInfo {
	return bot.Quote([]string{symbol})
}

// QuoteSingleAsync - gets the price for a stock, but uses a channel to inform the caller when the quote is ready
func (bot *Stockbot) QuoteSingleAsync(symbol string) {
	bot.QuoteAsync([]string{symbol})
}

// Quote - gets the price for a stock
func (bot *Stockbot) Quote(symbols []string) []QuoteInfo {
	n := len(symbols)
	quoteInfo := make([]QuoteInfo, n)

	for idx, symbol := range symbols {
		if symbol == "" {
			continue
		}
		price := bot.quoteProvider.FetchQuote(symbol)
		quoteInfo[idx].LastPrice = price
		quoteInfo[idx].Symbol = symbol
	}

	return quoteInfo[0:n]
}

// QuoteWG - gets the price for a stock, using wait groups
func (bot *Stockbot) QuoteWG(symbols []string) []QuoteInfo {
	n := len(symbols)
	quoteInfo := make([]QuoteInfo, n)

	var wg sync.WaitGroup

	for idx, symbol := range symbols {
		if symbol == "" {
			continue
		}
		wg.Add(1)

		go func(b *Stockbot, sym string, qi *QuoteInfo, w *sync.WaitGroup) {
			qi.Symbol = sym
			qi.LastPrice = b.quoteProvider.FetchQuote(sym)
			w.Done()
		}(bot, symbol, &quoteInfo[idx], &wg)
	}

	wg.Wait()

	return quoteInfo[0:n]
}

// quoteProviderFactory - a factory that creates a quote provider
func quoteProviderFactory(providerName string, apiKey string) (provider q.QuoteProvider, errs error) {
	switch strings.ToLower(providerName) {
	case "alphavantage":
		provider = av.CreateQuoteProvider(apiKey)
	case "worldtradingdata":
		provider = wtd.CreateQuoteProvider(apiKey)
	case "quandl":
		provider = quandl.CreateQuoteProvider(apiKey)
	default:
		return nil, errors.New("the Quote Provider cannot be found")
	}

	return provider, nil
}
