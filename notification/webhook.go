package notification

import (
	"encoding/json"
	"time"
)

type Webhook struct {
	ID        string          `json:"id"`
	Event     string          `json:"event"`
	Timestamp time.Time       `json:"timestamp"`
	Data      json.RawMessage `json:"data"`
	Version   string          `json:"version"`
}
