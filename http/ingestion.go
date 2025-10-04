package http

import (
	"bufio"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"

	"github.com/georgepsarakis/periscope/app"
	"github.com/georgepsarakis/periscope/ingestion"
)

type EventHandler struct {
	app  app.App
	aggr *ingestion.Aggregator
}

func NewEventHandler(app app.App, aggr *ingestion.Aggregator) EventHandler {
	return EventHandler{app: app, aggr: aggr}
}

func (h EventHandler) IngestionHandler() func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		ctx := r.Context()
		logger := h.app.Logger
		var ev Event
		defer r.Body.Close() //nolint:errcheck
		scanner := bufio.NewReader(r.Body)
		// Skip two lines
		for range 2 {
			if _, _, err := scanner.ReadLine(); err != nil {
				logger.Error("unable to scan lines", zap.Error(err))
			}
		}
		if err := json.NewDecoder(scanner).Decode(&ev); err != nil {
			w.WriteHeader(http.StatusBadRequest)
			logger.Error(err.Error())
			return
		}
		projectID := chi.URLParam(r, "project_id")

		project, err := h.app.Repository.ProjectFindByPublicID(ctx, projectID)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			logger.Error("invalid project ID", zap.Error(err))
			return
		}

		authHeader := strings.TrimSpace(r.Header.Get("X-Sentry-Auth"))
		if authHeader == "" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		// TODO: properly parse auth header
		parts := strings.Split(authHeader, "sentry_key=")
		if len(parts) < 2 {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		if !project.HasAccess(parts[1]) {
			w.WriteHeader(http.StatusForbidden)
			return
		}

		publishedEvent := ingestion.Event(ev)
		if err := h.aggr.Publish(project.ID, publishedEvent); err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			logger.Error("failed to publish event", zap.Error(err))
			return
		}

		w.WriteHeader(http.StatusOK)
	}
}
