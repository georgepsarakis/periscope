package repository

import (
	"database/sql"
	"encoding/json"
	"time"
)

type BaseModel struct {
	ID        uint      `json:"id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type Event struct {
	BaseModel
	EventID      string          `json:"event_id"`
	Title        string          `json:"title"`
	Fingerprint  string          `json:"fingerprint"`
	StackTrace   json.RawMessage `json:"stack_trace"`
	EventGroupID uint            `json:"event_group_id"`
	ProjectID    uint            `json:"project_id"`
	EmittedAt    time.Time       `json:"emitted_at"`
}

type Project struct {
	BaseModel
	Name                    string                   `json:"name"`
	PublicID                string                   `json:"public_id"`
	ProjectIngestionAPIKeys []ProjectIngestionAPIKey `json:"project_ingestion_api_keys"`
}

func (p Project) HasAccess(key string) bool {
	if len(p.ProjectIngestionAPIKeys) == 0 {
		return false
	}
	for _, apiKey := range p.ProjectIngestionAPIKeys {
		if apiKey.ExpiresAt != nil && apiKey.ExpiresAt.Before(time.Now()) {
			continue
		}
		if key == apiKey.Key {
			return true
		}
	}
	return false
}

type ProjectIngestionAPIKey struct {
	BaseModel
	Key       string     `json:"key"`
	ProjectID uint       `json:"project_id"`
	ExpiresAt *time.Time `json:"expires_at"`
}

type EventGroup struct {
	BaseModel
	TotalCount      int       `json:"total_count"`
	EventReceivedAt time.Time `json:"event_received_at"`
	ProjectID       uint      `json:"project_id"`
	AggregationKey  string    `json:"aggregation_key"`
}

type Alert struct {
	BaseModel
	ProjectID      uint         `json:"project_id"`
	EventGroupID   uint         `json:"event_group_id"`
	TriggeredAt    time.Time    `json:"triggered_at"`
	EscalatedAt    sql.NullTime `json:"escalated_at"`
	AcknowledgedAt sql.NullTime `json:"acknowledged_at"`
	NotifiedAt     sql.NullTime `json:"notified_at"`
}

type AlertDestinationNotification struct {
	BaseModel
	AlertID                   uint       `json:"alert_id"`
	ProjectAlertDestinationID uint       `json:"project_alert_destination_id"`
	CompletedAt               *time.Time `json:"completed_at"`
	TotalAttempts             int        `json:"total_attempts"`
}

type ProjectAlertDestination struct {
	BaseModel
	ProjectID              uint            `json:"project_id"`
	AlertDestinationTypeID uint            `json:"alert_destination_type_id"`
	Configuration          json.RawMessage `json:"configuration"`
}

type AlertDestinationType struct {
	BaseModel
	Title string `json:"title"`
	Key   string `json:"key"`
}
