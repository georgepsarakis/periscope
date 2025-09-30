package notification

import (
	"context"
	"encoding/json"
)

type EventAttributes struct {
	AlertID      string `json:"alert_id"`
	Title        string `json:"title"`
	ProjectID    string `json:"project_id"`
	EventGroupID string `json:"event_group_id"`
}

type Event struct {
	ID         string          `json:"id"`
	Type       string          `json:"type"`
	Data       json.RawMessage `json:"data"`
	Attributes EventAttributes `json:"attributes"`
}

type Channel interface {
	Serialize(event Event) ([]byte, error)
	Emit(ctx context.Context, event Event) error
	// OnSuccess configures a callback that will be called on successful emission
	OnSuccess(func(ctx context.Context, event Event) error)
}
