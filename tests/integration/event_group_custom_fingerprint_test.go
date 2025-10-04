package main

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/georgepsarakis/go-httpclient"
	"github.com/getsentry/sentry-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/georgepsarakis/periscope/http"
	"github.com/georgepsarakis/periscope/repository"
	"github.com/georgepsarakis/periscope/service"
)

func TestEventForwarding_CustomFingerprint(t *testing.T) {
	t.Setenv("API_SECRET_KEY_ADMIN",
		repository.RandomString(repository.CharsetAlphanumeric, 10))

	tempFile, err := os.CreateTemp("", "tmp-*.db")
	require.NoError(t, err)
	tempFilePath := tempFile.Name()
	t.Cleanup(func() {
		os.Remove(tempFilePath)
	})
	t.Logf("temporary database file: %s", tempFilePath)
	t.Setenv("SQLITE_PATH", tempFilePath)
	server, cleanup, _ := service.NewHTTPService(service.Options{OSSignalListenerDisabled: true})
	go func() {
		require.NoError(t, server.Run())
	}()
	t.Cleanup(func() {
		cleanup() //nolint:errcheck
		require.NoError(t, server.Close())
	})

	time.Sleep(time.Second)

	baseURL := url.URL{
		Scheme: "http",
		Path:   "api/admin/",
		Host:   server.Address(),
	}
	adminAPIClient, err := httpclient.New().WithDefaultHeaders(map[string]string{
		"Content-Type":  "application/json",
		"Authorization": fmt.Sprintf("Bearer %v", os.Getenv("API_SECRET_KEY_ADMIN")),
	}).WithBaseURL(baseURL.String())
	require.NoError(t, err)

	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	resp, err := adminAPIClient.Post(ctx, "projects", strings.NewReader(`{"name": "test project 1"}`))
	require.NoError(t, err)

	p := http.ProjectCreateResponse{}
	require.NoError(t, httpclient.DeserializeJSON(resp, &p))

	ingestionKey := p.Project.IngestionAPIKeys[0]

	sentryClient, err := sentry.NewClient(sentry.ClientOptions{
		Dsn:   fmt.Sprintf("http://%s@%s/%s", ingestionKey, server.Address(), p.Project.PublicID),
		Debug: true,
	})
	require.NoError(t, err)
	hub := sentry.NewHub(sentryClient, sentry.NewScope())
	defer hub.Flush(time.Second)
	hub.ConfigureScope(func(scope *sentry.Scope) {
		scope.SetLevel(sentry.LevelError)
		scope.SetFingerprint([]string{"test", "1", "2", "3"})
		for n := range 10 {
			hub.CaptureException(fmt.Errorf("test error %d", n))
		}
	})
	time.Sleep(time.Second)
	resp, err = adminAPIClient.Get(ctx, fmt.Sprintf("projects/%d/alerts", p.Project.ID))
	require.NoError(t, err)
	alertList := http.AlertListResponse{}
	require.NoError(t, httpclient.DeserializeJSON(resp, &alertList))
	require.NotEmpty(t, alertList.Alerts)
	assert.Equal(t, p.Project.ID, alertList.Alerts[0].ProjectID)
}
