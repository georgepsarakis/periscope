package http

import (
	"net/http"
	"net/http/httputil"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/go-chi/cors"
	"go.uber.org/zap"

	"github.com/georgepsarakis/periscope/app"
	"github.com/georgepsarakis/periscope/newcontext"
)

type requestLogger struct {
	middleware.LoggerInterface
	logger *zap.Logger
}

func (rl requestLogger) Print(v ...any) {
	rl.logger.Info("request completed", zap.Any("accessLog", v))
}

func NewRouter(application app.App) *chi.Mux {
	r := chi.NewRouter()

	middleware.DefaultLogger = middleware.RequestLogger(
		&middleware.DefaultLogFormatter{
			Logger:  requestLogger{logger: application.Logger},
			NoColor: true,
		},
	)

	r.Use(middleware.RequestID)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			next.ServeHTTP(w, r)
			w.Header().Add(middleware.RequestIDHeader, middleware.GetReqID(r.Context()))
		})
	})
	r.Use(middleware.RealIP)
	r.Use(middleware.Logger)
	// Inject logger in the context
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			ctx := newcontext.WithLogger(r.Context(), application.Logger)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	})
	r.Use(cors.Handler(cors.Options{
		// TODO: transfer to configuration
		AllowedOrigins:   application.HTTPAllowedOrigins(),
		AllowedMethods:   []string{http.MethodGet, http.MethodPost, http.MethodPut, http.MethodDelete, http.MethodOptions},
		AllowedHeaders:   []string{"Accept", "Authorization", "Content-Type"},
		AllowCredentials: false,
		MaxAge:           300, // Maximum value not ignored by any of major browsers
	}))
	r.Use(middleware.Recoverer)
	r.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Add("Content-Type", "application/json")
			next.ServeHTTP(w, r)
		})
	})

	r.Use(middleware.Timeout(application.HTTPServerRequestTimeout()))
	r.Use(middleware.StripSlashes)
	r.Use(middleware.CleanPath)
	r.Use(middleware.AllowContentType(
		"application/x-sentry-envelope",
		"application/json"))
	if application.DebugEnabled() {
		r.Use(func(next http.Handler) http.Handler {
			return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				rb, err := httputil.DumpRequest(r, true)
				if err != nil {
					w.WriteHeader(http.StatusInternalServerError)
					application.Logger.Error("failed to dump request", zap.Error(err))
				}
				application.Logger.Info(string(rb))
				next.ServeHTTP(w, r)
			})
		})
	}
	return r
}
