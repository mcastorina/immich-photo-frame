package api

import (
	"encoding/json"
	"io"
	"path"

	"fyne.io/fyne/v2"
)

// AssetID is the immich ID for an asset, usually in the shape of UUIDv4.
type AssetID string

// AssetMetadata contains relevant asset information retrieved from the immich API.
//
// See: https://api.immich.app/endpoints/assets/getAssetInfo
type AssetMetadata struct {
	ID       AssetID          `json:"id"`
	Type     string           `json:"type"`
	Name     string           `json:"originalFileName"`
	Duration string           `json:"duration"`
	ExifInfo ExifInfo         `json:"exifInfo"`
	People   []map[string]any `json:"people"`
}

// ExifInfo contains relevant EXIF data associated with an asset.
//
// See: https://api.immich.app/models/ExifResponseDto
type ExifInfo struct {
	City             string  `json:"city"`
	State            string  `json:"state"`
	Country          string  `json:"country"`
	DateTimeOriginal string  `json:"dateTimeOriginal"`
	TimeZone         string  `json:"timeZone"`
	Latitude         float32 `json:"latitude"`
	Longitude        float32 `json:"longitude"`
}

// Asset implements fyne.Resource for displaying the asset.
var _ fyne.Resource = Asset{}

// Asset combines AssetMetadata with the actual asset data.
type Asset struct {
	Meta AssetMetadata
	Data []byte
}

func (a Asset) Content() []byte { return a.Data }
func (a Asset) Name() string    { return a.Meta.Name }

// GetAssetPreview gets the metadata associated with an asset.
//
// See: https://api.immich.app/endpoints/assets/getAssetInfo
func (c Client) GetAssetPreview(id AssetID) (*AssetMetadata, error) {
	// Get asset metadata.
	resp, err := c.Get(path.Join("/assets", string(id)))
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	var md AssetMetadata
	if err := json.NewDecoder(resp.Body).Decode(&md); err != nil {
		return nil, err
	}
	return &md, nil
}

// GetAsset gets the asset associated with the metadata.
//
// See: https://api.immich.app/endpoints/assets/viewAsset
func (c Client) GetAsset(md AssetMetadata) (*Asset, error) {
	p := path.Join("/assets", string(md.ID), "thumbnail")
	resp, err := c.Get(p + "?size=preview")
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	return &Asset{
		Meta: md,
		Data: data,
	}, nil
}

// GetAssetByID retrieves the requested asset along with its metadata. This
// method is a convenience method for calling [GetAssetPreview] and [GetAsset].
func (c Client) GetAssetByID(id AssetID) (*Asset, error) {
	md, err := c.GetAssetPreview(id)
	if err != nil {
		return nil, err
	}
	return c.GetAsset(*md)
}
