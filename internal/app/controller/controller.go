package controller

import (
	"errors"
	"immich-photo-frame/internal/app/controller/planners"
	"immich-photo-frame/internal/app/display"
	"immich-photo-frame/internal/immich"
	"log/slog"
	"time"
)

var (
	Next cmd = "next"
	Prev cmd = "prev"
)

type cmd string

type Config struct {
	ImmichAlbums  []string
	ImageDelay    time.Duration
	HistorySize   int
	PlanAlgorithm planners.PlanAlgorithm
}

type Controller struct {
	conf             Config
	configuredAlbums []immich.Album
	// TODO: Change display.Display to an interface.
	disp *display.Display
	// TODO: Change immich.Client to an interface.
	client *immich.Client
	cmd    chan cmd
	// TODO: Should we only store asset metadata since we can get the DecodedAsset from that?
	history      []display.DecodedAsset
	historyIndex int
}

func New(conf Config, client *immich.Client, disp *display.Display) (*Controller, error) {
	albums, err := getConfiguredAlbums(client, conf.ImmichAlbums)
	if err != nil {
		return nil, err
	}
	if n := countAssets(albums); n == 0 {
		return nil, errors.New("no assets found")
	}
	return &Controller{
		conf:             conf,
		configuredAlbums: albums,
		disp:             disp,
		client:           client,
		cmd:              make(chan cmd, 10),
		history:          make([]display.DecodedAsset, conf.HistorySize+1),
		historyIndex:     conf.HistorySize,
	}, nil
}

func (c *Controller) Next() {
	c.cmd <- Next
}

func (c *Controller) Prev() {
	c.cmd <- Prev
}

func (c *Controller) Run() {
	// Initialize planner.
	c.conf.PlanAlgorithm.Init(c.client, c.configuredAlbums)
	// Initialize display by getting the first asset and showing it.
	c.nextHistory()
	c.disp.Show(c.currentAsset())

	ticker := time.NewTicker(c.conf.ImageDelay)
	for {
		select {
		case <-ticker.C:
			c.nextHistory()
		case cmd := <-c.cmd:
			ticker.Reset(c.conf.ImageDelay)
			switch cmd {
			case Next:
				c.nextHistory()
			case Prev:
				c.prevHistory()
			}
		}
		c.disp.Show(c.currentAsset())
	}
}

func (c *Controller) currentAsset() display.DecodedAsset {
	return c.history[c.historyIndex]
}

func (c *Controller) nextHistory() {
	if c.historyIndex < len(c.history)-1 {
		c.historyIndex++
		return
	}
	da, err := c.nextAssetFromPlan()
	if err != nil {
		slog.Error("failed to advance", "error", err)
		return
	}
	c.history = append(c.history, *da)
	c.history = c.history[1:]
}

func (c *Controller) prevHistory() {
	if c.historyIndex > 0 && c.history[c.historyIndex-1].Img != nil {
		c.historyIndex--
	}
}

func (c *Controller) nextAssetFromPlan() (*display.DecodedAsset, error) {
	for range 5 {
		md := c.conf.PlanAlgorithm.Next()
		if md == nil {
			slog.Error("failed to get next asset metadata from planner")
			continue
		}
		ass, err := c.client.GetAsset(*md)
		if err != nil {
			slog.Error("failed to get asset", "name", md.Name, "id", md.ID)
			continue
		}
		da, err := c.disp.DecodeAsset(ass)
		if err != nil {
			slog.Error("failed to decode asset", "name", md.Name, "id", md.ID)
			continue
		}
		return da, nil
	}
	return nil, errors.New("could not get the next asset after 5 tries")
}

func getConfiguredAlbums(client *immich.Client, albums []string) ([]immich.Album, error) {
	// Get all albums.
	allAlbums, err := client.GetAlbums()
	if err != nil {
		return nil, err
	}
	slog.Info("found albums", "count", len(allAlbums))

	// If no albums are configured, use all of the ones we found.
	if len(albums) == 0 {
		return allAlbums, nil
	}

	// Build set of configured album names.
	albumNameSet := make(map[string]struct{})
	for _, album := range albums {
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

func countAssets(albums []immich.Album) int {
	n := 0
	for _, album := range albums {
		n += album.AssetCount
	}
	return n
}
