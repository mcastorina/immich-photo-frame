package main

import (
	"digital-photo-frame/internal/immich"
	"sync"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
)

const (
	assetID = "5c01c1cd-3c85-47bb-a3d0-007197fc2dbd"
	// video asset ID
	// assetID = "416e4480-5754-4a76-b3a8-5499a7c4b5e7"
)

func main() {
	client := immich.NewClientFromEnv()
	if err := client.IsConnected(); err != nil {
		panic(err)
	}

	album, err := client.GetAlbumByName("Chrismukkah 2025")
	if err != nil {
		panic(err)
	}
	albumAssets, err := client.GetAlbumAssets(album.ID)
	if err != nil {
		panic(err)
	}

	assets := make([]*immich.Asset, len(albumAssets))
	var wg sync.WaitGroup
	for i, md := range albumAssets {
		wg.Go(func() {
			assets[i], _ = client.GetAsset(md)
		})
	}
	wg.Wait()

	a := app.New()
	w := a.NewWindow("immich")
	w.SetFullScreen(true)
	showImage(w, assets[0])
	go func() {
		index := 0
		ticker := time.NewTicker(3 * time.Second)
		for range ticker.C {
			fyne.Do(func() {
				for assets[index].Meta.Type != "IMAGE" {
					index = (index + 1) % len(albumAssets)
				}
				// asset, err := client.GetAsset(albumAssets[index])
				showImage(w, assets[index])
				index = (index + 1) % len(albumAssets)
			})
		}
	}()
	w.ShowAndRun()
}

func showImage(w fyne.Window, r fyne.Resource) {
	img := canvas.NewImageFromResource(r)
	img.FillMode = canvas.ImageFillContain

	w.SetContent(img)
}
