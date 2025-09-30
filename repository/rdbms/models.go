package rdbms

import (
	"database/sql"
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type Event struct {
	gorm.Model
	EventID      string          `json:"event_id" gorm:"not null"`
	Title        string          `json:"title" gorm:"not null"`
	Fingerprint  string          `gorm:"not null"`
	StackTrace   json.RawMessage `gorm:"type:json"`
	EventGroupID uint            `gorm:"not null;index:idx_event_group_id"`
	ProjectID    uint            `gorm:"not null;index:idx_project_id_emitted_at,priority:1"`
	EmittedAt    time.Time       `gorm:"not null;index:idx_project_id_emitted_at,priority:2"`
}

type Project struct {
	gorm.Model
	Name             string `gorm:"not null;index:uq_project_name,unique"`
	PublicID         string `gorm:"not null;index:uq_project_public_id,unique"`
	IngestionAPIKeys []ProjectIngestionAPIKey
}

type ProjectIngestionAPIKey struct {
	gorm.Model
	Key       string     `gorm:"not null"`
	ProjectID uint       `gorm:"not null;index:idx_project_id"`
	ExpiresAt *time.Time `gorm:"null"`
}

type EventGroup struct {
	gorm.Model
	TotalCount       int          `gorm:"not null"`
	EventReceivedAt  time.Time    `gorm:"not null"`
	ProjectID        uint         `gorm:"not null;index:idx_proj_aggr_key,priority:1"`
	AggregationKey   string       `gorm:"not null;index:idx_proj_aggr_key,priority:2"`
	AlertTriggeredAt sql.NullTime `gorm:"null"`
}

type ProjectAlertDestination struct {
	gorm.Model
	ProjectID              uint `gorm:"not null"`
	AlertDestinationTypeID uint `gorm:"not null"`
	AlertDestinationType   AlertDestinationType
	Configuration          json.RawMessage `gorm:"type:json"`
}

type AlertDestinationType struct {
	gorm.Model
	Title string `gorm:"not null"`
	Key   string `gorm:"not null;index:uq_project_alert_destination_type_key,unique"`
}

const (
	AlertDestinationTypeKeyInternalLogger = "internal.logger"
)

type Alert struct {
	gorm.Model
	ProjectID      uint         `gorm:"not null;index:idx_alert_project_id"`
	EventGroupID   uint         `gorm:"not null;index:idx_alert_event_grp_key"`
	TriggeredAt    time.Time    `gorm:"not null;index:idx_alert_triggered_at_key"`
	EscalatedAt    sql.NullTime `gorm:"null"`
	AcknowledgedAt sql.NullTime `gorm:"null"`
	NotifiedAt     sql.NullTime `gorm:"null"`
	Title          string       `gorm:"not null"`
	Description    string       `gorm:"null"`
}

type AlertDestinationNotification struct {
	gorm.Model
	AlertID                   uint       `gorm:"not null;index:idx_alert_destinations_alert_id"`
	ProjectAlertDestinationID uint       `gorm:"not null"`
	CompletedAt               *time.Time `gorm:"null;index:idx_alert_destination_notifications_completed_at"`
	TotalAttempts             int        `gorm:"not null"`
}
