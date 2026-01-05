package controller

import (
	"immich-photo-frame/internal/immich"
	"testing"
)

func Test_getConfiguredAlbums_Stable(t *testing.T) {
	got := getConfiguredAlbums(
		[]immich.Album{
			{Name: "album-2"},
			{Name: "album-1"},
		},
		[]string{"album-1", "album-2"},
	)
	if len(got) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(got))
	}
	if got[0].Name != "album-1" {
		t.Fatalf(`expected first element to be "album-1", found %q`, got[0].Name)
	}
	if got[1].Name != "album-2" {
		t.Fatalf(`expected second element to be "album-2", found %q`, got[1].Name)
	}
}

func Test_getConfiguredAlbums_Duplicate(t *testing.T) {
	got := getConfiguredAlbums(
		[]immich.Album{{Name: "album-1"}},
		[]string{"album-1", "album-1"},
	)
	if len(got) != 2 {
		t.Fatalf("expected 2 elements, got %d", len(got))
	}
	if got[0].Name != "album-1" {
		t.Fatalf(`expected first element to be "album-1", found %q`, got[0].Name)
	}
	if got[1].Name != "album-1" {
		t.Fatalf(`expected second element to be "album-1", found %q`, got[1].Name)
	}
}
