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

	"./stockbot"
	"github.com/nlopes/slack"
)

var theBot *stockbot.Stockbot

func main() {
	logfileName := os.Getenv("LOGFILE")
	if logfileName == "" {
		logfileName = "./logs/stockbot.log"
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

	appSettings := theBot.Config()
	log.Printf("Got the app settings: the port is %d\n", appSettings.Port)

	// Get the signing secret from the config
	var signingSecret string
	signingSecret = appSettings.SlackSecret
	if signingSecret == "" {
		log.Fatal("The signing secret is not in the appSettings.json file")
	}

	// The HTTP request handler
	http.HandleFunc("/quote", func(w http.ResponseWriter, r *http.Request) {
		slashCommand, err := processIncomingRequest(r, w, signingSecret)
		if err != nil {
			return
		}

		// See which slash command the message contains
		switch slashCommand.Command {
		case "/quote":
			getQuotes(slashCommand, w)

		default:
			// Unknown command
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	http.HandleFunc("/quoted", func(w http.ResponseWriter, r *http.Request) {
		slashCommand, err := processIncomingRequest(r, w, appSettings.SlackSecretLocal)
		if err != nil {
			return
		}

		// See which slash command the message contains
		switch slashCommand.Command {
		case "/quoted":
			getQuotes(slashCommand, w)

		default:
			// Unknown command
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	})

	// Get the port from the config file
	port := appSettings.Port
	if port == 0 {
		port = 5000
	}

	// Start the web server
	log.Printf("Listening on port %d\n\n", port)
	http.ListenAndServe(":"+strconv.Itoa(port), nil)
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
		// Create an output message for Slack and turn it into Json
		outputPayload := &slack.Msg{Text: outputText}
		bytes, err := json.Marshal(outputPayload)

		// Was there a problem marshalling?
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		// Send the output back to Slack
		w.Header().Set("Content-Type", "application/json")
		w.Write(bytes)

	case <-time.After(3 * time.Second):
		w.WriteHeader(http.StatusInternalServerError)
	}
}
