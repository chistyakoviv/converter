package converter

import (
	"path/filepath"
	"strings"

	"github.com/chistyakoviv/converter/internal/http-server/request"
	"github.com/chistyakoviv/converter/internal/model"
)

func ToConversionInfoFromRequest(dto request.ConversionRequest) *model.ConversionInfo {
	// Extract the file name with extension
	fileName := filepath.Base(dto.Path)

	// Extract the file extension
	fileExt := filepath.Ext(dto.Path)

	return &model.ConversionInfo{
		Fullpath:  dto.Path,
		Path:      filepath.Dir(dto.Path),
		Filestem:  strings.TrimSuffix(fileName, fileExt),
		Ext:       strings.ToLower(strings.TrimPrefix(fileExt, ".")),
		ConvertTo: dto.ConvertTo,
	}
}
