package service

import (
	"context"
	"errors"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	apikey "github.com/georgepsarakis/chi-api-key-auth"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/georgepsarakis/periscope/alerting"
	"github.com/georgepsarakis/periscope/app"
	periscopeHttp "github.com/georgepsarakis/periscope/http"
	"github.com/georgepsarakis/periscope/ingestion"
	"github.com/georgepsarakis/periscope/newcontext"
	"github.com/georgepsarakis/periscope/repository"
)

type OnShutdownErrorFunc func(error)
type CleanupFunc func() error

type Options struct {
	OSSignalListenerDisabled bool
}

func NewHTTPService(opts Options) (*periscopeHttp.Server, CleanupFunc, OnShutdownErrorFunc) {
	application, cleanup, err := app.New()
	if err != nil {
		panic(err)
	}

	httpServer := periscopeHttp.NewServer(application.Logger, periscopeHttp.NetworkAddress{
		Host: application.HTTPServerHost(),
		Port: application.HTTPServerListeningPort(),
	})

	ctx := newcontext.WithLogger(context.Background(), application.Logger)
	ctx, cancel := context.WithCancel(ctx)

	aggr := ingestion.NewAggregator(application.Logger)
	if err := aggr.Subscribe(ctx); err != nil {
		application.Logger.Fatal("aggregator subscribe failed", zap.Error(err))
	}

	grp, ctx := errgroup.WithContext(ctx)
	grp.Go(aggr.Consumer(ctx))
	grp.Go(httpServer.ShutdownHandler(!opts.OSSignalListenerDisabled, func() error {
		application.Logger.Info("running shutdown callback")
		cancel()
		return nil
	}))
	grp.Go(ingestion.NewPersistence(application, time.Second, aggr).Scheduler(ctx))
	grp.Go(alerting.NewAlerting(application, time.Second).Scheduler(ctx))
	httpServer.OnShutdown(grp.Wait)

	eventHandler := periscopeHttp.NewEventHandler(application, aggr)
	adtHandler := periscopeHttp.NewAlertDestinationHandler(application)

	r := periscopeHttp.NewRouter(application)
	r.Post("/api/{project_id}/envelope", eventHandler.IngestionHandler())
	prjHandler := periscopeHttp.NewProjectHandler(application)
	alertHandler := periscopeHttp.NewAlertHandler(application)
	r.Route("/api/admin", func(r chi.Router) {
		apiKeyOpts := apikey.Options{
			SecretProvider: &apikey.EnvironmentSecretProvider{
				CurrentSecretHeaderName: "API_SECRET_KEY_ADMIN",
			},
			HeaderAuthProvider: apikey.AuthorizationHeader{},
		}
		r.Use(apikey.Authorize(apiKeyOpts))
		r.Post("/projects", prjHandler.Create)
		r.Get("/projects/{id}", prjHandler.Read)
		r.Group(func(r chi.Router) {
			r.Use(func(next http.Handler) http.Handler {
				return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					projectID := chi.URLParam(r, "project_id")
					pid, err := strconv.Atoi(projectID)
					if err != nil {
						w.WriteHeader(http.StatusBadRequest)
						return
					}
					ctx := r.Context()
					_, err = application.Repository.ProjectFindByID(ctx, uint(pid))
					if err != nil {
						if errors.Is(err, repository.ErrRecordNotFound) {
							w.WriteHeader(http.StatusNotFound)
							return
						}
						w.WriteHeader(http.StatusInternalServerError)
						application.Logger.Error(
							"failed to retrieve project",
							zap.String("project_id", projectID), zap.Error(err))
						return
					}
					next.ServeHTTP(w, r)
				})
			})
			r.Get("/projects/{project_id}/alerts", alertHandler.List)
			r.Post("/projects/{project_id}/alert_notification_destinations", adtHandler.Create)
		})
	})
	httpServer.SetHandler(r)

	return httpServer, cleanup, func(err error) {
		if err == nil {
			return
		}
		log.Println("error during server shutdown:" + err.Error())
		if err := cleanup(); err != nil {
			log.Fatal(err.Error())
		}
		os.Exit(1)
	}
}
