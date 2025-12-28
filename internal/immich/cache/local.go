package cache

import (
	"errors"

	"immich-photo-frame/internal/immich"
)

type localStorageClient struct{}

// GetAsset implements rwClient.
func (l localStorageClient) GetAsset(md immich.AssetMetadata) (*immich.Asset, error) {
	return nil, errors.New("unimplemented")
}

// IsConnected implements rwClient.
func (l localStorageClient) IsConnected() error {
	return errors.New("unimplemented")
}

// StoreAsset implements rwClient.
func (l localStorageClient) StoreAsset(asset *immich.Asset) error {
	return errors.New("unimplemented")
}

func newLocalStorageClient(conf LocalConfig) localStorageClient {
	return localStorageClient{}
}
