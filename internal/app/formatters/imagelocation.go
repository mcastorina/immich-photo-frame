package formatters

import (
	"fmt"

	"immich-photo-frame/internal/immich"
)

type ImageLocation struct{}

func (ImageLocation) Name() string { return "image-location" }

func (ImageLocation) Format(meta immich.AssetMetadata) string {
	exifInfo := meta.ExifInfo
	const usa = "United States of America"
	city, state, country := exifInfo.City, exifInfo.State, exifInfo.Country
	switch {
	case country != usa && country != "" && city != "":
		return fmt.Sprintf("%s, %s", city, country)
	case country != usa && country != "":
		return country
	case country == usa && city != "" && state != "":
		return fmt.Sprintf("%s, %s", city, state)
	case country == usa && city != "":
		return city
	case country == usa && state != "":
		return state
	}
	return ""
}
