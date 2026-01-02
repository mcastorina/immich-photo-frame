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

func (pf *photoFrame) startSlideshowWorker(slidesCh <-chan decodedAsset, keyCh <-chan *fyne.KeyEvent) <-chan decodedAsset {
	// displayCh should be unbuffered since it controls when to display the
	// next image.
	displayCh := make(chan decodedAsset)
	go func() {
		ticker := time.NewTicker(pf.conf.App.ImageDelay)
		history := make([]decodedAsset, pf.conf.App.HistorySize+1)
		historyIndex := len(history) - 1
		next := func() {
			if historyIndex < len(history)-1 {
				historyIndex++
				return
			}
			history = append(history, <-slidesCh)
			history = history[1:]
		}
		prev := func() {
			if historyIndex > 0 && history[historyIndex-1].img != nil {
				historyIndex--
			}
		}

		// Initialize history object and display the first image.
		next()
		displayCh <- history[historyIndex]
		for {
			select {
			case <-ticker.C:
				next()
			case k := <-keyCh:
				ticker.Reset(pf.conf.App.ImageDelay)
				switch k.Name {
				case fyne.KeyRight:
					next()
				case fyne.KeyLeft:
					prev()
				}
			}
			// Send the image to be displayed.
			displayCh <- history[historyIndex]
		}
	}()
	return displayCh
}

// startDisplayWorker starts a goroutine that waits for an asset from assCh and
// displays it.
func (pf *photoFrame) startDisplayWorker(img *canvas.Image, assCh <-chan decodedAsset) {
	go func() {
		// Wait for something to display.
		for da := range assCh {
			// Tell fyne to display it.
			fyne.DoAndWait(func() {
				slog.Info("displaying image", "name", da.meta.Name, "id", da.meta.ID)
				img.Image = da.img
				img.Refresh()
			})
		}
	}()
}

// assetWorker iterates through the albums and assets and puts them on the
// asset queue.
func (pf *photoFrame) startSlidesWorker(albums []immich.Album) <-chan decodedAsset {
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
