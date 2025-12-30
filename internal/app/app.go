package app

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
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
	App struct {
		ImmichAlbums []string
		ImageDelay   time.Duration
	}
}

type photoFrame struct {
	conf     Config
	client   *immich.Client
	win      fyne.Window
	assQueue chan *immich.Asset
}

func (pf *photoFrame) run() error {
	albums, err := pf.getConfiguredAlbums()
	if err != nil {
		return err
	}
	if n := countAssets(albums); n == 0 {
		return errors.New("no assets found")
	}
	pf.initWindow()
	go pf.displayWorker()
	go pf.assetWorker(albums)

	pf.win.ShowAndRun()
	return nil
}

func (pf *photoFrame) initWindow() {
	pf.win = app.New().NewWindow("immich")
	pf.win.SetFullScreen(true)
}

func (pf *photoFrame) getConfiguredAlbums() ([]immich.Album, error) {
	// Get all albums.
	allAlbums, err := pf.client.GetAlbums()
	if err != nil {
		return nil, err
	}
	slog.Info("found albums", "count", len(allAlbums))

	// If no albums are configured, use all of the ones we found.
	if len(pf.conf.App.ImmichAlbums) == 0 {
		return allAlbums, nil
	}

	// Build set of configured album names.
	configuredAlbumNames := pf.conf.App.ImmichAlbums
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

func InitApp(conf Config) (*photoFrame, error) {
	client := immich.NewClient(
		immich.WithRemote(conf.Remote),
		immich.WithLocalStorage(conf.LocalStorage),
		immich.WithInMemoryCache(conf.InMemoryCache),
	)
	slog.Info("created immich client")
	slog.Info("client diagnostics", "diagnostics", client.Diagnostics())
	return &photoFrame{
		client:   client,
		conf:     conf,
		assQueue: make(chan *immich.Asset, 10),
	}, nil
}

func countAssets(albums []immich.Album) int {
	n := 0
	for _, album := range albums {
		n += album.AssetCount
	}
	return n
}
