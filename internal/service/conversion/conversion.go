package conversion

import (
	"context"
	"fmt"

	"github.com/chistyakoviv/converter/internal/lib/conversion"
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

func (s *serv) Add(ctx context.Context, info *model.ConversionInfo) (int64, error) {
	if !conversion.IsSupported(info.Ext) {
		return -1, fmt.Errorf("file type \"%s\" not supported", info.Ext)
	}
	if info.ConvertTo == nil {
		defaultFormat, err := conversion.Default(info.Ext)
		if err != nil {
			return -1, fmt.Errorf("failed to get default format: %w", err)
		}
		info.ConvertTo = []string{defaultFormat}
	} else {
		for _, ext := range info.ConvertTo {
			if !conversion.IsConvertable(info.Ext, ext) {
				return -1, fmt.Errorf("file type \"%s\" is not convertable to \"%s\"", info.Ext, ext)
			}
		}
	}
	return s.conversionRepository.Create(ctx, info)
}

func (s *serv) Delete(ctx context.Context, fullpath string) error {
	return nil
}
