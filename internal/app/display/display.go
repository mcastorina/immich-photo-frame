package display

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"log/slog"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"

	"github.com/disintegration/imaging"
	"github.com/dustin/go-humanize"

	"immich-photo-frame/internal/immich"
)

type Config struct {
	ImageScale float32
}

type Display struct {
	conf Config
	win  fyne.Window
	img  *canvas.Image
	text *canvas.Text
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

	text := canvas.NewText("", color.White)
	text.TextSize = 20
	text.Alignment = fyne.TextAlignCenter

	// Create a container with the image and bottom-right aligned text.
	content := container.NewStack(
		img,
		container.NewBorder(nil,
			container.NewHBox(layout.NewSpacer(), text),
			nil, nil),
	)
	win.SetContent(content)

	return &Display{conf, win, img, text}
}

func (d *Display) SetKeyBinds(f func(*fyne.KeyEvent)) {
	d.win.Canvas().SetOnTypedKey(f)
}

func (d *Display) Show(da DecodedAsset) {
	fyne.Do(func() {
		slog.Info("displaying image", "name", da.Meta.Name, "id", da.Meta.ID)
		d.img.Image = da.Img
		d.text.Text = d.formattedDateTime(da.Meta.ExifInfo)
		d.img.Refresh()
		d.text.Refresh()
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

func (d *Display) formattedDateTime(exifInfo immich.ExifInfo) string {
	// Parse EXIF timestamp.
	t, err := time.Parse("2006-01-02T15:04:05.999Z07:00", exifInfo.DateTimeOriginal)
	if err != nil {
		return ""
	}

	// Parse EXIF timezone.
	loc, err := parseTimeZone(exifInfo.TimeZone)
	if err != nil {
		return ""
	}
	t = t.In(loc)

	// Format based on how long ago the asset was.
	elapsed := time.Since(t)
	switch {
	case elapsed < 1*humanize.Week:
		return t.Format("Monday 3:04 PM")
	case elapsed < 3*humanize.Month:
		return humanize.Time(t)
	default:
		return t.Format("January 2, 2006")
	}
}

func parseTimeZone(tz string) (*time.Location, error) {
	// First try loading it as a location.
	if loc, err := time.LoadLocation(tz); err == nil {
		return loc, nil
	}
	// Then try parsing it as a UTC offset in the format "UTC-6" or "UTC+9"
	if len(tz) < 4 || tz[:3] != "UTC" {
		return nil, errors.New("unexpected timezone format")
	}

	hours, err := strconv.Atoi(tz[3:])
	if err != nil {
		return nil, err
	}

	seconds := hours * 60 * 60
	return time.FixedZone(tz, seconds), nil
}
