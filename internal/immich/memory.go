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
	key := albumKey(id)
	val, err := i.get(key)
	if err != nil {
		return nil, err
	}
	assets, ok := val.([]AssetMetadata)
	if !ok {
		return nil, fmt.Errorf("unexpected album asset type: %T", val)
	}

	// Make a copy so the cache cannot be modified.
	assetCopy := make([]AssetMetadata, len(assets))
	copy(assetCopy, assets)
	return assetCopy, nil
}

// StoreAlbumAssets writes the asset metadata for the given album to the cache.
func (i inMemoryCache) StoreAlbumAssets(id AlbumID, assets []AssetMetadata) error {
	// Make a copy so the cache cannot be modified.
	assetCopy := make([]AssetMetadata, len(assets))
	copy(assetCopy, assets)

	key := albumKey(id)
	i.Add(key, assetCopy)
	return nil
}

// GetAlbums attempts to retrieve the list of albums from the cache. An error
// is returned if the data is not available.
func (i inMemoryCache) GetAlbums() (*GetAlbumsResponse, error) {
	key := albumsKey()
	val, err := i.get(key)
	if err != nil {
		return nil, err
	}
	resp, ok := val.(GetAlbumsResponse)
	if !ok {
		return nil, fmt.Errorf("unexpected asset type: %T", val)
	}

	// Make a copy so the cache cannot be modified.
	albumCopy := make([]Album, len(resp.Albums))
	copy(albumCopy, resp.Albums)
	resp.Albums = albumCopy
	return &resp, nil
}

// StoreAlbums writes the list of albums to the cache.
func (i inMemoryCache) StoreAlbums(resp GetAlbumsResponse) error {
	// Make a copy so the cache cannot be modified.
	albumCopy := make([]Album, len(resp.Albums))
	copy(albumCopy, resp.Albums)
	resp.Albums = albumCopy

	key := albumsKey()
	i.Add(key, resp)
	return nil
}

// GetAsset attempts to retrieve the asset from the cache. An error is returned
// if the data is not available.
func (i inMemoryCache) GetAsset(md AssetMetadata) (*Asset, error) {
	key := assetKey(md.ID)
	val, err := i.get(key)
	if err != nil {
		return nil, err
	}
	ass, ok := val.(*Asset)
	if !ok {
		return nil, fmt.Errorf("unexpected asset type: %T", val)
	}
	// Let's assume callers will be responsible and not modify ass.Data
	return ass, nil
}

// StoreAsset writes the asset to the cache.
func (i inMemoryCache) StoreAsset(ass *Asset) error {
	// Let's assume callers will be responsible and not modify ass.Data
	// TODO: Evict when cache size > configured size.
	key := assetKey(ass.Meta.ID)
	i.Add(key, ass)
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

// newInMemoryCacheClient initializes an [inMemoryCache] client.
func newInMemoryCacheClient(conf InMemoryConfig) inMemoryCache {
	avgAssetSize, _ := humanize.ParseBytes("350 kB")
	cacheSize := 1
	if configuredSize := uint64(conf.InMemoryCacheSize) / avgAssetSize; configuredSize > 0 {
		cacheSize = int(configuredSize)
	}
	l, _ := lru.New[string, any](cacheSize)
	return inMemoryCache{l, conf}
}
