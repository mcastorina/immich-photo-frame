package immich

import (
	"errors"
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/dustin/go-humanize"

	"immich-photo-frame/internal/immich/api"
)

// Client provides an API for retrieving immich albums and assets with seamless
// in-memory and local storage caching.
type Client struct {
	refreshInterval time.Duration
	cache           rwClient
	local           rwClient
	remote          remoteClient
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
	GetAlbumAssets(id AlbumID) (*GetAlbumAssetsResponse, error)
}

// writeClient is a client that can store immich albums and assets.
type writeClient interface {
	StoreAsset(asset *Asset) error
	StoreAlbums(resp GetAlbumsResponse) error
	StoreAlbumAssets(id AlbumID, resp GetAlbumAssetsResponse) error
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
	log := slog.With("id", md.ID, "name", md.Name)
	{
		ass, err := c.cache.GetAsset(md)
		if err == nil {
			log.Debug("found asset in cache", "size", humanize.Bytes(uint64(len(ass.Data))))
			return ass, nil
		}
		log.Debug("failed to get asset from cache", "error", err)
	}
	{
		ass, err := c.local.GetAsset(md)
		if err == nil {
			log.Debug("found asset in local storage", "size", humanize.Bytes(uint64(len(ass.Data))))
			if err := c.cache.StoreAsset(ass); err != nil {
				log.Debug("failed to store asset in cache", "error", err)
			}
			return ass, nil
		}
		log.Debug("failed to get asset from local storage", "error", err)
	}
	{
		log.Debug("fetching asset from remote")
		ass, err := c.remote.GetAsset(md)
		if err == nil {
			log.Info("fetched asset from remote", "size", humanize.Bytes(uint64(len(ass.Data))))
			if err := c.cache.StoreAsset(ass); err != nil {
				log.Debug("failed to store asset in cache", "error", err)
			}
			if err := c.local.StoreAsset(ass); err != nil {
				log.Debug("failed to store asset in local storage", "error", err)
			}
			return ass, nil
		}
		log.Debug("failed to get asset from remote", "error", err)
	}
	return nil, errors.New("could not get asset")
}

// GetAlbums retrieves all immich albums. It first checks the in-memory cache,
// then local storage, then the remote server. On success, the in-memory cache
// and (if applicable) the local storage are updated.
func (c Client) GetAlbums() ([]Album, error) {
	var foundResp *GetAlbumsResponse
	{
		resp, err := c.cache.GetAlbums()
		if err == nil && !c.shouldRefresh(resp.ResponseTime) {
			slog.Debug("found albums in cache",
				"age", time.Since(resp.ResponseTime).String(),
				"maxAge", c.refreshInterval.String())
			return resp.Albums, nil
		} else if err == nil {
			slog.Debug("found stale albums in cache",
				"age", time.Since(resp.ResponseTime).String(),
				"maxAge", c.refreshInterval.String())
			foundResp = resp
		} else {
			slog.Debug("failed to get albums from cache", "error", err)
		}
	}
	{
		resp, err := c.local.GetAlbums()
		if err == nil && !c.shouldRefresh(resp.ResponseTime) {
			slog.Debug("found albums in local storage",
				"age", time.Since(resp.ResponseTime).String(),
				"maxAge", c.refreshInterval.String())
			if err := c.cache.StoreAlbums(*resp); err != nil {
				slog.Debug("failed to store albums in cache", "error", err)
			}
			return resp.Albums, nil
		} else if err == nil {
			slog.Debug("found stale albums in local storage",
				"age", time.Since(resp.ResponseTime).String(),
				"maxAge", c.refreshInterval.String())
			foundResp = resp
		} else {
			slog.Debug("failed to get albums from local storage", "error", err)
		}
	}
	{
		slog.Info("fetching albums from remote")
		resp, err := c.remote.GetAlbums()
		if err == nil {
			slog.Debug("fetched albums from remote")
			if err := c.cache.StoreAlbums(*resp); err != nil {
				slog.Debug("failed to store albums in cache", "error", err)
			}
			if err := c.local.StoreAlbums(*resp); err != nil {
				slog.Debug("failed to store albums in local storage", "error", err)
			}
			return resp.Albums, nil
		}
		slog.Debug("failed to get albums from remote", "error", err)
	}
	if foundResp != nil {
		slog.Debug("failed to get albums, using stale response",
			"age", time.Since(foundResp.ResponseTime).String(),
			"maxAge", c.refreshInterval.String())
		return foundResp.Albums, nil
	}
	return nil, errors.New("could not get albums")
}

// GetAlbumAssets gets the asset metadata for the given immich album ID. It
// first checks the in-memory cache, then local storage, then the remote
// server. On success, the in-memory cache and (if-applicable) the local
// storage are updates.
func (c Client) GetAlbumAssets(id AlbumID) ([]AssetMetadata, error) {
	log := slog.With("id", id)
	var foundResp *GetAlbumAssetsResponse
	{
		resp, err := c.cache.GetAlbumAssets(id)
		if err == nil && !c.shouldRefresh(resp.ResponseTime) {
			log.Debug("found album asset metadata in cache",
				"age", time.Since(resp.ResponseTime).String(),
				"maxAge", c.refreshInterval.String())
			return resp.AssetMetadatas, nil
		} else if err == nil {
			log.Debug("found stale album asset metadata in cache",
				"age", time.Since(resp.ResponseTime).String(),
				"maxAge", c.refreshInterval.String())
			foundResp = resp
		} else {
			log.Debug("failed to get album asset metadata from cache", "error", err)
		}
	}
	{
		resp, err := c.local.GetAlbumAssets(id)
		if err == nil && !c.shouldRefresh(resp.ResponseTime) {
			log.Debug("found album asset metadata in local storage",
				"age", time.Since(resp.ResponseTime).String(),
				"maxAge", c.refreshInterval.String())
			if err := c.cache.StoreAlbumAssets(id, *resp); err != nil {
				log.Debug("failed to store album asset metadata in cache", "error", err)
			}
			return resp.AssetMetadatas, nil
		} else if err == nil {
			log.Debug("found stale album asset metadata in local storage",
				"age", time.Since(resp.ResponseTime).String(),
				"maxAge", c.refreshInterval.String())
			foundResp = resp
		} else {
			log.Debug("failed to get album asset metadata from local storage", "error", err)
		}
	}
	{
		log.Info("fetching album asset metadata from remote")
		resp, err := c.remote.GetAlbumAssets(id)
		if err == nil {
			log.Debug("fetched album asset metadata from remote")
			if err := c.cache.StoreAlbumAssets(id, *resp); err != nil {
				log.Debug("failed to store album asset metadata in cache", "error", err)
			}
			if err := c.local.StoreAlbumAssets(id, *resp); err != nil {
				log.Debug("failed to store album asset metadata in local storage", "error", err)
			}
			return resp.AssetMetadatas, nil
		}
		log.Debug("failed to get album asset metadata from remote", "error", err)
	}
	if foundResp != nil {
		log.Debug("failed to get album asset metadata, using stale response",
			"age", time.Since(foundResp.ResponseTime).String(),
			"maxAge", c.refreshInterval.String())
		return foundResp.AssetMetadatas, nil
	}
	return nil, errors.New("could not get album asset metadata")
}

func (c Client) shouldRefresh(respTime time.Time) bool {
	if c.refreshInterval == 0 {
		return false
	}
	return time.Since(respTime) >= c.refreshInterval
}

// clientOpt is used for configuring the [Client].
type clientOpt func(*Client)

// WithRefreshInterval controls how long an immich server response is valid. A
// value of 0 means never refresh the responses. Stale values will continue to
// be used if the client is unable to reach the server.
func WithRefreshInterval(d time.Duration) clientOpt {
	return func(c *Client) { c.refreshInterval = d }
}

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

func (noopClient) GetAlbumAssets(AlbumID) (*GetAlbumAssetsResponse, error) {
	return nil, errors.New("noop")
}
func (noopClient) GetAlbums() (*GetAlbumsResponse, error)                 { return nil, errors.New("noop") }
func (noopClient) GetAsset(AssetMetadata) (*Asset, error)                 { return nil, errors.New("noop") }
func (noopClient) IsConnected() error                                     { return errors.New("noop") }
func (noopClient) StoreAlbumAssets(AlbumID, GetAlbumAssetsResponse) error { return errors.New("noop") }
func (noopClient) StoreAlbums(GetAlbumsResponse) error                    { return errors.New("noop") }
func (noopClient) StoreAsset(*Asset) error                                { return errors.New("noop") }
