package converter

import (
	"github.com/chistyakoviv/converter/internal/file"
	"github.com/chistyakoviv/converter/internal/http-server/request"
	"github.com/chistyakoviv/converter/internal/model"
)

func ToConversionInfoFromRequest(dto request.ConversionRequest) *model.ConversionInfo {
	finfo := file.ExtractInfo(dto.Path)
	cinfo := model.ToConversionInfoFromFileInfo(finfo)
	cinfo.ConvertTo = dto.ConvertTo
	return cinfo
}
