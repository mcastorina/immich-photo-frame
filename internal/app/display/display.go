package display

import (
	"bytes"
	"image"
	"image/color"
	"log/slog"
	"math"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/canvas"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/driver/desktop"
	"fyne.io/fyne/v2/layout"
	"fyne.io/fyne/v2/theme"

	"github.com/disintegration/imaging"

	"immich-photo-frame/internal/app/formatters"
	"immich-photo-frame/internal/immich"
)

// Config holds configuration values for displaying assets.
//
// It is organized to take advantage of TOML parsing, however this package does
// not handle parsing and has no expectation on how it will be initialized.
type Config struct {
	ImageScale float32
	ImageText  []formatters.FormatConfig
}

// Display controls the actual GUI application, such as the window, image, and
// text overrlay.
type Display struct {
	conf  Config
	win   fyne.Window
	img   *canvas.Image
	texts []*canvas.Text
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

	texts := make([]*canvas.Text, len(conf.ImageText))
	textObjs := make([]fyne.CanvasObject, len(conf.ImageText))
	for i := range len(conf.ImageText) {
		texts[i] = canvas.NewText("", color.White)
		texts[i].Alignment = fyne.TextAlignTrailing
		texts[i].TextSize = conf.ImageText[i].Size()
		textObjs[i] = texts[i]
	}

	textBlock := container.NewVBox(textObjs...)

	// Create a container with the image and bottom-right aligned text.
	content := container.NewStack(
		img,
		container.NewBorder(nil,
			container.NewHBox(layout.NewSpacer(), textBlock),
			nil, nil),
		hiddenCursorOverlay{},
	)
	win.SetContent(content)

	return &Display{conf, win, img, texts}
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
		d.img.Refresh()
		for i := range d.texts {
			d.texts[i].Text = d.conf.ImageText[i].Format(da.Meta)
			d.texts[i].Refresh()
		}
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
