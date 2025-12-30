package app

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/BurntSushi/toml"

	"immich-photo-frame/internal/immich"
)

// Config is the top-level configuration struct that is loaded via TOML
// decoding of the file specified by the IMMICH_PHOTO_FRAME_CONFIG environment
// variable (or "config.toml" if empty).
//
// This is the primary way to configure the application.
type Config struct {
	immich.Config
	App struct{ ImmichAlbums []string }
}

type app struct {
	client *immich.Client
	conf   Config
}

func (a *app) run() error {
	albums, err := a.getConfiguredAlbums()
	if err != nil {
		return err
	}
	if n := countAssets(albums); n == 0 {
		return errors.New("no assets found")
	}
	iter := NewAssetMetadataIter(a.client, albums)
	for asset := iter.Next(); asset != nil; asset = iter.Next() {
		// img, err := a.client.GetAsset(*asset)
		// if err != nil {
		// 	slog.Error("failed to get asset", "asset", asset, "error", err)
		// 	continue
		// }
		// a.showImage()
	}
	return nil
}

func (a *app) getConfiguredAlbums() ([]immich.Album, error) {
	// Get all albums.
	allAlbums, err := a.client.GetAlbums()
	if err != nil {
		return nil, err
	}
	slog.Info("found albums", "count", len(allAlbums))

	// Build set of configured album names.
	configuredAlbumNames := a.conf.App.ImmichAlbums
	albumNameSet := make(map[string]struct{})
	for _, album := range configuredAlbumNames {
		albumNameSet[album] = struct{}{}
	}

	// Iterate through all albums and build a list of the albums that are found in the set.
	var configuredAlbums []immich.Album
	foundAlbums := make(map[string]struct{})
	for _, album := range allAlbums {
		if _, ok := albumNameSet[album.Name]; ok {
			slog.Info("found album", "name", album.Name, "id", album.ID, "asset_count", album.AssetCount)
			configuredAlbums = append(configuredAlbums, album)
			foundAlbums[album.Name] = struct{}{}
		}
	}

	// Log if we didn't find some of the albums that were configured.
	if len(foundAlbums) != len(albumNameSet) {
		var albumsMissing []string
		for albumName := range albumNameSet {
			if _, ok := foundAlbums[albumName]; !ok {
				albumsMissing = append(albumsMissing, albumName)
			}
		}
		slog.Warn("some albums not found", "albums_missing", albumsMissing)
	}
	return configuredAlbums, nil
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

	// TOML-decode config file contents.
	var conf Config
	if _, err := toml.DecodeFile(configFilePath, &conf); err != nil {
		return nil, err
	}

	// Load values from environment variables.
	conf.Remote.HydrateFromEnv()

	return &conf, nil
}

func InitApp(conf Config) (*app, error) {
	client := immich.NewClient(
		immich.WithRemote(conf.Remote),
		immich.WithLocalStorage(conf.LocalStorage),
		immich.WithInMemoryCache(conf.InMemoryCache),
	)
	slog.Info("created immich client")
	slog.Info("client diagnostics", "diagnostics", client.Diagnostics())
	return &app{client: client, conf: conf}, nil
}

func countAssets(albums []immich.Album) int {
	n := 0
	for _, album := range albums {
		n += album.AssetCount
	}
	return n
}
