package immich

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/dustin/go-humanize"
)

// localStorageClient is a rwClient for storing and retrieving assets and
// metadata from local persistent storage.
type localStorageClient struct {
	conf LocalConfig
}

// GetAlbumAssets attempts to retrieve the asset metadata for the given album
// from the filesystem. An error is returned if the data is not available.
func (l localStorageClient) GetAlbumAssets(id AlbumID) (*GetAlbumAssetsResponse, error) {
	key := albumKey(id)
	data, err := l.get(key)
	if err != nil {
		return nil, err
	}
	var resp GetAlbumAssetsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// StoreAlbumAssets attempts to write the asset metadata for the given album to
// the filesystem.
func (l localStorageClient) StoreAlbumAssets(id AlbumID, resp GetAlbumAssetsResponse) error {
	key := albumKey(id)
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return l.store(key, data)
}

// GetAlbums attempts to retrieve the list of albums from the filesystem. An
// error is returned if the data is not available.
func (l localStorageClient) GetAlbums() (*GetAlbumsResponse, error) {
	key := albumsKey()
	data, err := l.get(key)
	if err != nil {
		return nil, err
	}
	var resp GetAlbumsResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// StoreAlbums attempts to write the list of albums to the filesystem.
func (l localStorageClient) StoreAlbums(resp GetAlbumsResponse) error {
	key := albumsKey()
	data, err := json.Marshal(resp)
	if err != nil {
		return err
	}
	return l.store(key, data)
}

// GetAsset attempts to retrieve the asset from the filesystem. An error is
// returned if the data is not available.
func (l localStorageClient) GetAsset(md AssetMetadata) (*Asset, error) {
	key := assetKey(md.ID)
	data, err := l.get(key)
	if err != nil {
		return nil, err
	}
	return &Asset{
		Meta: md,
		Data: data,
	}, nil
}

// StoreAsset attempts to write the asset to the filesystem.
func (l localStorageClient) StoreAsset(asset *Asset) error {
	key := assetKey(asset.Meta.ID)
	return l.store(key, asset.Data)
}

// get is a helper method to convert the key to a filepath and read the
// contents of the file.
func (l localStorageClient) get(key string) ([]byte, error) {
	return os.ReadFile(filepath.Join(l.conf.LocalStoragePath, key))
}

// store is a helper method to convert the key to a filepath and write the data
// to disk.
func (l localStorageClient) store(key string, data []byte) error {
	path := filepath.Join(l.conf.LocalStoragePath, key)
	if hasSpace, err := l.hasSpace(path, int64(len(data))); err == nil && !hasSpace {
		if err := l.evict(len(data)); err != nil {
			return fmt.Errorf("not enough space: %w", err)
		}
	} else if err != nil {
		return fmt.Errorf("failed to check local storage space: %w", err)
	}
	return os.WriteFile(filepath.Join(l.conf.LocalStoragePath, key), data, 0644)
}

// hasSpace calculates the amount of space used in the configured path and
// checks if there is enough space to hold the amount of bytes requested. The
// path argument is the path in which the amount of bytes will go, so if there
// is an existing file, we can "reclaim" some of that space.
func (l localStorageClient) hasSpace(path string, bytesRequested int64) (bool, error) {
	bytesUsed, err := l.bytesInUse()
	if err != nil {
		return false, fmt.Errorf("failed to calculate bytes used: %w", err)
	}

	bytesToReplace := int64(0)
	{
		info, err := os.Stat(path)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			slog.Debug("could not stat file", "path", path, "error", err)
			return false, fmt.Errorf("failed to get existing bytes for file: %w", err)
		} else if err == nil {
			bytesToReplace = info.Size()
		}
	}

	slog.Debug("calculated local storage space",
		"storage_configured", l.conf.LocalStorageSize.String(),
		"storage_used", humanize.Bytes(uint64(bytesUsed)),
		"storage_needed", humanize.Bytes(uint64(bytesRequested-bytesToReplace)),
	)
	return bytesUsed-bytesToReplace+bytesRequested <= int64(l.conf.LocalStorageSize), nil
}

// bytesInUse calculates the number of bytes used in the configured path. It
// takes a best-effort approach and only returns an error if there was an error
// opening the top-level directory.
func (l localStorageClient) bytesInUse() (int64, error) {
	var totalSize int64
	err := filepath.WalkDir(l.conf.LocalStoragePath, func(path string, d fs.DirEntry, err error) error {
		// If there was an error opening the root path, surface that error.
		if path == l.conf.LocalStoragePath && err != nil {
			return err
		}
		// Otherwise, ignore errors and directories.
		if err != nil || d.IsDir() {
			return nil
		}
		// Ignore any errors to do a best effort approach.
		if info, err := d.Info(); err == nil {
			totalSize += info.Size()
		}
		return nil
	})
	return totalSize, err
}

// evict attempts to make enough room in the configured directory path to hold
// nbytes more bytes.
func (l localStorageClient) evict(nbytes int) error {
	return errors.New("no eviction policy")
}

// newInMemoryCacheClient initializes a [localStorageClient] client.
func newLocalStorageClient(conf LocalConfig) localStorageClient {
	conf.LocalStoragePath = filepath.Clean(conf.LocalStoragePath)
	return localStorageClient{conf}
}
