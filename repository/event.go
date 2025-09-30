package repository

import (
	"context"
	"database/sql"
	"errors"

	"gorm.io/gorm"

	"github.com/georgepsarakis/periscope/repository/rdbms"
)

func (r *Repository) CreateEvents(ctx context.Context, project Project, events []Event) (EventGroup, []*Event, error) {
	if len(events) == 0 {
		return EventGroup{}, nil, nil
	}
	var dbGroup rdbms.EventGroup
	groupKey := events[0].Fingerprint
	tx := r.database.WithContext(ctx).Where(
		&EventGroup{ProjectID: project.ID, AggregationKey: groupKey},
	).First(&dbGroup)
	if tx.Error != nil && !errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		return EventGroup{}, nil, tx.Error
	}
	if errors.Is(tx.Error, gorm.ErrRecordNotFound) {
		if err := r.database.Transaction(func(tx *gorm.DB) error {
			dbGroup = rdbms.EventGroup{
				EventReceivedAt:  r.now(),
				ProjectID:        project.ID,
				AggregationKey:   groupKey,
				TotalCount:       1,
				AlertTriggeredAt: sql.NullTime{Time: r.now(), Valid: true},
			}
			if tx := tx.WithContext(ctx).Create(&dbGroup); tx.Error != nil {
				return tx.Error
			}
			alert := rdbms.Alert{
				EventGroupID: dbGroup.ID,
				ProjectID:    project.ID,
				TriggeredAt:  r.now(),
				Title:        events[0].Title,
				Description:  events[0].Fingerprint + "\n" + string(events[0].StackTrace),
			}
			if tx := tx.WithContext(ctx).Create(&alert); tx.Error != nil {
				return tx.Error
			}
			return nil
		}); err != nil {
			return EventGroup{}, nil, err
		}
	} else {
		u := map[string]any{
			"total_count":       gorm.Expr("total_count + ?", len(events)),
			"event_received_at": r.now(),
		}
		if tx := r.database.WithContext(ctx).Model(&dbGroup).Updates(u); tx.Error != nil {
			return EventGroup{}, nil, tx.Error
		}
	}
	group := EventGroup{
		BaseModel: BaseModel{
			ID:        dbGroup.ID,
			CreatedAt: dbGroup.CreatedAt,
			UpdatedAt: dbGroup.UpdatedAt,
		},
		TotalCount:      dbGroup.TotalCount,
		EventReceivedAt: dbGroup.EventReceivedAt,
		ProjectID:       dbGroup.ProjectID,
		AggregationKey:  dbGroup.AggregationKey,
	}
	var newEvents []*rdbms.Event
	for _, event := range events {
		newEvents = append(newEvents, &rdbms.Event{
			EventID:      event.EventID,
			EventGroupID: dbGroup.ID,
			Fingerprint:  event.Fingerprint,
			ProjectID:    project.ID,
			Title:        event.Title,
			EmittedAt:    r.now(), // TODO: replace with client event timestamp
		})
	}
	if tx := r.database.WithContext(ctx).Create(&newEvents); tx.Error != nil {
		return EventGroup{}, nil, tx.Error
	}
	re := make([]*Event, 0, len(newEvents))
	for _, event := range newEvents {
		re = append(re, &Event{
			EventID:      event.EventID,
			Title:        event.Title,
			Fingerprint:  event.Fingerprint,
			EventGroupID: event.EventGroupID,
			ProjectID:    event.ProjectID,
			EmittedAt:    event.EmittedAt,
		})
	}
	return group, re, nil
}

func (r *Repository) EventFindLatestByProjectAndEventGroup(ctx context.Context, projectID, eventGroupID uint) (Event, error) {
	tx := r.dbExecutor(ctx)
	ev := rdbms.Event{}
	res := tx.Model(&ev).Where("project_id = ? AND event_group_id = ?", projectID, eventGroupID).
		Order("created_at DESC").First(&ev)
	if res.Error != nil {
		return Event{}, res.Error
	}
	return Event{
		BaseModel: BaseModel{
			ID:        ev.ID,
			CreatedAt: ev.CreatedAt,
			UpdatedAt: ev.UpdatedAt,
		},
		EventID:      ev.EventID,
		ProjectID:    ev.ProjectID,
		EventGroupID: ev.EventGroupID,
		EmittedAt:    ev.EmittedAt,
		Fingerprint:  ev.Fingerprint,
		StackTrace:   ev.StackTrace,
	}, nil
}
