package repository

import (
	"context"

	"github.com/chistyakoviv/converter/internal/model"
)

type ConversionQueueRepository interface {
	Create(ctx context.Context, file *model.Conversion) (int64, error)
	GetByFullpath(ctx context.Context, fullpath string) (*model.Conversion, error)
}
