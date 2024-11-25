package request

type DeletionRequest struct {
	Path string `json:"path" validate:"required"`
}
