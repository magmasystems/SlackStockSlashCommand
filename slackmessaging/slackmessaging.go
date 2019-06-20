// Package slackmessaging - Contains functions that interface with Slack
package slackmessaging

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"

	config "github.com/magmasystems/SlackStockSlashCommand/configuration"
	slack "github.com/nlopes/slack"
)

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
func WriteResponse(writer http.ResponseWriter, outputText string) {
	// Create an output message for Slack and turn it into Json
	outputPayload := &slack.Msg{Text: outputText, ResponseType: "ephemeral"}
	jsonValue, err := json.Marshal(outputPayload)

	// Was there a problem marshalling?
	if err != nil {
		writer.WriteHeader(http.StatusInternalServerError)
		return
	}

	// Send the output back to Slack
	writer.Header().Set("Content-Type", "application/json")
	writer.Write(jsonValue)
}

// PostSlackNotification - posts a message to either a Slack Channel or to a user directly
func PostSlackNotification(slackUserName string, slackChannel string, outputText string, appSettings *config.AppSettings) {
	/*
		outputPayload := &slack.Msg{Text: outputText, User: slackUserName}
		jsonValue, _ := json.Marshal(outputPayload)

		webhook := getWebhook(slackChannel, appSettings)

		_, err := http.Post(webhook, "application/json", bytes.NewBuffer(jsonValue))
		if err != nil {
			fmt.Println(err)
		}
	*/

	// text := fmt.Sprintf("<!channel> %s :smile:\nSee <https://api.slack.com/docs/message-formatting#linking_to_channels_and_users>", outputText)

	attachment := slack.Attachment{
		Color:    "good",
		Fallback: "You successfully posted by Incoming Webhook URL!",
		//AuthorName:    "nlopes/slack",
		//AuthorSubname: "github.com",
		//AuthorLink:    "https://github.com/nlopes/slack",
		//AuthorIcon:    "https://avatars2.githubusercontent.com/u/652790",
		Text: outputText,
		//Footer:        "slack api",
		//FooterIcon:    "https://platform.slack-edge.com/img/default_application_icon.png",
		Ts: json.Number(strconv.FormatInt(time.Now().Unix(), 10)),
	}
	msg := slack.WebhookMessage{
		Attachments: []slack.Attachment{attachment},
		Username:    slackUserName,
		Channel:     slackChannel,
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
