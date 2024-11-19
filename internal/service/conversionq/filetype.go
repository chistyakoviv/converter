package conversionq

func isSupported(fileType string) bool {
	_, supported := FileTypeToFormatMap[fileType]
	return supported
}

func isConvertible(from, to string) bool {
	formatInfo, supported := FileTypeToFormatMap[from]
	if !supported {
		return false
	}
	_, convertible := formatInfo.SupportedFormats[to]
	return convertible
}

func IsImage(ext string) bool {
	_, isImage := ImageFormats[ext]
	return isImage
}

func IsVideo(ext string) bool {
	_, isVideo := VideoFormats[ext]
	return isVideo
}
