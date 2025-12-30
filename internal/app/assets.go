package app

import (
	"immich-photo-frame/internal/immich"
	"iter"
	"log/slog"
)

type albumAssetGetter interface {
	GetAlbumAssets(id immich.AlbumID) ([]immich.AssetMetadata, error)
}

type assetMetadataIter struct {
	source           albumAssetGetter
	albums           []immich.Album
	albumsIndex      int
	albumAssets      []immich.AssetMetadata
	albumAssetsIndex int
}

func NewAssetMetadataIter(source albumAssetGetter, albums []immich.Album) *assetMetadataIter {
	return &assetMetadataIter{
		source: source,
		albums: albums,
		// Initialize to -1 so the first call to Next() moves to the
		// first album.
		albumsIndex: -1,
	}
}

func (a *assetMetadataIter) Next() *immich.AssetMetadata {
	if len(a.albums) == 0 {
		// Nothing to do.
		return nil
	}
	for attempts := 0; attempts < len(a.albums); attempts++ {
		if asset, ok := a.nextAsset(); ok {
			return asset
		}
		// No more assets, go to the next album.
		a.nextAlbum()
	}
	return nil
}

func (a *assetMetadataIter) nextAsset() (*immich.AssetMetadata, bool) {
	if a.albumAssetsIndex >= len(a.albumAssets) {
		return nil, false
	}
	ass := a.albumAssets[a.albumAssetsIndex]
	a.albumAssetsIndex++
	return &ass, true
}

func (a *assetMetadataIter) nextAlbum() {
	a.albumsIndex = (a.albumsIndex + 1) % len(a.albums)
	album := a.albums[a.albumsIndex]
	a.albumAssets = nil
	a.albumAssetsIndex = 0
	for attempts := 0; attempts < 3; attempts++ {
		assets, err := a.source.GetAlbumAssets(album.ID)
		if err != nil {
			slog.Error("failed to load album", "album", album, "error", err)
			continue
		}
		a.albumAssets = assets
		break
	}
}

func AssetMetadataIter(source albumAssetGetter, albums []immich.Album) iter.Seq[*immich.AssetMetadata] {
	a := NewAssetMetadataIter(source, albums)
	return func(yield func(*immich.AssetMetadata) bool) {
		for {
			asset := a.Next()
			if asset == nil || !yield(asset) {
				return
			}
		}
	}
}
