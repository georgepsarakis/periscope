package app

import (
	"context"
	"net"
	"net/url"
	"strconv"
	"time"

	"github.com/sethvargo/go-envconfig"
	"go.uber.org/zap"
	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"github.com/georgepsarakis/periscope/repository"
	"github.com/georgepsarakis/periscope/repository/rdbms"
)

type Configuration struct {
	Port              int           `env:"PORT,default=8000"`
	Host              string        `env:"HOST,default=localhost"`
	AllowedOrigins    []string      `env:"ALLOWED_ORIGINS,default=http://localhost"`
	RequestTimeout    time.Duration `env:"REQUEST_TIMEOUT,default=30s"`
	PostgresHost      string        `env:"POSTGRES_HOST,default=localhost"`
	PostgresPort      int           `env:"POSTGRES_PORT,default=5432"`
	PostgresUser      string        `env:"POSTGRES_USER,default=pguser"`
	PostgresPassword  string        `env:"POSTGRES_PASSWORD"`
	PostgresDatabase  string        `env:"POSTGRES_DATABASE,default=periscope"`
	PostgresEnabled   bool          `env:"POSTGRES_ENABLED,default=false"`
	SqlitePath        string        `env:"SQLITE_PATH,default=tmp/periscope.db"`
	Debug             bool          `env:"DEBUG,default=false"`
	ApiSecretKeyAdmin string        `env:"API_SECRET_KEY_ADMIN"`
}

type App struct {
	cfg                Configuration
	Logger             *zap.Logger
	PostgresConnection *gorm.DB
	Repository         *repository.Repository
}

func (a App) HTTPServerRequestTimeout() time.Duration {
	return a.cfg.RequestTimeout
}

func (a App) HTTPServerListeningPort() int {
	return a.cfg.Port
}

func (a App) HTTPServerHost() string {
	return a.cfg.Host
}

func (a App) HTTPAllowedOrigins() []string {
	return a.cfg.AllowedOrigins
}

func (a App) DebugEnabled() bool {
	return a.cfg.Debug
}

func New() (App, func() error, error) {
	app := App{}
	cfg := Configuration{}
	envconfig.MustProcess(context.Background(), &cfg)
	app.cfg = cfg

	appLogger, _ := zap.NewProduction()
	app.Logger = appLogger

	var dbLogger logger.Interface
	if app.cfg.Debug {
		dbLogger = logger.Default.LogMode(logger.Silent)
	} else {
		dbLogger = logger.Discard
	}
	gormCfg := gorm.Config{
		Logger: dbLogger,
	}

	var database *gorm.DB
	if app.cfg.PostgresEnabled {
		dsn := url.URL{
			Scheme: "postgres",
			Host:   net.JoinHostPort(app.cfg.PostgresHost, strconv.Itoa(app.cfg.PostgresPort)),
			Path:   app.cfg.PostgresDatabase,
			User:   url.UserPassword(app.cfg.PostgresUser, app.cfg.PostgresPassword),
		}
		app.cfg.PostgresPassword = ""
		db, err := gorm.Open(postgres.New(postgres.Config{DSN: dsn.String()}), &gormCfg)
		if err != nil {
			return app, nil, err
		}
		app.PostgresConnection = db
		database = db
	} else {
		db, err := gorm.Open(sqlite.Open(cfg.SqlitePath), &gormCfg)
		if err != nil {
			panic("failed to connect database")
		}
		database = db
	}
	if !app.cfg.PostgresEnabled {
		app.Repository = repository.New(database)
		if err := database.AutoMigrate(
			&rdbms.EventGroup{},
			&rdbms.Event{},
			&rdbms.Project{},
			&rdbms.Alert{},
			&rdbms.AlertDestinationNotification{},
			&rdbms.AlertDestinationType{},
			&rdbms.ProjectAlertDestination{},
			&rdbms.ProjectIngestionAPIKey{},
			&rdbms.AlertDestinationNotificationWebhookConfiguration{},
		); err != nil {
			panic(err)
		}
	}

	return app, func() error {
		return app.Logger.Sync()
	}, nil
}
