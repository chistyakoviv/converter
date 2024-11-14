package conversion

import "fmt"

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

func defaultFormatFor(fileType string) (string, error) {
	if !isSupported(fileType) {
		return "", fmt.Errorf("file type '%s' not supported", fileType)
	}
	return FileTypeToFormatMap[fileType].DefaultFormat, nil
}
