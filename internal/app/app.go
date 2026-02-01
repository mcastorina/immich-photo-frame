package app

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"github.com/BurntSushi/toml"

	"immich-photo-frame/internal/app/controller"
	"immich-photo-frame/internal/app/controller/planners"
	"immich-photo-frame/internal/app/display"
	"immich-photo-frame/internal/immich"
)

// Config is the top-level configuration struct that is loaded via TOML
// decoding of the file specified by the IMMICH_PHOTO_FRAME_CONFIG environment
// variable (or "config.toml" if empty).
//
// This is the primary way to configure the application.
type Config struct {
	immich.Config
	App struct {
		ControllerConfig
		DisplayConfig
		ImmichAlbumRefreshInterval time.Duration
	}
}

type DisplayConfig = display.Config
type ControllerConfig = controller.Config

type photoFrame struct {
	conf   Config
	client *immich.Client
}

func (pf *photoFrame) run() error {
	disp := display.New(pf.conf.App.DisplayConfig)
	ctrl, err := controller.New(pf.conf.App.ControllerConfig, pf.client, disp)
	if err != nil {
		return err
	}

	disp.SetKeyBinds(func(ke *fyne.KeyEvent) {
		switch ke.Name {
		case fyne.KeyRight:
			ctrl.Next()
		case fyne.KeyLeft:
			ctrl.Prev()
		}
	})

	go ctrl.Run()

	disp.ShowAndRun()
	return nil
}

func Run() error {
	conf, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	// Debug level since conf has sensitive values.
	slog.Debug("loaded config", "config", conf)

	app, err := InitApp(*conf)
	if err != nil {
		return fmt.Errorf("failed to init app: %w", err)
	}
	slog.Info("successfully initialized app")
	return app.run()
}

func LoadConfig() (*Config, error) {
	// Determine config file path.
	configFilePath := "config.toml"
	if envConfigFilePath := os.Getenv("IMMICH_PHOTO_FRAME_CONFIG"); envConfigFilePath != "" {
		configFilePath = envConfigFilePath
	}
	if _, err := os.Stat(configFilePath); os.IsNotExist(err) {
		return nil, errors.New("config file not found")
	} else if err != nil {
		return nil, err
	}

	// Set config defaults.
	var conf Config
	conf.App.ImageDelay = 5 * time.Second
	conf.App.ImageScale = 1
	conf.App.HistorySize = 10
	conf.App.PlanAlgorithm.PlanIter = &planners.Sequential{}
	conf.App.ImmichAlbumRefreshInterval = 24 * time.Hour

	// TOML-decode config file contents.
	if _, err := toml.DecodeFile(configFilePath, &conf); err != nil {
		return nil, err
	}

	// Load values from environment variables.
	conf.Remote.HydrateFromEnv()
	conf.LocalStorage.LocalStoragePath = os.ExpandEnv(conf.LocalStorage.LocalStoragePath)

	// Validate config values.
	if err := conf.LocalStorage.Valid(); err != nil {
		return nil, err
	}
	if conf.App.ImageScale <= 0 || conf.App.ImageScale > 1 {
		slog.Warn("invalid imageScale value, resetting to default",
			"error", "expected a value between 0 and 1",
		)
		conf.App.ImageScale = 1
	}
	if conf.App.HistorySize < 0 {
		slog.Warn("invalid historySize value, setting to 0",
			"error", "historySize must be at least 0",
		)
		conf.App.HistorySize = 0
	}

	return &conf, nil
}

func InitApp(conf Config) (*photoFrame, error) {
	client := immich.NewClient(
		immich.WithRemote(conf.Remote),
		immich.WithLocalStorage(conf.LocalStorage),
		immich.WithInMemoryCache(conf.InMemoryCache),
		immich.WithRefreshInterval(conf.App.ImmichAlbumRefreshInterval),
	)
	slog.Info("created immich client")
	slog.Info("client diagnostics", "diagnostics", client.Diagnostics())
	return &photoFrame{
		client: client,
		conf:   conf,
	}, nil
}
