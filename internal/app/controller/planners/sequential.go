package planners

import (
	"log/slog"
	"slices"

	"immich-photo-frame/internal/immich"
)

// Sequential implements PlanIter by sequentially iterating the albums by how
// they were configured and their assets by how the album is configured in
// immich ("asc" or "desc").
type Sequential struct {
	source     AssetClient
	albums     []immich.Album
	albumIndex int
	assets     []immich.AssetMetadata
	assetIndex int
}

func (s *Sequential) Name() string { return "sequential" }

// Init implements PlanIter and initializes the Sequential object.
func (s *Sequential) Init(source AssetClient, albums []immich.Album) {
	*s = Sequential{
		source: source,
		albums: albums,
		// Initialize to -1 so the first call to Next() loads the first
		// album.
		albumIndex: -1,
	}
}

// Next implements PlanIter and retrieves the next AssetMetadata.
func (s *Sequential) Next() *immich.AssetMetadata {
	if len(s.albums) == 0 {
		return nil
	}
	// There are still assets left to show.
	if s.assetIndex < len(s.assets) {
		md := s.assets[s.assetIndex]
		s.assetIndex++
		return &md
	}

	// Clear assets cache.
	s.assets = nil
	s.assetIndex = 0

	// Iterate through the albums until we find one with assets.
	for rounds := 0; rounds < len(s.albums); rounds++ {
		// Go to next album.
		s.albumIndex = (s.albumIndex + 1) % len(s.albums)
		// Get the assets.
		assets, err := s.getAlbumAssetsInOrder(s.albums[s.albumIndex])
		if err != nil {
			slog.Error("failed to load assets", "error", err)
			continue
		}
		s.assets = assets
		break
	}

	// Return the first asset, if any.
	if len(s.assets) == 0 {
		return nil
	}
	md := s.assets[s.assetIndex]
	s.assetIndex++
	return &md
}

// getAlbumAssetsInOrder is a helper method to get the album asset metadata in
// the order it is configured in immich (e.g. "asc" or "desc").
func (s *Sequential) getAlbumAssetsInOrder(album immich.Album) ([]immich.AssetMetadata, error) {
	mds, err := s.source.GetAlbumAssets(album.ID)
	if err != nil {
		return nil, err
	}
	if album.Order == "asc" {
		slices.Reverse(mds)
	}
	return mds, nil
}
