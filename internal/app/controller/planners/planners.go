package planners

import (
	"fmt"
	"strings"

	"immich-photo-frame/internal/immich"
)

// PlanIter defines how to iterate over the configured albums and assets.
type PlanIter interface {
	Init(source AssetClient, albums []immich.Album)
	Next() *immich.AssetMetadata
}

// AssetClient describes an object that, given an AlbumID, can retrieve a list
// of AssetMetadata.
type AssetClient interface {
	GetAlbumAssets(immich.AlbumID) ([]immich.AssetMetadata, error)
}

// PlanAlgorithm is a concrete object that embeds a PlanIter interface. This
// struct allows us to take advantage of custom TOML-decoding into a PlanIter
// object based on a name. See [PlanAlgorithm.UnmarshalText].
type PlanAlgorithm struct {
	PlanIter
}

// planAlgorithms is the LUT for all the PlanIter objects and their names.
var planAlgorithms = map[string]PlanIter{
	"sequential": &Sequential{},
}

// UnmarshalText implements toml.TextUnmarshaler.
func (p *PlanAlgorithm) UnmarshalText(text []byte) error {
	iter, ok := planAlgorithms[strings.ToLower(string(text))]
	if ok {
		p.PlanIter = iter
		return nil
	}
	var validAlgos []string
	for key := range planAlgorithms {
		validAlgos = append(validAlgos, key)
	}
	return fmt.Errorf(
		"unsupported plan algorithm %q, expected one of %v",
		string(text), validAlgos,
	)
}
