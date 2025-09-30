package repository

import (
	"context"
	"errors"
	"time"

	"gorm.io/gorm"

	"github.com/georgepsarakis/periscope/newcontext"
	"github.com/georgepsarakis/periscope/repository/rdbms"
)

var RecordNotFound = errors.New("record not found")

func (r *Repository) AlertFindByNotNotified(ctx context.Context) (Alert, error) {
	alert := rdbms.Alert{}
	if tx := r.database.WithContext(ctx).
		Where("notified_at IS NULL").
		First(&alert); tx.Error != nil {
		if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
			return Alert{}, RecordNotFound
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
			return RecordNotFound
		}
		res = tx.Model(&n).Update("total_attempts", gorm.Expr("total_attempts + ?", 1))
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
	return ProjectAlertDestination{
		BaseModel: BaseModel{
			ID:        pad.ID,
			CreatedAt: pad.CreatedAt,
			UpdatedAt: pad.UpdatedAt,
		},
		ProjectID:              pad.ProjectID,
		AlertDestinationTypeID: pad.AlertDestinationTypeID,
		Configuration:          pad.Configuration,
	}, nil
}

func (r *Repository) AlertDestinationNotificationUpdateCompletedAt(ctx context.Context, id uint, ts time.Time) (AlertDestinationNotification, error) {
	tx := r.dbExecutor(ctx)
	res := tx.Model(&rdbms.AlertDestinationNotification{}).Where("id = ?", id).Updates(map[string]any{"completed_at": ts})
	if res.Error != nil {
		return AlertDestinationNotification{}, res.Error
	}
	return AlertDestinationNotification{}, nil
}
