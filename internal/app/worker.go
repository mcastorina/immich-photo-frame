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
	pf.displayAsset(img, <-pf.imgQueue)
	for {
		select {
		case <-ticker.C:
		case _ = <-keyPress:
			ticker.Reset(pf.conf.App.ImageDelay)
		}
		pf.displayAsset(img, <-pf.imgQueue)
	}
}

func (pf *photoFrame) displayAsset(img *canvas.Image, content image.Image) {
	fyne.DoAndWait(func() {
		img.Image = content
		img.Refresh()
	})
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

				img, _, err := image.Decode(bytes.NewReader(ass.Data))
				if err != nil {
					slog.Error("failed to decode image", "asset", assMeta, "error", err)
					continue
				}
				winHeight := pf.win.Canvas().Size().Height
				resizeHeight := int(winHeight * pf.conf.App.ImageScale)
				img = imaging.Resize(img, 0, resizeHeight, imaging.Lanczos)

				pf.imgQueue <- img
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
