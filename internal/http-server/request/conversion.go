package request

type ConversionRequest struct {
	Path           string   `json:"path" validate:"required"`
	ConvertTo      []string `json:"convert_to,omitempty"`
	ReplaceOrigExt bool     `json:"replace_orig_ext,omitempty"`
}
