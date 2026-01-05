package controller

import (
	"errors"
	"log/slog"
	"time"

	"immich-photo-frame/internal/app/controller/planners"
	"immich-photo-frame/internal/app/display"
	"immich-photo-frame/internal/immich"
)

var (
	Next cmd = "next"
	Prev cmd = "prev"
)

// cmd is an internal type representing a requested action performed by the
// user.
type cmd string

// Config holds configuration values for controlling the photo-frame behavior.
//
// It is organized to take advantage of TOML parsing, however this package does
// not handle parsing and has no expectation on how it will be initialized.
type Config struct {
	ImmichAlbums  []string
	ImageDelay    time.Duration
	HistorySize   int
	PlanAlgorithm planners.PlanAlgorithm
}

// Controller gathers assets and drives the Display.
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

// New initializes the Controller. An error is returned if it could not find
// any albums or assets to give to the Display.
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

// Next requests that the next asset be shown immediately.
func (c *Controller) Next() {
	c.cmd <- Next
}

// Prev requests that the previous asset be shown immediately.
func (c *Controller) Prev() {
	c.cmd <- Prev
}

// Run drives the Display indefinitely.
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

// currentAsset is a helper method to get the current asset to be displayed.
func (c *Controller) currentAsset() display.DecodedAsset {
	return c.history[c.historyIndex]
}

// nextHistory is a helper method to modify history or historyIndex to advance
// the display.
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

// prevHistory is a helper method to move historyIndex back one, if possible.
func (c *Controller) prevHistory() {
	if c.historyIndex > 0 && c.history[c.historyIndex-1].Img != nil {
		c.historyIndex--
	}
}

// nextAssetFromPlan is a helper method to get the next immich asset from the
// configured plan, download it, and decode it into a DecodedAsset. It retries
// up to 5 times to get an asset and returns an error if it could not.
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

// getConfiguredAlbums is a helper function to convert a list of album names
// into a list of immich Album objects. An error is returned iff there was a
// problem getting all of the albums from the immich Client.
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

// countAssets is a helper function to sum all of the reported asset counts in
// the albums as a sanity check that there will probably be something to
// display.
func countAssets(albums []immich.Album) int {
	n := 0
	for _, album := range albums {
		n += album.AssetCount
	}
	return n
}
