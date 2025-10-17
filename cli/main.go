package main

import (
	"context"
	"errors"
	"log"
	"os"
	"time"

	"github.com/getsentry/sentry-go"
	"github.com/urfave/cli/v3"
)

func main() {
	sentryClient, err := sentry.NewClient(sentry.ClientOptions{
		Dsn:   os.Getenv("PERISCOPE_DSN"),
		Debug: true,
	})
	if err != nil {
		log.Fatal(err)
	}
	cmd := &cli.Command{
		Name:  "send-event",
		Usage: "Send event to a Periscope instance",
		Flags: []cli.Flag{
			&cli.StringFlag{
				Name: "fingerprint",
			},
			&cli.StringFlag{
				Name: "message",
			},
		},
		Action: func(ctx context.Context, cmd *cli.Command) error {
			hub := sentry.NewHub(sentryClient, sentry.NewScope())
			defer hub.Flush(time.Second)
			hub.ConfigureScope(func(scope *sentry.Scope) {
				scope.SetLevel(sentry.LevelError)
				scope.SetFingerprint([]string{cmd.String("fingerprint")})
				hub.CaptureException(errors.New(cmd.String("message")))
			})
			return nil
		},
	}

	if err := cmd.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
