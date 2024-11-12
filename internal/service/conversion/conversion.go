package conversion

import (
	"context"

	"github.com/chistyakoviv/converter/internal/model"
	"github.com/chistyakoviv/converter/internal/repository"
	"github.com/chistyakoviv/converter/internal/service"
)

type serv struct {
	conversionRepository repository.ConversionQueueRepository
}

func NewService(
	conversionRepository repository.ConversionQueueRepository,
) service.ConversionService {
	return &serv{
		conversionRepository: conversionRepository,
	}
}

func (s *serv) Convert(ctx context.Context, info *model.ConversionInfo) (int64, error) {
	return s.conversionRepository.Create(ctx, info)
}

func (s *serv) Delete(ctx context.Context, fullpath string) error {
	return nil
}
