package request

import "github.com/chistyakoviv/converter/internal/model"

type ConversionRequest struct {
	Path      string            `json:"path" validate:"required"`
	ConvertTo []model.ConvertTo `json:"convert_to,omitempty"`
}
