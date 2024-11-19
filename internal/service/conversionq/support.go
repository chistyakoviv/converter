package conversionq

type FormatInfo struct {
	SupportedFormats map[string]bool // Maps each supported format to `true`
}

var (
	ImageFormats = FormatInfo{
		SupportedFormats: map[string]bool{
			"webp": true,
			"avif": true,
		},
	}
	VideoFormats = FormatInfo{
		SupportedFormats: map[string]bool{
			"webm": true,
		},
	}
	FileTypeToFormatMap = map[string]*FormatInfo{
		"jpg":  &ImageFormats,
		"jpeg": &ImageFormats,
		"png":  &ImageFormats,
		"mp4":  &VideoFormats,
	}
)
