package formatters

import (
	"fmt"
	"immich-photo-frame/internal/immich"
	"strconv"
	"strings"
)

type TextFormatter interface {
	Name() string
	Format(immich.AssetMetadata) string
}

type TextSizeFormatter interface {
	TextFormatter
	Size() float32
}

type FormatConfig struct {
	TextSizeFormatter
}

type SizeWrapper struct {
	TextFormatter
	size float32
}

func (s SizeWrapper) Size() float32 { return s.size }

var (
	// formatters is the list of available text formatters.
	formatters = []TextFormatter{
		new(ImageDateTime),
		new(ImageLocation),
	}

	// formattersByName is a LUT of name to TextFormatter, built via [init].
	formattersByName = map[string]TextFormatter{}
)

// UnmarshalText implements toml.TextUnmarshaler.
func (f *FormatConfig) UnmarshalText(text []byte) error {
	name, size, err := parseFormatter(string(text))
	if err == nil {
		fc, ok := formattersByName[name]
		if ok {
			f.TextSizeFormatter = SizeWrapper{TextFormatter: fc, size: size}
			return nil
		}
	}
	if _, ok := formattersByName[name]; ok {
		return fmt.Errorf("failed parsing %q: %w", string(text), err)
	}
	// Unrecognized formatter.
	var validFormatters []string
	for key := range formattersByName {
		validFormatters = append(validFormatters, fmt.Sprintf("%q", key))
	}
	return fmt.Errorf(
		"unsupported text formatter %q, expected one of %v",
		string(text), validFormatters,
	)
}

// parseFormatter is a helper function to parse a string representation of a
// formatter into its name and size.
//
// Valid formats:
// - "name"
// - "name:size"
func parseFormatter(text string) (string, float32, error) {
	name, sizeText, found := strings.Cut(text, ":")
	name = strings.ToLower(name)
	if !found {
		return name, 16, nil
	}
	size, err := strconv.ParseFloat(sizeText, 32)
	if err != nil {
		return name, 0, fmt.Errorf("invalid size %q, expected float", sizeText)
	}
	return name, float32(size), nil
}

// MarshalJSON implements json.Marshaler.
func (f *FormatConfig) MarshalJSON() ([]byte, error) {
	return fmt.Appendf(nil, `{"name":%q,"size":%f}`,
		f.TextSizeFormatter.Name(), f.TextSizeFormatter.Size(),
	), nil
}

func init() {
	for _, fc := range formatters {
		formattersByName[fc.Name()] = fc
	}
}

