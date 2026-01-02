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

func (l localStorageClient) StoreAlbumAssets(id AlbumID, assets []AssetMetadata) error {
	key := albumKey(id)
	data, err := json.Marshal(assets)
	if err != nil {
		return err
	}
	return l.store(key, data)
}

func (l localStorageClient) GetAlbums() ([]Album, error) {
	key := albumsKey()
	data, err := l.get(key)
	if err != nil {
		return nil, err
	}
	var albums []Album
	if err := json.Unmarshal(data, &albums); err != nil {
		return nil, err
	}
	return albums, nil
}

func (l localStorageClient) StoreAlbums(albums []Album) error {
	key := albumsKey()
	data, err := json.Marshal(albums)
	if err != nil {
		return err
	}
	return l.store(key, data)
}

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

func (l localStorageClient) StoreAsset(asset *Asset) error {
	key := assetKey(asset.Meta.ID)
	return l.store(key, asset.Data)
}

func (l localStorageClient) get(key string) ([]byte, error) {
	return os.ReadFile(filepath.Join(l.conf.LocalStoragePath, key))
}

func (l localStorageClient) store(key string, data []byte) error {
	// TODO: Manage amount of disk space used.
	return os.WriteFile(filepath.Join(l.conf.LocalStoragePath, key), data, 0644)
}

func newLocalStorageClient(conf LocalConfig) localStorageClient {
	return localStorageClient{conf}
}
