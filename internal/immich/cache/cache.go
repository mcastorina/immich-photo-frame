package cache

import "immich-photo-frame/internal/immich"

type Client struct {
	cache  rwClient
	local  rwClient
	remote readClient
}

type rwClient interface {
	readClient
	writeClient
}

type writeClient interface {
	StoreAsset(asset *immich.Asset) error
}

type readClient interface {
	IsConnected() error
	GetAsset(md immich.AssetMetadata) (*immich.Asset, error)
}

type clientOpt func(*Client)

func WithInMemoryCache(conf InMemoryConfig) clientOpt {
	return func(c *Client) {
		if !conf.UseInMemoryCache {
			return
		}
		panic("not implemented")
	}
}

func WithLocalStorage(conf LocalConfig) clientOpt {
	return func(c *Client) {
		if !conf.UseLocalStorage {
			return
		}
		panic("not implemented")
	}
}

func WithRemote(conf immich.Config) clientOpt {
	return func(c *Client) {
		c.remote = immich.NewClient(conf)
	}
}

func NewClient(opts ...clientOpt) *Client {
	client := &Client{}
	for _, opt := range opts {
		opt(client)
	}
	return client
}
