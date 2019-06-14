package main

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	_ "github.com/lib/pq"

	alerts "./alerts"
	config "./configuration"
	"./stockbot"
	"github.com/nlopes/slack"
)

var theBot *stockbot.Stockbot
var theAlertManager *alerts.AlertManager
var priceBreachCheckingTicker *time.Ticker

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
	appSettings := configMgr.Config()
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
		slashCommand, err := processIncomingRequest(r, w, signingSecret)
		if err != nil {
			return
		}

		// See which slash command the message contains
		switch slashCommand.Command {
		case "/quote", "/quoted":
			getQuotes(slashCommand, w)

		case "/quote-alert":
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

	priceBreachCheckingTicker = time.NewTicker(time.Duration(appSettings.QuoteCheckInterval) * time.Minute)
	defer priceBreachCheckingTicker.Stop()
	go func() {
		for range priceBreachCheckingTicker.C {
			fmt.Println("Checking for price breaches at " + time.Now().String())
			theAlertManager.CheckForPriceBreaches(theBot, func(notification alerts.PriceBreachNotification) {
				fmt.Println(notification)
			})
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

func respondToSlack(text string, w http.ResponseWriter) {
	outputPayload := &slack.Msg{Text: text}
	bytes, err := json.Marshal(outputPayload)

	// Was there a problem marshalling?
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Send the output back to Slack
	w.Header().Set("Content-Type", "application/json")
	w.Write(bytes)
}

func processIncomingRequest(r *http.Request, w http.ResponseWriter, signingSecret string) (slashCommand slack.SlashCommand, errs error) {
	log.Println("Got a /quote request")
	verifier, err := slack.NewSecretsVerifier(r.Header, signingSecret)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	r.Body = ioutil.NopCloser(io.TeeReader(r.Body, &verifier))
	slashCommand, err = slack.SlashCommandParse(r)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return slashCommand, err
	}
	log.Printf("The slash command is %s and the text is %s\n", slashCommand.Command, slashCommand.Text)

	// Verify that the request came from Slack
	if err = verifier.Ensure(); err != nil {
		w.WriteHeader(http.StatusUnauthorized)
		return slashCommand, err
	}

	return slashCommand, nil
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

		respondToSlack(outputText, w)

	case <-time.After(3 * time.Second):
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func checkForPriceBreaches(w http.ResponseWriter) {
	// Get the latest quotes
	prices := theAlertManager.GetQuotesForAlerts(theBot)
	if prices == nil {
		respondToSlack("No price breaches", w)
		return
	}

	// Save the prices to the database
	theAlertManager.SavePrices(prices)

	// Check for any price breaches
	notifications := theAlertManager.GetPriceBreaches()

	// Go through all of the price breaches and notify the Slack user
	outputText := ""
	for _, notification := range notifications {
		//if notification.WebHook == "" {
		//	continue
		//}

		outputText += fmt.Sprintf("%s has gone %s the target price of %3.2f. The current price is %3.2f.\n",
			notification.Symbol, notification.Direction, notification.TargetPrice, notification.CurrentPrice)

		// Do the notification to slack asynchronously
		go func() {
			println(outputText)
		}()
	}

	respondToSlack(outputText, w)
}
