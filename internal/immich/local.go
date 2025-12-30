package immich

import "errors"

// localStorageClient is a rwClient for storing and retrieving assets and
// metadata from local persistent storage.
type localStorageClient struct{}

func (l localStorageClient) GetAlbumAssets(id AlbumID) ([]AssetMetadata, error) {
	return nil, errors.New("unimplemented")
}

func (l localStorageClient) StoreAlbumAssets(id AlbumID, assets []AssetMetadata) error {
	return errors.New("unimplemented")
}

func (l localStorageClient) GetAlbums() ([]Album, error) {
	return nil, errors.New("unimplemented")
}

func (l localStorageClient) StoreAlbums([]Album) error {
	return errors.New("unimplemented")
}

func (l localStorageClient) GetAsset(md AssetMetadata) (*Asset, error) {
	return nil, errors.New("unimplemented")
}

func (l localStorageClient) IsConnected() error {
	return errors.New("unimplemented")
}

func (l localStorageClient) StoreAsset(asset *Asset) error {
	return errors.New("unimplemented")
}

func newLocalStorageClient(conf LocalConfig) localStorageClient {
	return localStorageClient{}
}
