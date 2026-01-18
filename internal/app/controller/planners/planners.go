package planners

import (
	"fmt"
	"strings"

	"immich-photo-frame/internal/immich"
)

// PlanIter defines how to iterate over the configured albums and assets.
type PlanIter interface {
	Name() string
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

var (
	// planAlgorithms is the list of all the PlanIter objects that can be
	// configured.
	planAlgorithms = []PlanIter{
		new(Sequential),
		new(Shuffle),
	}

	// planAlgorithmsByName is a LUT of name to PlanIter, built via [init].
	planAlgorithmsByName = map[string]PlanIter{}
)

// UnmarshalText implements toml.TextUnmarshaler.
func (p *PlanAlgorithm) UnmarshalText(text []byte) error {
	iter, ok := planAlgorithmsByName[strings.ToLower(string(text))]
	if ok {
		p.PlanIter = iter
		return nil
	}
	var validAlgos []string
	for key := range planAlgorithmsByName {
		validAlgos = append(validAlgos, key)
	}
	return fmt.Errorf(
		"unsupported plan algorithm %q, expected one of %v",
		string(text), validAlgos,
	)
}

// MarshalJSON implements json.Marshaler.
func (p *PlanAlgorithm) MarshalJSON() ([]byte, error) {
	return fmt.Appendf(nil, `%q`, p.PlanIter.Name()), nil
}

func init() {
	for _, algo := range planAlgorithms {
		planAlgorithmsByName[algo.Name()] = algo
	}
}
