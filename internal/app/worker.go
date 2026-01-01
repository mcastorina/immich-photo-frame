package app

import (
	"log/slog"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/canvas"

	"immich-photo-frame/internal/immich"
)

// displayWorker pulls from the asset queue every 3s and displays the image.
func (pf *photoFrame) displayWorker() {
	img := canvas.NewImageFromResource(nil)
	img.FillMode = canvas.ImageFillContain
	img.ScaleMode = canvas.ImageScaleSmooth
	pf.win.SetContent(img)

	keyPress := make(chan fyne.KeyName, 10)
	pf.win.Canvas().SetOnTypedKey(func(ke *fyne.KeyEvent) {
		switch ke.Name {
		case fyne.KeyRight:
			// Non-blocking channel write.
			select {
			case keyPress <- ke.Name:
			default:
			}
		}
	})

	ticker := time.NewTicker(pf.conf.App.ImageDelay)
	pf.displayAsset(img, pf.getNextAsset())
	for {
		select {
		case <-ticker.C:
		case _ = <-keyPress:
			ticker.Reset(pf.conf.App.ImageDelay)
		}
		ass := pf.getNextAsset()
		pf.displayAsset(img, ass)
	}
}

func (pf *photoFrame) displayAsset(img *canvas.Image, ass *immich.Asset) {
	fyne.DoAndWait(func() {
		slog.Info("displaying image",
			"name", ass.Meta.Name,
			"id", ass.Meta.ID,
		)
		img.Resource = ass
		img.Refresh()
	})
}

// getNextAsset gets the next asset from the asset queue. Currently only IMAGE
// assets are supported, so all others are skipped. This method blocks until a
// valid asset is found.
func (pf *photoFrame) getNextAsset() *immich.Asset {
	for {
		ass := <-pf.assQueue
		if ass.Meta.Type == "IMAGE" {
			return ass
		}
		slog.Warn("unsupported asset type, skipping",
			"type", ass.Meta.Type,
			"id", ass.Meta.ID,
		)
	}
}

// assetWorker iterates through the albums and assets and puts them on the
// asset queue.
func (pf *photoFrame) assetWorker(albums []immich.Album) {
	for {
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
				pf.assQueue <- ass
			}
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
