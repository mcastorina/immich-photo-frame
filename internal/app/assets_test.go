package app_test

import (
	"errors"
	"immich-photo-frame/internal/app"
	"immich-photo-frame/internal/immich"
	"testing"
)

type testAlbumAssetGetter map[immich.AlbumID][]immich.AssetMetadata

// GetAlbumAssets implements app.albumAssetGetter.
func (t testAlbumAssetGetter) GetAlbumAssets(id immich.AlbumID) ([]immich.AssetMetadata, error) {
	if t == nil {
		return nil, errors.New("empty")
	}
	mds, ok := t[id]
	if !ok {
		return nil, errors.New("not found")
	}
	return mds, nil
}

func TestAssetMetadataIter_Iter(t *testing.T) {
	source := testAlbumAssetGetter(map[immich.AlbumID][]immich.AssetMetadata{
		"album-1": []immich.AssetMetadata{
			{ID: "asset-1"},
			{ID: "asset-2"},
		},
		"album-2": []immich.AssetMetadata{
			{ID: "asset-3"},
			{ID: "asset-4"},
			{ID: "asset-5"},
		},
	})
	albums := []immich.Album{
		{ID: "album-1", AssetCount: 2},
		{ID: "album-2", AssetCount: 3},
	}
	count := 0
	var got []immich.AssetMetadata
	app.AssetMetadataIter(source, albums)(func(md *immich.AssetMetadata) bool {
		got = append(got, *md)
		count++
		return count < 6
	})
	if len(got) != 6 {
		t.Fatalf("expected 6 elements, got %d", len(got))
	}
	if id := got[0].ID; id != "asset-1" {
		t.Fatalf(`first element should be "asset-1", got %q`, id)
	}
	if id := got[1].ID; id != "asset-2" {
		t.Fatalf(`second element should be "asset-2", got %q`, id)
	}
	if id := got[2].ID; id != "asset-3" {
		t.Fatalf(`third element should be "asset-3", got %q`, id)
	}
	if id := got[3].ID; id != "asset-4" {
		t.Fatalf(`fourth element should be "asset-4", got %q`, id)
	}
	if id := got[4].ID; id != "asset-5" {
		t.Fatalf(`fifth element should be "asset-5", got %q`, id)
	}
	if id := got[5].ID; id != "asset-1" {
		t.Fatalf(`sixth element should be "asset-1", got %q`, id)
	}
}

func TestAssetMetadataIter_Iter_NoAlbums(t *testing.T) {
	source := testAlbumAssetGetter(nil)
	funcCalled := false
	app.AssetMetadataIter(source, nil)(func(md *immich.AssetMetadata) bool {
		funcCalled = true
		return true
	})
	if funcCalled {
		t.Fatal("the iter callback should not be executed if there are no albums")
	}
}

func TestAssetMetadataIter_Iter_NoAssetCounts(t *testing.T) {
	source := testAlbumAssetGetter(nil)
	albums := []immich.Album{
		{ID: "album-1", AssetCount: 0},
		{ID: "album-2", AssetCount: 0},
	}
	funcCalled := false
	app.AssetMetadataIter(source, albums)(func(md *immich.AssetMetadata) bool {
		funcCalled = true
		return true
	})
	if funcCalled {
		t.Fatal("the iter callback should not be executed if there are no asset counts")
	}
}

func TestAssetMetadataIter_Iter_NoAssets(t *testing.T) {
	source := testAlbumAssetGetter(map[immich.AlbumID][]immich.AssetMetadata{
		"album-1": nil,
		"album-2": nil,
	})
	albums := []immich.Album{
		{ID: "album-1", AssetCount: 2},
		{ID: "album-2", AssetCount: 3},
	}
	funcCalled := false
	app.AssetMetadataIter(source, albums)(func(md *immich.AssetMetadata) bool {
		funcCalled = true
		return true
	})
	if funcCalled {
		t.Fatal("the iter callback should not be executed if there are no assets")
	}
}
