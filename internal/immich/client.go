package immich

import (
	"errors"
	"fmt"
	"log/slog"
	"os"

	"github.com/dustin/go-humanize"

	"immich-photo-frame/internal/immich/api"
)

// Client provides an API for retrieving immich albums and assets with seamless
// in-memory and local storage caching.
type Client struct {
	cache  rwClient
	local  rwClient
	remote remoteClient
}

// rwClient is a client that can both read and write, typically local clients,
// not the remote immich server.
type rwClient interface {
	readClient
	writeClient
}

// readClient is a client that can provide immich albums and assets.
type readClient interface {
	GetAsset(md AssetMetadata) (*Asset, error)
	GetAlbums() (*GetAlbumsResponse, error)
	GetAlbumAssets(id AlbumID) ([]AssetMetadata, error)
}

// writeClient is a client that can store immich albums and assets.
type writeClient interface {
	StoreAsset(asset *Asset) error
	StoreAlbums(resp GetAlbumsResponse) error
	StoreAlbumAssets(id AlbumID, assets []AssetMetadata) error
}

// remoteClient is a read-only client with a connection check.
type remoteClient interface {
	IsConnected() error
	readClient
}

// GetAsset retrieves an immich asset given its metadata. It first checks the
// in-memory cache, then local storage, then the remote server. On success, the
// in-memory cache and (if applicable) the local storage are updated.
func (c Client) GetAsset(md AssetMetadata) (*Asset, error) {
	if ass, err := c.cache.GetAsset(md); err == nil {
		return ass, nil
	}
	if ass, err := c.local.GetAsset(md); err == nil {
		_ = c.cache.StoreAsset(ass)
		return ass, nil
	}
	slog.Debug("fetching asset from remote", "id", md.ID, "name", md.Name)
	ass, err := c.remote.GetAsset(md)
	if err == nil {
		slog.Info("fetched asset from remote",
			"id", md.ID,
			"name", md.Name,
			"size", humanize.Bytes(uint64(len(ass.Data))),
		)
		_ = c.cache.StoreAsset(ass)
		_ = c.local.StoreAsset(ass)
	}
	return ass, err
}

// GetAlbums retrieves all immich albums. It first checks the in-memory cache,
// then local storage, then the remote server. On success, the in-memory cache
// and (if applicable) the local storage are updated.
func (c Client) GetAlbums() ([]Album, error) {
	var errs []error
	if resp, err := c.cache.GetAlbums(); err == nil {
		return resp.Albums, nil
	} else {
		errs = append(errs, err)
	}
	if resp, err := c.local.GetAlbums(); err == nil {
		c.cache.StoreAlbums(*resp)
		return resp.Albums, nil
	} else {
		errs = append(errs, err)
	}
	slog.Info("fetching albums from remote")
	if resp, err := c.remote.GetAlbums(); err == nil {
		c.cache.StoreAlbums(*resp)
		c.local.StoreAlbums(*resp)
		return resp.Albums, nil
	} else {
		errs = append(errs, err)
	}
	return nil, errors.Join(errs...)
}

// GetAlbumAssets gets the asset metadata for the given immich album ID. It
// first checks the in-memory cache, then local storage, then the remote
// server. On success, the in-memory cache and (if-applicable) the local
// storage are updates.
func (c Client) GetAlbumAssets(id AlbumID) ([]AssetMetadata, error) {
	if assets, err := c.cache.GetAlbumAssets(id); err == nil {
		return assets, nil
	}
	if assets, err := c.local.GetAlbumAssets(id); err == nil {
		c.cache.StoreAlbumAssets(id, assets)
		return assets, nil
	}
	slog.Info("fetching album asset metadata from remote", "id", id)
	assets, err := c.remote.GetAlbumAssets(id)
	if err == nil {
		c.cache.StoreAlbumAssets(id, assets)
		c.local.StoreAlbumAssets(id, assets)
	}
	return assets, err
}

// clientOpt is used for configuring the [Client].
type clientOpt func(*Client)

// WithInMemoryCache adds an in-memory cache to the Client, if configured. Only
// one in-memory cache can be configured. If multiple are provided, the last is
// used.
func WithInMemoryCache(conf InMemoryConfig) clientOpt {
	return func(c *Client) {
		if !conf.UseInMemoryCache {
			return
		}
		c.cache = newInMemoryCacheClient(conf)
	}
}

// WithLocalStorage adds a local-storage client, if configured. Only one local
// storage client can be configured. If multiple are provided, the last is
// used.
func WithLocalStorage(conf LocalConfig) clientOpt {
	return func(c *Client) {
		if !conf.UseLocalStorage {
			return
		}
		if err := os.MkdirAll(conf.LocalStoragePath, 0755); err != nil {
			slog.Error("could not create local storage directory", "error", err)
			return
		}
		c.local = newLocalStorageClient(conf)
	}
}

// WithRemote adds a remote client. Only one remote client can be configured.
// If multiple are provided, the last is used.
func WithRemote(conf api.Config) clientOpt {
	return func(c *Client) {
		c.remote = api.NewClient(conf)
	}
}

// NewClient initialized a new client with the provided options. See
// [WithInMemoryCache], [WithLocalStorage], and [WithRemote].
func NewClient(opts ...clientOpt) *Client {
	noop := noopClient{}
	client := &Client{
		cache:  noop,
		local:  noop,
		remote: noop,
	}
	for _, opt := range opts {
		opt(client)
	}
	return client
}

// Helper functions for generating keys.
//
// Used for both in-memory keys and local storage filenames. They don't need to
// match across implementations, but it's simpler if it does.
func assetKey(id AssetID) string { return fmt.Sprintf("asset-%s", id) }
func albumKey(id AlbumID) string { return fmt.Sprintf("album-%s", id) }
func albumsKey() string          { return "albums" }

// noopClient provides a noop implementation for the cache, local, and remote
// clients.
type noopClient struct{}

func (noopClient) GetAlbumAssets(AlbumID) ([]AssetMetadata, error) { return nil, errors.New("noop") }
func (noopClient) GetAlbums() (*GetAlbumsResponse, error)          { return nil, errors.New("noop") }
func (noopClient) GetAsset(AssetMetadata) (*Asset, error)          { return nil, errors.New("noop") }
func (noopClient) IsConnected() error                              { return errors.New("noop") }
func (noopClient) StoreAlbumAssets(AlbumID, []AssetMetadata) error { return errors.New("noop") }
func (noopClient) StoreAlbums(GetAlbumsResponse) error             { return errors.New("noop") }
func (noopClient) StoreAsset(*Asset) error                         { return errors.New("noop") }
