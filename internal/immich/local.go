package immich

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// localStorageClient is a rwClient for storing and retrieving assets and
// metadata from local persistent storage.
type localStorageClient struct {
	conf LocalConfig
}

// GetAlbumAssets attempts to retrieve the asset metadata for the given album
// from the filesystem. An error is returned if the data is not available.
func (l localStorageClient) GetAlbumAssets(id AlbumID) ([]AssetMetadata, error) {
	key := albumKey(id)
	data, err := l.get(key)
	if err != nil {
		return nil, err
	}
	var assets []AssetMetadata
	if err := json.Unmarshal(data, &assets); err != nil {
		return nil, err
	}
	return assets, nil
}

// StoreAlbumAssets attempts to write the asset metadata for the given album to
// the filesystem.
func (l localStorageClient) StoreAlbumAssets(id AlbumID, assets []AssetMetadata) error {
	key := albumKey(id)
	data, err := json.Marshal(assets)
	if err != nil {
		return err
	}
	return l.store(key, data)
}

// GetAlbums attempts to retrieve the list of albums from the filesystem. An
// error is returned if the data is not available.
func (l localStorageClient) GetAlbums() (*GetAlbumsResponse, error) {
	key := albumsKey()
	data, err := l.get(key)
	if err != nil {
		return nil, err
	}
	var resp GetAlbumsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// StoreAlbums attempts to write the list of albums to the filesystem.
func (l localStorageClient) StoreAlbums(resp GetAlbumsResponse) error {
	key := albumsKey()
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return l.store(key, data)
}

// GetAsset attempts to retrieve the asset from the filesystem. An error is
// returned if the data is not available.
func (l localStorageClient) GetAsset(md AssetMetadata) (*Asset, error) {
	key := assetKey(md.ID)
	data, err := l.get(key)
	if err != nil {
		return nil, err
	}
	return &Asset{
		Meta: md,
		Data: data,
	}, nil
}

// StoreAsset attempts to write the asset to the filesystem.
func (l localStorageClient) StoreAsset(asset *Asset) error {
	key := assetKey(asset.Meta.ID)
	return l.store(key, asset.Data)
}

// get is a helper method to convert the key to a filepath and read the
// contents of the file.
func (l localStorageClient) get(key string) ([]byte, error) {
	return os.ReadFile(filepath.Join(l.conf.LocalStoragePath, key))
}

// store is a helper method to convert the key to a filepath and write the data
// to disk.
func (l localStorageClient) store(key string, data []byte) error {
	// TODO: Manage amount of disk space used.
	return os.WriteFile(filepath.Join(l.conf.LocalStoragePath, key), data, 0644)
}

// newInMemoryCacheClient initializes a [localStorageClient] client.
func newLocalStorageClient(conf LocalConfig) localStorageClient {
	conf.LocalStoragePath = filepath.Clean(conf.LocalStoragePath)
	return localStorageClient{conf}
}
