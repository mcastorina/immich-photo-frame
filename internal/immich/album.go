package immich

import (
	"encoding/json"
	"errors"
	"path"
)

type AlbumID string

type Album struct {
	Name        string  `json:"albumName"`
	Description string  `json:"description"`
	ID          AlbumID `json:"id"`
	AssetCount  int     `json:"assetCount"`
}

func (c Client) GetAlbums() ([]Album, error) {
	resp, err := c.Get("/albums")
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	var albums []Album
	if err := json.NewDecoder(resp.Body).Decode(&albums); err != nil {
		return nil, err
	}
	return albums, nil
}

func (c Client) GetAlbumByName(name string) (*Album, error) {
	albums, err := c.GetAlbums()
	if err != nil {
		return nil, err
	}
	for _, album := range albums {
		if album.Name == name {
			return &album, nil
		}
	}
	return nil, errors.New("album not found")
}

func (c Client) GetAlbumAssets(id AlbumID) ([]AssetMetadata, error) {
	resp, err := c.Get(path.Join("/albums", string(id)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	type albumResp struct {
		Assets []AssetMetadata `json:"assets"`
	}
	var ar albumResp
	if err := json.NewDecoder(resp.Body).Decode(&ar); err != nil {
		return nil, err
	}
	return ar.Assets, nil
}
