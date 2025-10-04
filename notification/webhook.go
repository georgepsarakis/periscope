package notification

import (
	"time"

	"github.com/georgepsarakis/periscope/repository"
)

type Webhook struct {
	ID        string         `json:"id"`
	Event     string         `json:"event"`
	Timestamp time.Time      `json:"timestamp"`
	Version   string         `json:"version"`
	Data      WebhookPayload `json:"data"`
}

type WebhookPayload struct {
	Alert *WebhookPayloadAlert `json:"alert"`
}

type WebhookPayloadAlert struct {
	repository.Alert
}
