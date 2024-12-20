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
