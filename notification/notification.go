package notification

import (
	"context"
	"time"

	"github.com/georgepsarakis/periscope/repository"
)

type Event struct {
	ID      string           `json:"id"`
	Type    string           `json:"type"`
	Alert   repository.Alert `json:"alert"`
	Details EventDetails     `json:"details"`
}

type EventDetails struct {
	AlertID      string `json:"alert_id"`
	Title        string `json:"title"`
	ProjectID    string `json:"project_id"`
	EventGroupID string `json:"event_group_id"`
}

type Channel interface {
	Serialize(event Event) ([]byte, error)
	Emit(ctx context.Context, event Event) error
}

func DefaultClock() time.Time {
	return time.Now().UTC()
}
