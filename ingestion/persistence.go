package ingestion

import (
	"context"
	"time"

	"go.uber.org/zap"

	"github.com/georgepsarakis/periscope/app"
	"github.com/georgepsarakis/periscope/repository"
)

type Persistence struct {
	application       app.App
	schedulerInterval time.Duration
	aggregator        *Aggregator
}

func NewPersistence(application app.App, schedulerInterval time.Duration, aggr *Aggregator) Persistence {
	return Persistence{
		application:       application,
		schedulerInterval: schedulerInterval,
		aggregator:        aggr,
	}
}

func (p Persistence) Scheduler(ctx context.Context) func() error {
	return func() error {
		ticker := time.NewTicker(p.schedulerInterval)
		defer ticker.Stop()
		logger := p.application.Logger

		logger.Info("event persistence ticker started")

		for {
			select {
			case <-ticker.C:
				logger.Info("event persistence ticker tick")
				// TODO: timeout should be configurable
				if err := p.flush(); err != nil {
					logger.Error("flush operation failed", zap.Error(err))
				}
			case <-ctx.Done():
				logger.Info("ticker stopped due to timeout or cancellation")
				logger.Info("draining event queue")
				if err := p.flush(); err != nil {
					logger.Error("flush operation failed", zap.Error(err))
				}
				return nil
			}
		}
	}
}

func (p Persistence) flush() error {
	log := p.application.Logger

	// TODO: timeout should be configurable
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	log.Info("event persistence flush started")

	for _, batch := range p.aggregator.Flush() {
		ev := batch[0]
		project, err := p.application.Repository.ProjectFindByID(ctx, ev.ProjectEvent.ProjectID)
		if err != nil {
			log.Error("failed to find project",
				zap.Error(err),
				zap.Uint("project_id", ev.ProjectEvent.ProjectID))
			continue
		}
		events := make([]repository.Event, 0, len(batch))
		for _, event := range batch {
			events = append(events, repository.Event{
				EventID:     event.ProjectEvent.EventID,
				Fingerprint: event.ProjectEvent.Fingerprint,
				ProjectID:   ev.ProjectEvent.ProjectID,
				StackTrace:  ev.ProjectEvent.Trace,
				Title:       event.ProjectEvent.Title,
			})
		}
		grp, createdEvents, err := p.application.Repository.CreateEvents(ctx, project, events)
		if err != nil {
			return err
		}
		log.Info("events persisted successfully",
			zap.Uint("projectID", batch[0].ProjectEvent.ProjectID),
			zap.Uint("eventGroupID", grp.ID),
			zap.Int("count", len(createdEvents)))
	}
	log.Info("event persistence flush completed")

	return nil
}
