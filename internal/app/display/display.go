package display

import (
	"bytes"
	"errors"
	"image"
	"image/color"
	"log/slog"
	"math"
	"strconv"
	"time"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"

	"github.com/disintegration/imaging"
	"github.com/dustin/go-humanize"

	"immich-photo-frame/internal/immich"
)

// Config holds configuration values for displaying assets.
//
// It is organized to take advantage of TOML parsing, however this package does
// not handle parsing and has no expectation on how it will be initialized.
type Config struct {
	ImageScale float32
}

// Display controls the actual GUI application, such as the window, image, and
// text overrlay.
type Display struct {
	conf Config
	win  fyne.Window
	img  *canvas.Image
	text *canvas.Text
}

// DecodedAsset is an asset that is ready to be displayed.
type DecodedAsset struct {
	Meta immich.AssetMetadata
	Img  image.Image
}

// New initializes a Display with the provided configuration.
func New(conf Config) *Display {
	a := app.New()
	// TODO: Make a custom theme since DarkTheme is deprecated.
	a.Settings().SetTheme(theme.DarkTheme())
	a.Driver().SetDisableScreenBlanking(true)
	win := a.NewWindow("immich")
	win.Resize(fyne.NewSize(200, 200))
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
		hiddenCursorOverlay{},
	)
	win.SetContent(content)

	return &Display{conf, win, img, text}
}

// SetKeyBinds registers the provided callback to be executed when a key is
// pressed.
func (d *Display) SetKeyBinds(f func(*fyne.KeyEvent)) {
	d.win.Canvas().SetOnTypedKey(f)
}

// Show tells the Display to display the DecodedAsset now.
func (d *Display) Show(da DecodedAsset) {
	fyne.Do(func() {
		slog.Info("displaying image", "name", da.Meta.Name, "id", da.Meta.ID)
		d.img.Image = da.Img
		d.text.Text = d.formattedDateTime(da.Meta.ExifInfo)
		d.img.Refresh()
		d.text.Refresh()
	})
}

// DecodeAsset takes an immich.Asset and transforms it in preparation for
// display. This allows the work to be done ahead of calling [Show], since
// decoding can be a significant amount of work.
func (d *Display) DecodeAsset(ass *immich.Asset) (*DecodedAsset, error) {
	img, _, err := image.Decode(bytes.NewReader(ass.Data))
	if err != nil {
		return nil, err
	}
	imgHeight := img.Bounds().Dy()
	resizeHeight := int(float32(imgHeight) * d.conf.ImageScale)
	if resizeHeight != imgHeight {
		img = imaging.Resize(img, 0, resizeHeight, imaging.Lanczos)
	}
	return &DecodedAsset{
		Meta: ass.Meta,
		Img:  img,
	}, nil
}

// ShowAndRun starts the GUI and runs the application. This method must be
// called from the main thread and blocks until the application is closed.
func (d *Display) ShowAndRun() {
	d.win.ShowAndRun()
}

// formattedDateTime is a helper method to format the EXIF time information
// into human-readable text.
func (d *Display) formattedDateTime(exifInfo immich.ExifInfo) string {
	// Parse EXIF timestamp.
	t, err := time.Parse("2006-01-02T15:04:05.999Z07:00", exifInfo.DateTimeOriginal)
	if err != nil {
		slog.Error("failed to parse timestamp",
			"error", err,
			"timezone", exifInfo.DateTimeOriginal,
		)
		return ""
	}

	// Parse EXIF timezone.
	loc, err := parseTimeZone(exifInfo.TimeZone)
	if err != nil {
		slog.Error("failed to parse timezone",
			"error", err,
			"timezone", exifInfo.TimeZone,
		)
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

// parseTimeZone is a helper function to parse the EXIF timezone string into a
// time.Location. It first tries to load the location directly, and if that
// doesn't work, it tries parsing it as a UTC offset.
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

// hiddenCursorOverlay implements fyne.CanvasObject and desktop.Cursorable to
// sit on top of the window and hide the cursor.
type hiddenCursorOverlay struct{}

func (h hiddenCursorOverlay) Hide()                   {}
func (h hiddenCursorOverlay) MinSize() fyne.Size      { return fyne.NewSize(0, 0) }
func (h hiddenCursorOverlay) Move(fyne.Position)      {}
func (h hiddenCursorOverlay) Position() fyne.Position { return fyne.NewPos(-10, -10) }
func (h hiddenCursorOverlay) Refresh()                {}
func (h hiddenCursorOverlay) Resize(fyne.Size)        {}
func (h hiddenCursorOverlay) Show()                   {}
func (h hiddenCursorOverlay) Size() fyne.Size         { return fyne.NewSize(math.MaxFloat32, math.MaxFloat32) }
func (h hiddenCursorOverlay) Visible() bool           { return true }
func (h hiddenCursorOverlay) Cursor() desktop.Cursor  { return desktop.HiddenCursor }
