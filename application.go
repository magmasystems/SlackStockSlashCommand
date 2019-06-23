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

	"github.com/magmasystems/SlackStockSlashCommand/alerts"
	config "github.com/magmasystems/SlackStockSlashCommand/configuration"
	"github.com/magmasystems/SlackStockSlashCommand/framework/logging"
	"github.com/magmasystems/SlackStockSlashCommand/slackmessaging"
	"github.com/magmasystems/SlackStockSlashCommand/stockbot"
	"github.com/nlopes/slack"
)

var (
	theBot                    *stockbot.Stockbot
	theAlertManager           *alerts.AlertManager
	priceBreachCheckingTicker *time.Ticker
	appSettings               *config.AppSettings
)

func main() {
	logfileName := os.Getenv("LOGFILE")
	if logfileName == "" {
		logfileName = "./stockbot.log"
	}

	logging.Infof("Application: Creating the logfile [%s]\n", logfileName)
	f, _ := os.Create(logfileName)
	defer f.Close()
	log.SetOutput(f)

	logging.Infoln("Application: About to create the Stockbot")

	theBot = stockbot.CreateStockbot()
	defer theBot.Close()

	go func() {
		theBot.QuoteSingleAsync("MSFT")
		quoteInfo := <-theBot.QuoteReceived
		fmt.Println(quoteInfo[0])
	}()

	configMgr := new(config.ConfigManager)
	appSettings = configMgr.Config()
	logging.Infof("Application: Got the app settings: the port is %d\n", appSettings.Port)

	// Get the signing secret from the config
	var signingSecret string
	signingSecret = appSettings.SlackSecret
	if signingSecret == "" {
		logging.Fatal("Application: The signing secret is not in the appSettings.json file")
	}

	// Create the AlertManager
	logging.Infoln("Application: About to create the Alert Manager")
	theAlertManager = alerts.CreateAlertManager(theBot)
	defer theAlertManager.Dispose()
	logging.Infoln("Application: Created the Alert Manager")

	// The HTTP request handler
	http.HandleFunc("/quote", func(w http.ResponseWriter, r *http.Request) {
		handleHTTPRequest(w, r, signingSecret)
	})

	//postSlackNotification("UKBM681GV", "This is an unsolicited message from the quote alerter")

	// Create a ticker that will continually check for a price breach
	if !appSettings.DisablePriceBreachChecking {
		logging.Infof("Application: About to create the Price Breach Ticker with interval %d \n", appSettings.QuoteCheckInterval)
		priceBreachCheckingTicker = time.NewTicker(time.Duration(appSettings.QuoteCheckInterval) * time.Second)
		defer priceBreachCheckingTicker.Stop()

		// Every time the ticker elapses, we check for a price breach
		go func() {
			for range priceBreachCheckingTicker.C {
				logging.Infoln("Application: Ticker elapsed: checking price breaches")
				onPriceBreachTickerElapsed()
			}
		}()
	}

	// Get the port from the config file
	port := appSettings.Port
	if port == 0 {
		port = 5000
	}

	// Start the web server
	logging.Infof("Application: Listening on port %d\n\n", port)
	http.ListenAndServe(":"+strconv.Itoa(port), nil)
}

func postSlackNotification(notification alerts.PriceBreachNotification, outputText string) {
	format := slackmessaging.SlackMessageFormat{
		Color:   "good",
		Text:    outputText,
		Title:   "Price Target Reached",
		UseTime: true,
	}

	if notification.Direction == "BELOW" {
		format.Color = "#FF0000"
	}

	slackmessaging.PostSlackNotificationFormatted(notification.SlackUserName, notification.Channel, format)
}

func handleHTTPRequest(w http.ResponseWriter, r *http.Request, signingSecret string) {
	slashCommand, err := slackmessaging.ProcessIncomingSlashCommand(r, w, signingSecret)
	logging.Infof("The slash command is: [%s]\n", slashCommand)
	if err != nil {
		return
	}

	// See which slash command the message contains
	switch slashCommand.Command {
	case "/quote", "/quoted":
		switch strings.ToUpper(slashCommand.Text) {
		case "HELP":
			helpQuoteCommand(w)
		default:
			getQuotes(slashCommand, w)
		}

	case "/quote-alert", "/quoted-alert":
		switch strings.ToUpper(slashCommand.Text) {
		case "HELP":
			helpQuoteAlertCommand(w)
		case "CHECK":
			checkForPriceBreaches(w)
		default:
			theAlertManager.HandleQuoteAlert(slashCommand, w)
		}

	default:
		// Unknown command
		w.WriteHeader(http.StatusInternalServerError)
	}
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
		logging.Infoln("The notification to Slack is:")
		logging.Infoln(fmt.Sprint(notification))
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
		outputText = fmt.Sprintf("%s has gone %s the target price of %3.2f. The current price is %3.2f.\n",
			notification.Symbol, notification.Direction, notification.TargetPrice, notification.CurrentPrice)

		// Do the notification to slack synchronously
		postSlackNotification(notification, outputText)

		go func() {
			println(outputText)
		}()
	}
}

func helpQuoteCommand(w http.ResponseWriter) {
	text := `/quote symbol[,symbol,symbol,...]
	Retrieves the current price one or more stocks. Each stock can be separated by a comma`
	slackmessaging.WriteResponse(w, text)
}

func helpQuoteAlertCommand(w http.ResponseWriter) {
	text := `/quote-alert [symbol price [below]] [symbol delete] [deleteall] [#channel]
	Sets up a subscription to a price alert for the specified symbol. 
	Examples:
	  /quote-alert - lists all of the alerts you have
	  /quote-alert MSFT 130 - sends an alert when Microsoft stock reaches $130
	  /quote-alert MSFT 130 #myalerts - sends an alert  to the #myalert Slack channel when Microsoft stock reaches $130
	  /quote-alert MSFT 130 BELOW - sends an alert when Microsoft stock goes below $130
	  /quote-alert MSFT delete - removes the existing alert on MSFT stock that you have subscribed to
	  /quote-alert deleteall - deletes all alerts that you have
	`
	slackmessaging.WriteResponse(w, text)
}
