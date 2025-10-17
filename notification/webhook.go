package notification

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/georgepsarakis/go-httpclient"
	"github.com/slack-go/slack"
)

type WebhookEvent struct {
	ID        string          `json:"id"`
	Event     string          `json:"event"`
	Timestamp time.Time       `json:"timestamp"`
	Version   string          `json:"version"`
	Data      json.RawMessage `json:"data"`
}

type GenericWebhookNotification struct {
	Channel
	httpClient  *httpclient.Client
	clock       func() time.Time
	httpHeaders map[string]string
	webhookURL  string
}

type GenericWebhookNotificationSettings struct {
	WebhookURL  string
	HTTPClient  *httpclient.Client
	HTTPHeaders map[string]string
	Clock       func() time.Time
}

func NewGenericWebhookNotification(s GenericWebhookNotificationSettings) GenericWebhookNotification {
	if s.Clock == nil {
		s.Clock = DefaultClock
	}
	return GenericWebhookNotification{
		httpClient:  s.HTTPClient,
		clock:       s.Clock,
		webhookURL:  s.WebhookURL,
		httpHeaders: s.HTTPHeaders,
	}
}

func (w GenericWebhookNotification) Serialize(event Event) ([]byte, error) {
	return json.Marshal(event)
}

const WebhookVersion = "1.0"
const WebhookUserAgent = "periscope/" + WebhookVersion

func (w GenericWebhookNotification) Emit(ctx context.Context, event Event) error {
	s, err := w.Serialize(event)
	if err != nil {
		return err
	}
	ev := WebhookEvent{
		Event:     "alert",
		ID:        event.ID,
		Timestamp: w.clock(),
		Data:      s,
		Version:   WebhookVersion,
	}
	s, err = json.Marshal(ev)
	if err != nil {
		return err
	}
	resp, err := w.httpClient.Post(ctx, w.webhookURL, bytes.NewReader(s),
		httpclient.WithHeaders(map[string]string{
			"user-agent": WebhookUserAgent,
		}),
		httpclient.WithHeaders(w.httpHeaders))
	if err != nil {
		return err
	}
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("webhook returned non-200 status code: %d", resp.StatusCode)
	}
	return nil
}

type SlackWebhookNotification struct {
	Channel
	httpClient *http.Client
	webhookURL string
}

func NewSlackWebhookNotification(httpClient *http.Client, webhookURL string) SlackWebhookNotification {
	return SlackWebhookNotification{
		httpClient: httpClient,
		webhookURL: webhookURL,
	}
}

func (w SlackWebhookNotification) Serialize(event Event) ([]byte, error) {
	return json.MarshalIndent(event, "", "  ")
}

func (w SlackWebhookNotification) Emit(ctx context.Context, event Event) error {
	b, err := w.Serialize(event)
	if err != nil {
		return err
	}
	return slack.PostWebhookCustomHTTPContext(ctx, w.webhookURL, w.httpClient, &slack.WebhookMessage{
		Text: fmt.Sprintf("[Periscope Alert] %s", event.Details.Title),
		Blocks: &slack.Blocks{
			BlockSet: []slack.Block{
				slack.NewMarkdownBlock("periscope-event-details", "```"+string(b)+"```"),
			},
		},
	})
}
