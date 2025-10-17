package alerting

import (
	"context"
	"errors"
	"fmt"
	"runtime/debug"
	"strconv"
	"time"

	"github.com/georgepsarakis/go-httpclient"
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
		log := a.application.Logger
		defer func() {
			if r := recover(); r != nil {
				log.Error("panic recovered",
					zap.String("panic", fmt.Sprintf("%s", r)),
					zap.Any("stacktrace", string(debug.Stack())))
			}
		}()
		ticker := time.NewTicker(a.schedulerInterval)
		defer ticker.Stop()

		adt, err := a.application.Repository.AlertDestinationTypeFindAll(ctx)
		if err != nil {
			log.Warn("unable to get alerting destination type", zap.Error(err))
			return err
		}

		destinationTypesById := make(map[uint]repository.AlertDestinationType)
		for _, d := range adt {
			destinationTypesById[d.ID] = d
		}

		destinationChannels := make(map[uint]notification.Channel)
		for _, d := range adt {
			switch d.Key {
			case rdbms.AlertDestinationTypeKeyInternalLogger:
				destinationChannels[d.ID] = notification.LogNotifier{
					Logger: log,
				}
			case rdbms.AlertDestinationTypeKeyGenericWebhook:
				// TODO: change interface for webhooks channel
				destinationChannels[d.ID] = notification.NewGenericWebhookNotification(
					notification.GenericWebhookNotificationSettings{
						HTTPClient: httpclient.New(),
					},
				)
			}
		}

		log.Info("alerting ticker started")
		defer log.Info("alerting ticker stopped")
		for {
			select {
			case <-ticker.C:
				// TODO: add watcher goroutine and emit heartbeats from all background workers
				alert, err := a.application.Repository.AlertFindByNotNotified(ctx)
				if err != nil && !errors.Is(err, repository.ErrRecordNotFound) {
					log.Error("failed to query alerting status", zap.Error(err))
					continue
				}
				if !errors.Is(err, repository.ErrRecordNotFound) {
					if err := a.alertNotifications(ctx, alert); err != nil {
						log.Error("failed to create alert notifications", zap.Error(err))
						continue
					}
				}

				n, err := a.application.Repository.FindAlertDestinationNotificationByNonCompleted(ctx)
				if err != nil {
					if !errors.Is(err, repository.ErrRecordNotFound) {
						log.Error("failed to query alerting destination notifications", zap.Error(err))
					}
					continue
				}
				ad, err := a.application.Repository.FindAlertDestinationByID(ctx, n.ProjectAlertDestinationID)
				if err != nil {
					log.Error("failed to find project alert destination", zap.Error(err))
					continue
				}

				log.Info("notifications sent",
					zap.Uint("project_id", ad.ProjectID),
					zap.Uint("alert_id", alert.ID),
					zap.Uint("alert_destination_type_id", ad.AlertDestinationTypeID))

				ev, err := a.application.Repository.EventFindLatestByProjectAndEventGroup(ctx, alert.ProjectID, alert.EventGroupID)
				if err != nil && !errors.Is(err, repository.ErrRecordNotFound) {
					log.Error("failed to query alerting events", zap.Error(err))
					continue
				}
				destinationType := destinationTypesById[ad.AlertDestinationTypeID]
				var ch notification.Channel
				switch destinationType.Key {
				case rdbms.AlertDestinationTypeKeyGenericWebhook:
					ch = notification.NewGenericWebhookNotification(notification.GenericWebhookNotificationSettings{
						HTTPClient:  httpclient.New(),
						WebhookURL:  ad.WebhookConfiguration.URL,
						HTTPHeaders: ad.WebhookConfiguration.Headers,
					})
				case rdbms.AlertDestinationTypeKeyInternalLogger:
					ch = notification.LogNotifier{
						Logger: log,
					}
				}
				if err := ch.Emit(ctx, notification.Event{
					ID:    ev.EventID,
					Type:  strconv.Itoa(int(ev.EventGroupID)),
					Alert: alert,
					Details: notification.EventDetails{
						Title:        ev.Title,
						AlertID:      strconv.Itoa(int(alert.ID)),
						EventGroupID: strconv.Itoa(int(alert.EventGroupID)),
						ProjectID:    strconv.Itoa(int(alert.ProjectID)),
					},
				}); err != nil {
					log.Error("failed to emit alerting event", zap.Error(err))
				}

				_, err = a.application.Repository.AlertDestinationNotificationUpdateCompletedAt(ctx, n.ID, time.Now().UTC())
				if err != nil && !errors.Is(err, repository.ErrRecordNotFound) {
					log.Error("failed to update alert destination notification", zap.Error(err))
					continue
				}
			case <-ctx.Done():
				log.Info("alerting stopped due to timeout or cancellation")
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
		ads, err := a.application.Repository.FindAlertDestinationsByProjectID(ctx, alert.ProjectID)
		if err != nil {
			return fmt.Errorf("failed to find alerting destinations: %w", err)
		}
		for _, ad := range ads {
			_, err = a.application.Repository.CreateAlertDestinationNotification(ctx, alert.ID, ad.ID)
			if err != nil {
				return fmt.Errorf("failed to create alert destination notification: %w", err)
			}
		}
		return nil
	})
}
