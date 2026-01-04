package planners_test

import (
	"errors"
	"testing"

	"immich-photo-frame/internal/app/controller/planners"
	"immich-photo-frame/internal/immich"
)

var _ planners.AssetClient = testAssetClient{}

// testAssetClient is a test implementation of [planners.AssetClient].
type testAssetClient struct {
	lut map[immich.AlbumID][]immich.AssetMetadata
}

// GetAlbumAssets implements planners.AssetClient.
func (t testAssetClient) GetAlbumAssets(id immich.AlbumID) ([]immich.AssetMetadata, error) {
	ass, ok := t.lut[id]
	if !ok {
		return nil, errors.New("not found")
	}
	// Make a copy since Sequential modifies the slice.
	c := make([]immich.AssetMetadata, len(ass))
	copy(c, ass)
	return c, nil
}

// TestSequentialPlanner tests SequentialPlanner iterates over configured
// albums and assets in the order received.
func TestSequentialPlanner(t *testing.T) {
	var seq planners.Sequential
	client := testAssetClient{
		lut: map[immich.AlbumID][]immich.AssetMetadata{
			"album-1": {
				{ID: "asset-1"},
				{ID: "asset-2"},
			},
			"album-2": {
				{ID: "asset-3"},
				{ID: "asset-4"},
				{ID: "asset-5"},
			},
		},
	}

	// Test configured album.
	t.Run("one configured albums", func(t *testing.T) {
		seq.Init(client, []immich.Album{{ID: "album-1"}})
		var gotIDs []immich.AssetID
		for range 3 {
			ass := seq.Next()
			if ass != nil {
				gotIDs = append(gotIDs, ass.ID)
			}
		}

		if len(gotIDs) != 3 {
			t.Fatalf("expected 3 items, found %d", len(gotIDs))
		}
		expectedIDs := []immich.AssetID{
			"asset-1", "asset-2", "asset-1",
		}
		for i := range 3 {
			if gotIDs[i] != expectedIDs[i] {
				t.Fatalf(`gotIDs[%d] should be %q, found %q`, i, expectedIDs[i], gotIDs[i])
			}
		}
	})

	// Test multiple configured albums.
	t.Run("multiple configured albums", func(t *testing.T) {
		seq.Init(client, []immich.Album{{ID: "album-1"}, {ID: "album-2"}})
		var gotIDs []immich.AssetID
		for range 7 {
			ass := seq.Next()
			if ass != nil {
				gotIDs = append(gotIDs, ass.ID)
			}
		}

		if len(gotIDs) != 7 {
			t.Fatalf("expected 7 items, found %d", len(gotIDs))
		}
		expectedIDs := []immich.AssetID{
			"asset-1", "asset-2",
			"asset-3", "asset-4", "asset-5",
			"asset-1", "asset-2",
		}
		for i := range 7 {
			if gotIDs[i] != expectedIDs[i] {
				t.Fatalf(`gotIDs[%d] should be %q, found %q`, i, expectedIDs[i], gotIDs[i])
			}
		}
	})
}

func TestSequentialAsc(t *testing.T) {
	var seq planners.Sequential
	client := testAssetClient{
		lut: map[immich.AlbumID][]immich.AssetMetadata{
			"album-1": {
				{ID: "asset-1"},
				{ID: "asset-2"},
				{ID: "asset-3"},
			},
		},
	}

	seq.Init(client, []immich.Album{{ID: "album-1", Order: "asc"}})
	var gotIDs []immich.AssetID
	for range 7 {
		ass := seq.Next()
		if ass != nil {
			gotIDs = append(gotIDs, ass.ID)
		}
	}

	if len(gotIDs) != 7 {
		t.Fatalf("expected 7 items, found %d", len(gotIDs))
	}
	expectedIDs := []immich.AssetID{
		"asset-3", "asset-2", "asset-1",
		"asset-3", "asset-2", "asset-1",
		"asset-3",
	}
	for i := range 7 {
		if gotIDs[i] != expectedIDs[i] {
			t.Fatalf(`gotIDs[%d] should be %q, found %q`, i, expectedIDs[i], gotIDs[i])
		}
	}
}
