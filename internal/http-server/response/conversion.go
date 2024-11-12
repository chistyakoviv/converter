package response

import (
	resp "github.com/chistyakoviv/converter/internal/lib/http/response"
)

type ConversionResponse struct {
	resp.Response
	Id int64 `json:"id"`
}
