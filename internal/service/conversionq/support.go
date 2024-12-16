package conversionq

type FormatInfo struct {
	SupportedFormats map[string]bool // Maps each supported format to `true`
}

var (
	ImageConversionFormats = FormatInfo{
		SupportedFormats: map[string]bool{
			"jpg":  true,
			"jpeg": true,
			"png":  true,
			"webp": true,
			"avif": true,
		},
	}

	VideoConversionFormats = FormatInfo{
		SupportedFormats: map[string]bool{
			"webm": true,
			"mp4":  true,
		},
	}

	FileTypeToFormatMap = map[string]*FormatInfo{
		"jpg":  &ImageConversionFormats,
		"jpeg": &ImageConversionFormats,
		"png":  &ImageConversionFormats,
		"mp4":  &VideoConversionFormats,
	}

	ImageFormats = map[string]bool{
		"jpg":  true,
		"jpeg": true,
		"png":  true,
	}

	VideoFormats = map[string]bool{
		"mp4":  true,
		"webm": true,
	}
)
