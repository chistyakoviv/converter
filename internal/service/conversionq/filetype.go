package conversionq

import "github.com/chistyakoviv/converter/internal/model"

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

func mediaTypeForExt(fileType string) (int, bool) {
	switch fileType {
	case "jpg", "jpeg", "png":
		return model.MediaTypeImage, true
	case "mp4", "webm":
		return model.MediaTypeVideo, true
	default:
		return 0, false
	}
}
