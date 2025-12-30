package immich

import (
	"github.com/dustin/go-humanize"

	"immich-photo-frame/internal/immich/api"
)

// Config holds configuration values for caching behavior.
//
// It is organized to take advantage of TOML parsing, however this package does
// not handle parsing and has no expectation on how it will be initialized.
type Config struct {
	// Local storage to persist data across restarts. Assets will be
	// fetched from local storage first before reaching out to the immich
	// server.
	LocalStorage LocalConfig

	// In memory cache for assets, either loaded from persistent storage or
	// the immich server.
	InMemoryCache InMemoryConfig

	// Remote configuration for connecting to the immich API.
	Remote api.Config
}

// Local storage to persist data across restarts. Assets will be fetched from
// local storage first before reaching out to the immich server.
type LocalConfig struct {
	UseLocalStorage  bool
	LocalStorageSize HumanBytes
	LocalStoragePath string
}

// In memory cache for assets, either loaded from persistent storage or the
// immich server.
type InMemoryConfig struct {
	UseInMemoryCache  bool
	InMemoryCacheSize HumanBytes
}

// HumanBytes is a custom type to decode human-readable byte values into an
// integer.
type HumanBytes uint64

// UnmarshalText implements toml.TextUnmarshaler.
func (h *HumanBytes) UnmarshalText(text []byte) error {
	nbytes, err := humanize.ParseBytes(string(text))
	*h = HumanBytes(nbytes)
	return err
}

// String converts the integer back into a human-readable representation.
func (h *HumanBytes) String() string {
	if h == nil {
		return ""
	}
	return humanize.Bytes(uint64(*h))
}
