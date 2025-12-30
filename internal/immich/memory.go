package immich

import (
	"errors"
	"fmt"

	"github.com/dustin/go-humanize"
	lru "github.com/hashicorp/golang-lru/v2"
)

// inMemoryCache is a rwClient for storing and retrieving assets and metadata
// in-memory.
type inMemoryCache struct {
	// TODO: Maybe don't use the LRU for metadata storage.
	*lru.Cache[string, any]
	conf InMemoryConfig
}

// GetAlbumAssets attempts to retrieve the asset metadata for the given album
// from the cache. An error is returned if the data is not available.
func (i inMemoryCache) GetAlbumAssets(id AlbumID) ([]AssetMetadata, error) {
	key := i.albumKey(id)
	val, err := i.get(key)
	if err != nil {
		return nil, err
	}
	assets, ok := val.([]AssetMetadata)
	if !ok {
		return nil, fmt.Errorf("unexpected album asset type: %T", val)
	}
	return assets, nil
}

// StoreAlbumAssets writes the asset metadata for the given album to the cache.
func (i inMemoryCache) StoreAlbumAssets(id AlbumID, assets []AssetMetadata) error {
	key := i.albumKey(id)
	i.Add(key, assets)
	return nil
}

// GetAlbums attempts to retrieve the list of albums from the cache. An error
// is returned if the data is not available.
func (i inMemoryCache) GetAlbums() ([]Album, error) {
	key := i.albumsKey()
	val, err := i.get(key)
	if err != nil {
		return nil, err
	}
	albums, ok := val.([]Album)
	if !ok {
		return nil, fmt.Errorf("unexpected asset type: %T", val)
	}
	return albums, nil
}

// StoreAlbums writes the list of albums to the cache.
func (i inMemoryCache) StoreAlbums(albums []Album) error {
	key := i.albumsKey()
	i.Add(key, albums)
	return nil
}

// GetAsset attempts to retrieve the asset from the cache. An error is returned
// if the data is not available.
func (i inMemoryCache) GetAsset(md AssetMetadata) (*Asset, error) {
	key := i.assetKey(md.ID)
	val, err := i.get(key)
	if err != nil {
		return nil, err
	}
	ass, ok := val.(*Asset)
	if !ok {
		return nil, fmt.Errorf("unexpected asset type: %T", val)
	}
	return ass, nil
}

// StoreAsset writes the asset to the cache.
func (i inMemoryCache) StoreAsset(asset *Asset) error {
	// TODO: Evict when cache size > configured size.
	key := i.assetKey(asset.Meta.ID)
	i.Add(key, asset)
	return nil
}

// get is a helper method to return an error if the key does not exist in the
// cache.
func (i inMemoryCache) get(key string) (any, error) {
	v, ok := i.Get(key)
	if !ok {
		return nil, errors.New("not found")
	}
	return v, nil
}

// Helper methods for generating keys.
func (i inMemoryCache) assetKey(id AssetID) string { return fmt.Sprintf("asset-%s", id) }
func (i inMemoryCache) albumKey(id AlbumID) string { return fmt.Sprintf("album-%s", id) }
func (i inMemoryCache) albumsKey() string          { return "albums" }

// newInMemoryCacheClient initializes an [inMemoryCache] client.
func newInMemoryCacheClient(conf InMemoryConfig) inMemoryCache {
	avgAssetSize, _ := humanize.ParseBytes("3 MB")
	cacheSize := 1
	if configuredSize := uint64(conf.InMemoryCacheSize) / avgAssetSize; configuredSize > 0 {
		cacheSize = int(configuredSize)
	}
	l, _ := lru.New[string, any](cacheSize)
	return inMemoryCache{l, conf}
}
