package conversion

import (
	"context"
	"fmt"
	"strings"

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
	if !isSupported(info.Ext) {
		return -1, fmt.Errorf("file type '%s' not supported", info.Ext)
	}

	// Assign default format if no target formats are specified
	if info.ConvertTo == nil {
		defaultFormat, err := defaultFormatFor(info.Ext)
		if err != nil {
			return -1, fmt.Errorf("failed to get default format: %w", err)
		}
		info.ConvertTo = []string{defaultFormat}
	} else {
		var unsupportedFormats []string
		for _, ext := range info.ConvertTo {
			if !isConvertible(info.Ext, ext) {
				unsupportedFormats = append(unsupportedFormats, fmt.Sprintf("'%s'", ext))
			}
		}
		if len(unsupportedFormats) > 0 {
			return -1, fmt.Errorf("file type '%s' is not convertible to %s", info.Ext, strings.Join(unsupportedFormats, ", "))
		}
	}

	return s.conversionRepository.Create(ctx, info)
}

func (s *serv) Delete(ctx context.Context, fullpath string) error {
	return nil
}
