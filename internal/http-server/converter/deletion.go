package converter

import (
	"github.com/chistyakoviv/converter/internal/http-server/request"
	"github.com/chistyakoviv/converter/internal/model"
)

func ToDeletionInfoFromRequest(dto request.DeletionRequest) *model.DeletionInfo {
	return &model.DeletionInfo{
		Fullpath: dto.Path,
	}
}
