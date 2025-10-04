package service

import (
	"context"
	"log"
	"os"
	"time"

	apikey "github.com/georgepsarakis/chi-api-key-auth"
	"github.com/go-chi/chi/v5"
	"go.uber.org/zap"
	"golang.org/x/sync/errgroup"

	"github.com/georgepsarakis/periscope/alerting"
	"github.com/georgepsarakis/periscope/app"
	"github.com/georgepsarakis/periscope/http"
	"github.com/georgepsarakis/periscope/ingestion"
	"github.com/georgepsarakis/periscope/newcontext"
)

type OnShutdownErrorFunc func(error)
type CleanupFunc func() error

type Options struct {
	OSSignalListenerDisabled bool
}

func NewHTTPService(opts Options) (*http.Server, CleanupFunc, OnShutdownErrorFunc) {
	application, cleanup, err := app.New()
	if err != nil {
		panic(err)
	}

	httpServer := http.NewServer(application.Logger, http.NetworkAddress{
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

	eventHandler := http.NewEventHandler(application, aggr)

	r := http.NewRouter(application)
	r.Post("/api/{project_id}/envelope", eventHandler.IngestionHandler())
	prjHandler := http.NewProjectHandler(application)
	alertHandler := http.NewAlertHandler(application)
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
		r.Get("/projects/{project_id}/alerts", alertHandler.List)
	})
	httpServer.SetHandler(r)

	return httpServer, cleanup, func(err error) {
		application.Logger.Error("error during server shutdown", zap.Error(err))
		if err := cleanup(); err != nil {
			log.Fatal(err.Error())
		}
		os.Exit(1)
	}
}
