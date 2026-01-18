package api

import (
	"encoding/json"
	"path"
	"time"
)

// AlbumID is the immich ID for an album, usually in the shape of UUIDv4.
type AlbumID string

// Album contains relevant album information retrieved from the immich API.
//
// See: https://api.immich.app/models/AlbumResponseDto
type Album struct {
	Name        string  `json:"albumName"`
	Description string  `json:"description"`
	ID          AlbumID `json:"id"`
	Order       string  `json:"order"`
	AssetCount  int     `json:"assetCount"`
}

// GetAlbumsResponse wraps the immich API response with some metadata.
type GetAlbumsResponse struct {
	ResponseTime time.Time
	Albums       []Album
}

// GetAlbums retrieves all albums from the immich API.
//
// See: https://api.immich.app/endpoints/albums/getAllAlbums
func (c Client) GetAlbums() (*GetAlbumsResponse, error) {
	resp, err := c.Get("/albums")
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	var albums []Album
	if err := json.NewDecoder(resp.Body).Decode(&albums); err != nil {
		return nil, err
	}
	return &GetAlbumsResponse{
		ResponseTime: time.Now(),
		Albums:       albums,
	}, nil
}

// GetAlbumsAssetsResponse wraps the immich API response with some metadata.
type GetAlbumsAssetsResponse struct {
	ResponseTime   time.Time
	AssetMetadatas []AssetMetadata
}

// GetAlbumAssets retrieves the album asset metadata for the provided album ID.
//
// See: https://api.immich.app/endpoints/albums/getAlbumInfo
func (c Client) GetAlbumAssets(id AlbumID) (*GetAlbumsAssetsResponse, error) {
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
	return &GetAlbumsAssetsResponse{
		ResponseTime:   time.Now(),
		AssetMetadatas: ar.Assets,
	}, nil
}
