package repository

import (
	"context"
	"errors"
	"fmt"
	"time"

	"gorm.io/gorm"

	"github.com/georgepsarakis/periscope/newcontext"
	"github.com/georgepsarakis/periscope/repository/rdbms"
)

var ErrRecordNotFound = errors.New("record not found")

func (r *Repository) AlertFindByNotNotified(ctx context.Context) (Alert, error) {
	alert := rdbms.Alert{}
	if tx := r.database.WithContext(ctx).
		Where("notified_at IS NULL").
		First(&alert); tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return Alert{}, ErrRecordNotFound
		}
		return Alert{}, tx.Error
	}
	return Alert{
		BaseModel: BaseModel{
			ID:        alert.ID,
			CreatedAt: alert.CreatedAt,
			UpdatedAt: alert.UpdatedAt,
		},
		ProjectID:    alert.ProjectID,
		EventGroupID: alert.EventGroupID,
		TriggeredAt:  alert.TriggeredAt,
		NotifiedAt:   alert.NotifiedAt,
		EscalatedAt:  alert.EscalatedAt,
	}, nil
}

type ListFilters struct {
	Status string
}

func (r *Repository) FindAlerts(ctx context.Context, projectID uint, options ListFilters) ([]Alert, error) {
	tx := r.dbExecutor(ctx)
	var alertList []rdbms.Alert
	res := tx.Model(&rdbms.Alert{}).Find(&alertList, "project_id = ?", projectID).
		Order("updated_at DESC, created_at DESC")
	if res.Error != nil {
		return nil, res.Error
	}
	if len(alertList) == 0 {
		return nil, nil
	}
	ra := make([]Alert, 0, len(alertList))
	for _, alert := range alertList {
		ra = append(ra, Alert{
			BaseModel: BaseModel{
				ID:        alert.ID,
				CreatedAt: alert.CreatedAt,
				UpdatedAt: alert.UpdatedAt,
			},
			ProjectID:    alert.ProjectID,
			EventGroupID: alert.EventGroupID,
			TriggeredAt:  alert.TriggeredAt,
			NotifiedAt:   alert.NotifiedAt,
			EscalatedAt:  alert.EscalatedAt,
		})
	}
	return ra, nil
}

func (r *Repository) AlertUpdateNotifiedAt(ctx context.Context, id uint, ts time.Time) (Alert, error) {
	tx := r.dbExecutor(ctx)
	res := tx.Model(&rdbms.Alert{}).Where("id = ?", id).Updates(map[string]any{"notified_at": ts})
	if res.Error != nil {
		return Alert{}, res.Error
	}
	return Alert{}, nil
}

func (r *Repository) dbExecutor(ctx context.Context) *gorm.DB {
	tx := newcontext.DBTransactionFromContext(ctx)
	if tx == nil {
		return r.database.WithContext(ctx)
	}
	return tx
}

func (r *Repository) CreateAlertDestinationNotification(ctx context.Context, alertID uint, projectAlertDestinationID uint) (AlertDestinationNotification, error) {
	tx := r.dbExecutor(ctx)
	n := rdbms.AlertDestinationNotification{
		AlertID:                   alertID,
		ProjectAlertDestinationID: projectAlertDestinationID,
	}
	tx = tx.Create(&n)
	if tx.Error != nil {
		return AlertDestinationNotification{}, tx.Error
	}
	return AlertDestinationNotification{
		BaseModel: BaseModel{
			ID:        n.ID,
			CreatedAt: n.CreatedAt,
			UpdatedAt: n.UpdatedAt,
		},
		AlertID:                   alertID,
		ProjectAlertDestinationID: projectAlertDestinationID,
	}, nil
}

const maxAttempts = 10

func (r *Repository) FindAlertDestinationNotificationByNonCompleted(ctx context.Context) (AlertDestinationNotification, error) {
	db := r.dbExecutor(ctx)
	n := AlertDestinationNotification{}
	err := db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		res := tx.
			Where("completed_at IS NULL").
			Where("total_attempts < ?", maxAttempts).
			Order("updated_at").Take(&n)
		if res.Error != nil {
			if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
				return nil
			}
		}
		if n.ID == 0 {
			return ErrRecordNotFound
		}
		res = tx.Model(&n).Updates(map[string]any{
			"total_attempts": gorm.Expr("total_attempts + ?", 1),
			"attempted_at":   r.now(),
		})
		if res.Error != nil {
			return res.Error
		}
		return nil
	})
	if err != nil {
		return AlertDestinationNotification{}, err
	}
	return n, nil
}

func (r *Repository) FindAlertDestinationByID(ctx context.Context, id uint) (ProjectAlertDestination, error) {
	db := r.dbExecutor(ctx)
	pad := rdbms.ProjectAlertDestination{
		Model: gorm.Model{
			ID: id,
		},
	}
	tx := db.Model(&pad).Joins("AlertDestinationType").Find(&pad)
	if tx.Error != nil {
		return ProjectAlertDestination{}, tx.Error
	}

	webhookCfg := &rdbms.AlertDestinationNotificationWebhookConfiguration{}
	tx = db.Model(&rdbms.AlertDestinationNotificationWebhookConfiguration{}).
		Where("project_alert_destination_id = ?", pad.ID).Find(webhookCfg)
	if tx.Error != nil {
		return ProjectAlertDestination{}, tx.Error
	}

	return ProjectAlertDestination{
		BaseModel: BaseModel{
			ID:        pad.ID,
			CreatedAt: pad.CreatedAt,
			UpdatedAt: pad.UpdatedAt,
		},
		ProjectID:              pad.ProjectID,
		AlertDestinationTypeID: pad.AlertDestinationTypeID,
		WebhookConfiguration: &AlertDestinationNotificationWebhookConfiguration{
			BaseModel: BaseModel{
				ID:        webhookCfg.ID,
				CreatedAt: webhookCfg.CreatedAt,
				UpdatedAt: webhookCfg.UpdatedAt,
			},
			ProjectAlertDestinationID: webhookCfg.ProjectAlertDestinationID,
			URL:                       webhookCfg.URL,
			Headers:                   webhookCfg.Headers,
		},
	}, nil
}

func (r *Repository) FindAlertDestinationsByProjectID(ctx context.Context, id uint) ([]ProjectAlertDestination, error) {
	db := r.dbExecutor(ctx)
	var ad []rdbms.ProjectAlertDestination
	tx := db.Model(&rdbms.ProjectAlertDestination{}).
		Where("project_id = ?", id).Find(&ad)
	if tx.Error != nil {
		return []ProjectAlertDestination{}, tx.Error
	}

	var result []ProjectAlertDestination
	for _, a := range ad {
		result = append(result, ProjectAlertDestination{
			BaseModel: BaseModel{
				ID:        a.ID,
				CreatedAt: a.CreatedAt,
				UpdatedAt: a.UpdatedAt,
			},
			ProjectID:              a.ProjectID,
			AlertDestinationTypeID: a.AlertDestinationTypeID,
		})
	}
	return result, nil
}

func (r *Repository) AlertDestinationNotificationUpdateCompletedAt(ctx context.Context, id uint, ts time.Time) (AlertDestinationNotification, error) {
	tx := r.dbExecutor(ctx)
	res := tx.Model(&rdbms.AlertDestinationNotification{}).Where("id = ?", id).Updates(map[string]any{"completed_at": ts})
	if res.Error != nil {
		return AlertDestinationNotification{}, res.Error
	}
	return AlertDestinationNotification{}, nil
}

func (r *Repository) AlertDestinationNotificationUpdateFailure(ctx context.Context, id uint, ts time.Time) (AlertDestinationNotification, error) {
	tx := r.dbExecutor(ctx)
	res := tx.Model(&rdbms.AlertDestinationNotification{}).Where("id = ?", id).Updates(map[string]any{"completed_at": ts})
	if res.Error != nil {
		return AlertDestinationNotification{}, res.Error
	}
	return AlertDestinationNotification{}, nil
}

func (r *Repository) CreateProjectAlertDestination(ctx context.Context, projectID uint, typeAlias string, webhookCfg *AlertDestinationNotificationWebhookConfiguration) (ProjectAlertDestination, error) {
	tx := r.dbExecutor(ctx)
	var searchKey string
	switch typeAlias {
	case "generic_webhook":
		searchKey = rdbms.AlertDestinationTypeKeyGenericWebhook
	case "slack_webhook":
		searchKey = rdbms.AlertDestinationTypeKeySlackWebhook
	case "internal_logger":
		searchKey = rdbms.AlertDestinationTypeKeyInternalLogger
	default:
		return ProjectAlertDestination{}, fmt.Errorf("unknown type: %s", typeAlias)
	}
	adt := &rdbms.AlertDestinationType{}
	err := tx.Model(&rdbms.AlertDestinationType{}).Where("key = ?", searchKey).First(&adt).Error
	if err != nil {
		return ProjectAlertDestination{}, err
	}
	d := rdbms.ProjectAlertDestination{
		ProjectID:              projectID,
		AlertDestinationTypeID: adt.ID,
	}
	if res := tx.Create(&d); res.Error != nil {
		return ProjectAlertDestination{}, tx.Error
	}
	if webhookCfg == nil {
		return ProjectAlertDestination{
			BaseModel: BaseModel{
				ID:        d.ID,
				CreatedAt: d.CreatedAt,
				UpdatedAt: d.UpdatedAt,
			},
			ProjectID:              projectID,
			AlertDestinationTypeID: adt.ID,
		}, nil
	}

	wc := rdbms.AlertDestinationNotificationWebhookConfiguration{
		ProjectAlertDestinationID: d.ID,
		URL:                       webhookCfg.URL,
		Headers:                   webhookCfg.Headers,
	}

	if res := tx.Create(&wc); res.Error != nil {
		return ProjectAlertDestination{}, tx.Error
	}
	return ProjectAlertDestination{
		BaseModel: BaseModel{
			ID:        d.ID,
			CreatedAt: d.CreatedAt,
			UpdatedAt: d.UpdatedAt,
		},
		ProjectID:              projectID,
		AlertDestinationTypeID: adt.ID,
		WebhookConfiguration: &AlertDestinationNotificationWebhookConfiguration{
			BaseModel: BaseModel{
				ID:        wc.ID,
				CreatedAt: wc.CreatedAt,
				UpdatedAt: wc.UpdatedAt,
			},
			URL:     wc.URL,
			Headers: wc.Headers,
		},
	}, nil
}
