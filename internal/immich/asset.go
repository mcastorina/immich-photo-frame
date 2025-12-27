package immich

import (
	"encoding/json"
	"io"
	"path"

	"fyne.io/fyne/v2"
)

type AssetID string

type AssetMetadata struct {
	ID       AssetID          `json:"id"`
	Type     string           `json:"type"`
	Name     string           `json:"originalFileName"`
	Duration string           `json:"duration"`
	ExifInfo map[string]any   `json:"exifInfo"`
	People   []map[string]any `json:"people"`
}

// Asset implements fyne.Resource.
var _ fyne.Resource = Asset{}

type Asset struct {
	Meta AssetMetadata
	Data []byte
}

func (a Asset) Content() []byte { return a.Data }
func (a Asset) Name() string    { return a.Meta.Name }

// GetAssetPreview gets the metadata associated with an asset.
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
func (c Client) GetAsset(md AssetMetadata) (*Asset, error) {
	resp, err := c.Get(path.Join("/assets", string(md.ID), "original"))
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()
	data, err := io.ReadAll(resp.Body)

	return &Asset{
		Meta: md,
		Data: data,
	}, nil
}

// GetAssetByID retrieves the requested asset along with its metadata.
func (c Client) GetAssetByID(id AssetID) (*Asset, error) {
	md, err := c.GetAssetPreview(id)
	if err != nil {
		return nil, err
	}
	return c.GetAsset(*md)
}
