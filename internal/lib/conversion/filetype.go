package conversion

import "fmt"

func IsSupported(fileType string) bool {
	_, ok := FileTypeToFormatMap[fileType]
	return ok
}

func IsConvertable(from string, to string) bool {
	if !IsSupported(from) {
		return false
	}
	return FileTypeToFormatMap[from].SupportedFormats[to]
}

func Default(fileType string) (string, error) {
	if !IsSupported(fileType) {
		return "", fmt.Errorf("file type \"%s\" not supported", fileType)
	}
	return FileTypeToFormatMap[fileType].DefaultFormat, nil
}
