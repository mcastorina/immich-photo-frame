package cache

import (
	"fmt"

	"github.com/dustin/go-humanize"
	lru "github.com/hashicorp/golang-lru/v2"

	"immich-photo-frame/internal/immich"
)

type inMemoryCache struct {
	*lru.Cache[string, any]
}

// GetAsset implements rwClient.
func (i inMemoryCache) GetAsset(md immich.AssetMetadata) (*immich.Asset, error) {
	key := i.assetKey(md.ID)
	val, ok := i.Get(key)
	if !ok {
		// Not having it in the cache is not an error.
		return nil, nil
	}
	ass, ok := val.(*immich.Asset)
	if !ok {
		// Not being the correct type _is_ an error.
		return nil, fmt.Errorf("unexpected asset type: %T", val)
	}
	return ass, nil
}

// IsConnected implements rwClient.
func (i inMemoryCache) IsConnected() error {
	return nil
}

// StoreAsset implements rwClient.
func (i inMemoryCache) StoreAsset(asset *immich.Asset) error {
	key := i.assetKey(asset.Meta.ID)
	i.Add(key, asset)
	return nil
}

func (i inMemoryCache) assetKey(id immich.AssetID) string { return fmt.Sprintf("asset-%s", id) }

func newInMemoryCacheClient(conf InMemoryConfig) inMemoryCache {
	avgAssetSize, _ := humanize.ParseBytes("3 MB")
	cacheSize := 1
	if configuredSize := uint64(conf.InMemoryCacheSize) / avgAssetSize; configuredSize > 0 {
		cacheSize = int(configuredSize)
	}
	l, _ := lru.New[string, any](cacheSize)
	return inMemoryCache{l}
}
