package slackmessaging

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/nlopes/slack"
)

// SlackMessageFormat - Slack-agnostic formatting options
type SlackMessageFormat struct {
	Text    string
	Title   string
	Footer  string
	Color   string
	UseTime bool
}

// ToAttachment - converts a SlackMessageFormat to a Slack Attachment
func (format *SlackMessageFormat) ToAttachment() *slack.Attachment {
	attachment := new(slack.Attachment)
	attachment.Text = format.Text

	if format.Title != "" {
		attachment.Title = format.Title
	}

	if format.Footer != "" {
		attachment.Footer = format.Footer
	}

	if format.Color != "" {
		attachment.Color = format.Color
	}

	if format.UseTime {
		attachment.Ts = json.Number(strconv.FormatInt(time.Now().Unix(), 10))
	}

	/*
		"fallback": "Required plain-text summary of the attachment.",
		"color": "#2eb886",
		"pretext": "Optional text that appears above the attachment block",
		"author_name": "Bobby Tables",
		"author_link": "http://flickr.com/bobby/",
		"author_icon": "http://flickr.com/icons/bobby.jpg",
		"title": "Slack API Documentation",
		"title_link": "https://api.slack.com/",
		"text": "Optional text that appears within the attachment",
		"fields": [
			{
				"title": "Priority",
				"value": "High",
				"short": false
			}
		],
		"image_url": "http://my-website.com/path/to/image.jpg",
		"thumb_url": "http://example.com/path/to/thumb.png",
		"footer": "Slack API",
		"footer_icon": "https://platform.slack-edge.com/img/default_application_icon.png",
		"ts": 123456789
	*/

	return attachment
}

// ToBlock - converts a SlackMessageFormat to a Slack Block
func (format *SlackMessageFormat) ToBlock() slack.Message {
	var blocks []slack.Block

	// Title Section
	if format.Title != "" {
		blocks = append(blocks, createTextBlock("*"+format.Title+"*"))
	}

	// The main text
	blocks = append(blocks, createTextBlock(format.Text))

	// The optional footer
	if format.Footer != "" {
		blocks = append(blocks, createContextBlock(format.Footer))
	}

	msg := slack.NewBlockMessage(blocks...)
	return msg
}

func createTextBlock(text string) *slack.SectionBlock {
	return slack.NewSectionBlock(slack.NewTextBlockObject("mrkdwn", text, false, false), nil, nil)
}

func createContextBlock(text string) *slack.ContextBlock {
	return slack.NewContextBlock("", []slack.MixedElement{
		slack.NewTextBlockObject("mrkdwn", text, false, false),
	}...)
}
