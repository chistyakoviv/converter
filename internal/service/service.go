package service

import (
	"context"

	"github.com/chistyakoviv/converter/internal/model"
)

type ConversionService interface {
	Convert(ctx context.Context, info *model.ConversionInfo) (int64, error)
	Delete(ctx context.Context, fullpath string) error
}
