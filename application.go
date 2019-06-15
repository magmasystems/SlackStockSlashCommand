package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"

	alerts "./alerts"
	config "./configuration"
	slackmessaging "./slackmessaging"
	"./stockbot"
	"github.com/nlopes/slack"
)

var theBot *stockbot.Stockbot
var theAlertManager *alerts.AlertManager
var priceBreachCheckingTicker *time.Ticker
var appSettings *config.AppSettings

func main() {
	logfileName := os.Getenv("LOGFILE")
	if logfileName == "" {
		logfileName = "./stockbot.log"
	}

	f, _ := os.Create(logfileName)
	defer f.Close()
	log.SetOutput(f)

	log.Println("About to create the Stockbot")

	theBot = stockbot.CreateStockbot()
	defer theBot.Close()

	go func() {
		theBot.QuoteSingleAsync("MSFT")
		quoteInfo := <-theBot.QuoteReceived
		fmt.Println(quoteInfo[0])
	}()

	configMgr := new(config.ConfigManager)
	appSettings = configMgr.Config()
	log.Printf("Got the app settings: the port is %d\n", appSettings.Port)

	// Get the signing secret from the config
	var signingSecret string
	signingSecret = appSettings.SlackSecret
	if signingSecret == "" {
		log.Fatal("The signing secret is not in the appSettings.json file")
	}

	// Create the AlertManager
	theAlertManager = alerts.CreateAlertManager()
	defer theAlertManager.Dispose()

	// The HTTP request handler
	http.HandleFunc("/quote", func(w http.ResponseWriter, r *http.Request) {
		slashCommand, err := slackmessaging.ProcessIncomingSlashCommand(r, w, signingSecret)
		fmt.Println("The slash command is:")
		fmt.Println(slashCommand)
		if err != nil {
			return
		}

		// See which slash command the message contains
		switch slashCommand.Command {
		case "/quote", "/quoted":
			getQuotes(slashCommand, w)

		case "/quote-alert", "/quoted-alert":
			if strings.ToUpper(slashCommand.Text) == "CHECK" {
				checkForPriceBreaches(w)
			} else {
				theAlertManager.HandleQuoteAlert(slashCommand, w)
			}

		default:
			// Unknown command
			w.WriteHeader(http.StatusInternalServerError)
		}
	})

	//postSlackNotification("UKBM681GV", "This is an unsolicited message from the quote alerter")

	// Create a ticker that will continually check for a price breach
	priceBreachCheckingTicker = time.NewTicker(time.Duration(appSettings.QuoteCheckInterval) * time.Second)
	defer priceBreachCheckingTicker.Stop()

	// Every time the ticker elapses, we check for a price breach
	go func() {
		for range priceBreachCheckingTicker.C {
			onPriceBreachTickerElapsed()
		}
	}()

	// Get the port from the config file
	port := appSettings.Port
	if port == 0 {
		port = 5000
	}

	// Start the web server
	log.Printf("Listening on port %d\n\n", port)
	http.ListenAndServe(":"+strconv.Itoa(port), nil)
}

func postSlackNotification(notification alerts.PriceBreachNotification, outputText string) {
	slackmessaging.PostSlackNotification(notification.SlackUserName, notification.Channel, outputText, appSettings)
}

func getQuotes(slashCommand slack.SlashCommand, w http.ResponseWriter) {
	outputText := ""

	symbols := strings.Split(slashCommand.Text, ",")
	go func() {
		theBot.QuoteAsync(symbols)
	}()

	select {
	case quotes := <-theBot.QuoteReceived:
		for _, q := range quotes {
			outputText += fmt.Sprintf("%s: %3.2f\n", strings.ToUpper(q.Symbol), q.LastPrice)
		}

		slackmessaging.WriteResponse(w, outputText)

	case <-time.After(3 * time.Second):
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// onPriceBreachTickerElapsed - This gets called every time the Price Breach Ticker ticks
func onPriceBreachTickerElapsed() {
	fmt.Println("Checking for price breaches at " + time.Now().String())

	theAlertManager.CheckForPriceBreaches(theBot, func(notification alerts.PriceBreachNotification) {
		fmt.Println("The notification to Slack is:")
		fmt.Println(notification)
		outputText := fmt.Sprintf("%s has gone %s the target price of %3.2f. The current price is %3.2f.\n",
			notification.Symbol, notification.Direction, notification.TargetPrice, notification.CurrentPrice)
		postSlackNotification(notification, outputText)
	})
}

// checkForPriceBreaches - this is called when we get a /quote-alert CHECK
func checkForPriceBreaches(w http.ResponseWriter) {
	// Get the latest quotes
	prices := theAlertManager.GetQuotesForAlerts(theBot)
	if prices == nil {
		slackmessaging.WriteResponse(w, "No price breaches")
		return
	}

	// Save the prices to the database
	theAlertManager.SavePrices(prices)

	// Check for any price breaches
	notifications := theAlertManager.GetPriceBreaches()

	// Go through all of the price breaches and notify the Slack user
	outputText := ""
	for _, notification := range notifications {
		outputText += fmt.Sprintf("%s has gone %s the target price of %3.2f. The current price is %3.2f.\n",
			notification.Symbol, notification.Direction, notification.TargetPrice, notification.CurrentPrice)

		// Do the notification to slack asynchronously
		go func() {
			println(outputText)
		}()
	}

	slackmessaging.WriteResponse(w, outputText)
}
