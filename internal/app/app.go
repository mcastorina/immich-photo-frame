package app

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/BurntSushi/toml"

	"immich-photo-frame/internal/immich"
	"immich-photo-frame/internal/immich/cache"
)

// Config is the top-level configuration struct that is loaded via TOML
// decoding of the file specified by the IMMICH_PHOTO_FRAME_CONFIG environment
// variable (or "config.toml" if empty).
//
// This is the primary way to configure the application.
type Config struct {
	cache.Config
	Immich immich.Config
}

type App struct{}

func Run() error {
	conf, err := LoadConfig()
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	_, err = InitApp(*conf)
	if err != nil {
		return fmt.Errorf("failed to init app: %w", err)
	}

	return nil
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

	// TOML-decode config file contents.
	var conf Config
	if _, err := toml.DecodeFile(configFilePath, &conf); err != nil {
		return nil, err
	}

	// Load values from environment variables.
	conf.Immich.HydrateFromEnv()

	return &conf, nil
}

func InitApp(conf Config) (*App, error) {
	client := cache.NewClient(
		cache.WithRemote(conf.Immich),
		cache.WithLocalStorage(conf.LocalStorage),
		cache.WithInMemoryCache(conf.InMemoryCache),
	)
	slog.Info("created immich client", "diagnostics", client.Diagnostics())
	return nil, errors.New("unimplemented")
}
