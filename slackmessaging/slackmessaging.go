// Package slackmessaging - Contains functions that interface with Slack
package slackmessaging

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strings"

	config "github.com/magmasystems/SlackStockSlashCommand/configuration"
	"github.com/nlopes/slack"
)

var appSettings *config.AppSettings

func init() {
	// Fetch the appSettings when we load this, because we need to get the Webhooks
	configMgr := new(config.ConfigManager)
	appSettings = configMgr.Config()
}

// ProcessIncomingSlashCommand - reads the incoming request and create a Slash Command
func ProcessIncomingSlashCommand(r *http.Request, w http.ResponseWriter, signingSecret string) (slashCommand slack.SlashCommand, errs error) {
	// Create a SecretsVerifier
	verifier, err := slack.NewSecretsVerifier(r.Header, signingSecret)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Get the command body from the request and parse it into a new Slash Command
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

// WriteResponse - writes text to a ResponseWriter that Slack will receive
func WriteResponse(writer http.ResponseWriter, outputText string) error {
	// Create an output message for Slack and turn it into Json
	outputPayload := &slack.Msg{Text: outputText, ResponseType: "ephemeral"}
	jsonValue, err := json.Marshal(outputPayload)

	// Was there a problem marshalling?
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return err
	}

	// Send the output back to Slack
	writer.Header().Set("Content-Type", "application/json")
	_, err = writer.Write(jsonValue)
	return err
}

// PostSlackNotification - posts a message to either a Slack Channel or to a user directly
func PostSlackNotification(slackUserName string, slackChannel string, outputText string) {
	// text := fmt.Sprintf("<!channel> %s :smile:\nSee <https://api.slack.com/docs/message-formatting#linking_to_channels_and_users>", outputText)

	format := SlackMessageFormat{
		Color:   "good",
		Text:    outputText,
		UseTime: true,
	}

	PostSlackNotificationFormatted(slackUserName, slackChannel, format)
}

// PostSlackNotificationFormatted - posts a message to Slack, but accepts a Slack Attachment as an argument.
// This gives the user the ability to pass in a format that has lots of options.
func PostSlackNotificationFormatted(slackUserName string, slackChannel string, format SlackMessageFormat) {
	msg := slack.WebhookMessage{
		Attachments: []slack.Attachment{*format.ToAttachment()},
	}

	webhook := getWebhook(slackChannel, appSettings)

	err := slack.PostWebhook(webhook, &msg)
	if err != nil {
		fmt.Println(err)
	}
}

func getWebhook(slackChannel string, appSettings *config.AppSettings) string {
	var webhook string

	if strings.Trim(slackChannel, " ") == "" {
		webhook = appSettings.DMWebhook
	} else {
		webhook = appSettings.Webhook
	}

	return webhook
}
