package app

import (
	"bytes"
	"image"
	"log/slog"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"

	"github.com/disintegration/imaging"

	"immich-photo-frame/internal/immich"
)

// decodedAsset combines the AssetMetadata with the decoded image.Image object.
// This is used to move some of the image processing to the background task, so
// the displayWorker has less work to do.
type decodedAsset struct {
	meta immich.AssetMetadata
	img  image.Image
}

// displayWorker pulls from the asset queue every 3s and displays the image.
func (pf *photoFrame) displayWorker(ch <-chan decodedAsset, img *canvas.Image) {
	keyPress := make(chan fyne.KeyName, 10)

	ticker := time.NewTicker(pf.conf.App.ImageDelay)
	pf.displayAsset(img, <-ch)
	for {
		select {
		case <-ticker.C:
		case _ = <-keyPress:
			ticker.Reset(pf.conf.App.ImageDelay)
		}
		pf.displayAsset(img, <-ch)
	}
}

func (pf *photoFrame) displayAsset(img *canvas.Image, da decodedAsset) {
	fyne.DoAndWait(func() {
		slog.Info("displaying image", "name", da.meta.Name, "id", da.meta.ID)
		img.Image = da.img
		img.Refresh()
	})
}

// assetWorker iterates through the albums and assets and puts them on the
// asset queue.
func (pf *photoFrame) startAssetWorker(albums []immich.Album) <-chan decodedAsset {
	ch := make(chan decodedAsset, 10)
	go func() {
		for {
			pf.enumerateAlbumsAndAssets(albums, ch)
		}
	}()
	return ch
}

func (pf *photoFrame) enumerateAlbumsAndAssets(albums []immich.Album, ch chan<- decodedAsset) {
	for _, album := range pf.shuffleAlbums(albums) {
		assMeta, err := pf.client.GetAlbumAssets(album.ID)
		if err != nil {
			slog.Error("failed to load album", "album", album, "error", err)
			continue
		}
		for _, assMeta := range pf.shuffleAssetMetadata(album, assMeta) {
			ass, err := pf.client.GetAsset(assMeta)
			if err != nil {
				slog.Error("failed to load asset", "asset", assMeta, "error", err)
				continue
			}

			img, _, err := image.Decode(bytes.NewReader(ass.Data))
			if err != nil {
				slog.Error("failed to decode image", "asset", ass.Meta, "error", err)
				continue
			}
			imgHeight := float32(img.Bounds().Dy())
			resizeHeight := int(imgHeight * pf.conf.App.ImageScale)
			img = imaging.Resize(img, 0, resizeHeight, imaging.Lanczos)

			ch <- decodedAsset{meta: assMeta, img: img}
		}
	}
}

// shuffleAlbums shuffles the display albums based on user configuration.
func (pf *photoFrame) shuffleAlbums(albums []immich.Album) []immich.Album {
	return albums
}

// shuffleAssetMetadata shuffles the assets in an album based on user
// configuration.
func (pf *photoFrame) shuffleAssetMetadata(
	album immich.Album,
	assets []immich.AssetMetadata,
) []immich.AssetMetadata {
	return assets
}
