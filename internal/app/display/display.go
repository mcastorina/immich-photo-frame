package display

import (
	"bytes"
	"image"
	"immich-photo-frame/internal/immich"
	"log/slog"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/theme"
	"github.com/disintegration/imaging"
)

type Config struct {
	ImageScale float32
}

type Display struct {
	conf Config
	win  fyne.Window
	img  *canvas.Image
}

type DecodedAsset struct {
	Meta immich.AssetMetadata
	Img  image.Image
}

func New(conf Config) *Display {
	a := app.New()
	// TODO: Make a custom theme since DarkTheme is deprecated.
	a.Settings().SetTheme(theme.DarkTheme())
	a.Driver().SetDisableScreenBlanking(true)
	win := a.NewWindow("immich")
	win.SetFullScreen(true)

	img := canvas.NewImageFromResource(nil)
	img.FillMode = canvas.ImageFillContain
	img.ScaleMode = canvas.ImageScaleSmooth
	win.SetContent(img)

	return &Display{conf, win, img}
}

func (d *Display) SetKeyBinds(f func(*fyne.KeyEvent)) {
	d.win.Canvas().SetOnTypedKey(f)
}

func (d *Display) Show(da DecodedAsset) {
	fyne.Do(func() {
		slog.Info("displaying image", "name", da.Meta.Name, "id", da.Meta.ID)
		d.img.Image = da.Img
		d.img.Refresh()
	})
}

func (d *Display) DecodeAsset(ass *immich.Asset) (*DecodedAsset, error) {
	img, _, err := image.Decode(bytes.NewReader(ass.Data))
	if err != nil {
		return nil, err
	}
	imgHeight := float32(img.Bounds().Dy())
	resizeHeight := int(imgHeight * d.conf.ImageScale)
	img = imaging.Resize(img, 0, resizeHeight, imaging.Lanczos)
	return &DecodedAsset{
		Meta: ass.Meta,
		Img:  img,
	}, nil
}

func (d *Display) ShowAndRun() {
	d.win.ShowAndRun()
}
