package http

import (
	"time"

	"github.com/georgepsarakis/periscope/ingestion"
)

type Event ingestion.SDKEvent

type ProjectEventMessage struct {
	ProjectID uint  `json:"project_id"`
	Event     Event `json:"event"`
}

type Project struct {
	ID               uint      `json:"id"`
	CreatedAt        time.Time `json:"created_at"`
	UpdatedAt        time.Time `json:"updated_at"`
	Name             string    `json:"name"`
	IngestionAPIKeys []string  `json:"ingestion_api_keys"`
	PublicID         string    `json:"public_id"`
}
