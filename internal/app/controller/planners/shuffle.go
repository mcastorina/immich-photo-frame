package planners

import (
	"log/slog"
	"math/rand/v2"

	"immich-photo-frame/internal/immich"
)

// Shuffle implements PlanIter by shuffling the albums and their assets. It
// works by saving all of the asset metadata in memory and shuffling, so this
// may not be a good approach for a large amount of assets.
type Shuffle struct {
	assets     []immich.AssetMetadata
	assetIndex int
}

func (s *Shuffle) Name() string { return "shuffle" }

// Init implements PlanIter and initializes the Shuffle object.
func (s *Shuffle) Init(source AssetClient, albums []immich.Album) {
	// Get all assets from all albums.
	var assets []immich.AssetMetadata
	for _, album := range albums {
		ass, err := source.GetAlbumAssets(album.ID)
		if err != nil {
			slog.Error("failed to get album assets to shuffle",
				"id", album.ID,
				"name", album.Name,
				"error", err,
			)
			continue
		}
		assets = append(assets, ass...)
	}
	*s = Shuffle{assets: assets}
	s.shuffle()
}

// Next implements PlanIter and retrieves the next AssetMetadata.
func (s *Shuffle) Next() *immich.AssetMetadata {
	if len(s.assets) == 0 {
		return nil
	}
	if s.assetIndex >= len(s.assets) {
		s.assetIndex = 0
		s.shuffle()
	}
	md := s.assets[s.assetIndex]
	s.assetIndex++
	return &md
}

// shuffle is a helper method to shuffle the contents of the assets slice.
func (s *Shuffle) shuffle() {
	rand.Shuffle(len(s.assets), func(i, j int) {
		s.assets[i], s.assets[j] = s.assets[j], s.assets[i]
	})
}
