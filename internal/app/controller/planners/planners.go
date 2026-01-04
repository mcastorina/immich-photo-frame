package planners

import (
	"fmt"
	"strings"

	"immich-photo-frame/internal/immich"
)

type AssetClient interface {
	GetAlbumAssets(immich.AlbumID) ([]immich.AssetMetadata, error)
}

type PlanIter interface {
	Init(source AssetClient, albums []immich.Album)
	Next() *immich.AssetMetadata
}

type PlanAlgorithm struct {
	PlanIter
}

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
