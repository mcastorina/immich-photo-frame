package formatters

import (
	"errors"
	"log/slog"
	"strconv"
	"time"

	"github.com/dustin/go-humanize"

	"immich-photo-frame/internal/immich"
)

type ImageDateTime struct{}

func (ImageDateTime) Name() string { return "image-date-time" }

func (ImageDateTime) Format(meta immich.AssetMetadata) string {
	exifInfo := meta.ExifInfo

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
