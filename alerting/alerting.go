package alerting

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strconv"
	"time"

	"go.uber.org/zap"
	"gorm.io/gorm"

	"github.com/georgepsarakis/periscope/app"
	"github.com/georgepsarakis/periscope/newcontext"
	"github.com/georgepsarakis/periscope/notification"
	"github.com/georgepsarakis/periscope/repository"
	"github.com/georgepsarakis/periscope/repository/rdbms"
)

type Alerting struct {
	application       app.App
	schedulerInterval time.Duration
}

func NewAlerting(application app.App, schedulerInterval time.Duration) Alerting {
	return Alerting{
		application:       application,
		schedulerInterval: schedulerInterval,
	}
}

func (a Alerting) Scheduler(ctx context.Context) func() error {
	return func() error {
		ticker := time.NewTicker(a.schedulerInterval)
		defer ticker.Stop()
		logger := a.application.Logger

		adt, err := a.application.Repository.AlertDestinationTypeFindAll(ctx)
		if err != nil {
			logger.Warn("unable to get alerting destination type", zap.Error(err))
			return err
		}

		destinationChannels := make(map[uint]notification.Channel)
		for _, d := range adt {
			switch d.Key {
			case rdbms.AlertDestinationTypeKeyInternalLogger:
				destinationChannels[d.ID] = notification.LogNotifier{
					Logger: logger,
				}
			}
		}

		logger.Info("alerting ticker started")
		defer logger.Info("alerting ticker stopped")
		for {
			select {
			case <-ticker.C:
				// TODO: add watcher goroutine and emit heartbeats from all background workers
				alert, err := a.application.Repository.AlertFindByNotNotified(ctx)
				if err != nil && !errors.Is(err, repository.ErrRecordNotFound) {
					logger.Error("failed to query alerting status", zap.Error(err))
				} else if errors.Is(err, repository.ErrRecordNotFound) {
					continue
				}
				if err := a.alertNotifications(ctx, alert); err != nil {
					logger.Error("failed to create alert notifications", zap.Error(err))
				}

				n, err := a.application.Repository.FindAlertDestinationNotificationByNonCompleted(ctx)
				if err != nil && !errors.Is(err, repository.ErrRecordNotFound) {
					logger.Error("failed to query alerting destination notifications", zap.Error(err))
					continue
				}
				ad, err := a.application.Repository.FindAlertDestinationByID(ctx, n.ProjectAlertDestinationID)
				if err != nil && !errors.Is(err, repository.ErrRecordNotFound) {
					logger.Error("failed to find project alert destination", zap.Error(err))
					continue
				}

				logger.Info("notifications sent",
					zap.Uint("project_id", ad.ProjectID),
					zap.Uint("alert_id", alert.ID),
					zap.Uint("alert_destination_type_id", ad.AlertDestinationTypeID))

				ev, err := a.application.Repository.EventFindLatestByProjectAndEventGroup(ctx, alert.ProjectID, alert.EventGroupID)
				if err != nil && !errors.Is(err, repository.ErrRecordNotFound) {
					logger.Error("failed to query alerting events", zap.Error(err))
					continue
				}
				ch := destinationChannels[ad.AlertDestinationTypeID]
				b, err := json.Marshal(ev)
				if err != nil {
					logger.Error("failed to marshal alerting event", zap.Error(err))
					continue
				}
				if err := ch.Emit(ctx, notification.Event{
					ID:   ev.EventID,
					Type: strconv.Itoa(int(ev.EventGroupID)),
					Data: b,
					Attributes: notification.EventAttributes{
						Title:        ev.Title,
						AlertID:      strconv.Itoa(int(alert.ID)),
						EventGroupID: strconv.Itoa(int(alert.EventGroupID)),
						ProjectID:    strconv.Itoa(int(alert.ProjectID)),
					},
				}); err != nil {
					logger.Error("failed to emit alerting event", zap.Error(err))
				}

				_, err = a.application.Repository.AlertDestinationNotificationUpdateCompletedAt(ctx, n.ID, time.Now().UTC())
				if err != nil && !errors.Is(err, repository.ErrRecordNotFound) {
					logger.Error("failed to update alert destination notification", zap.Error(err))
					continue
				}
			case <-ctx.Done():
				logger.Info("alerting stopped due to timeout or cancellation")
				return nil
			}
		}
	}
}

func (a Alerting) alertNotifications(ctx context.Context, alert repository.Alert) error {
	ctx, cancel := context.WithTimeout(ctx, time.Second)
	defer cancel()
	return a.application.Repository.NewTransaction(func(tx *gorm.DB) error {
		ctx := newcontext.WithDBTransaction(ctx, tx)
		_, err := a.application.Repository.AlertUpdateNotifiedAt(ctx, alert.ID, time.Now().UTC())
		if err != nil {
			return fmt.Errorf("failed to update alerting status: %w", err)
		}
		_, err = a.application.Repository.CreateAlertDestinationNotification(ctx, alert.ID, 1)
		if err != nil {
			return fmt.Errorf("failed to create alert destination notification: %w", err)
		}
		return nil
	})
}
