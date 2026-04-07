package converter

import (
	"context"

	"github.com/chistyakoviv/converter/internal/model"
)

type ImageConverter interface {
	Shutdowner
	Convert(from string, to string, conf ConversionConfig) error
}

type VideoConverter interface {
	Shutdowner
	Convert(from string, to string, conf ConversionConfig) error
}

type Shutdowner interface {
	Shutdown()
}

type Converter interface {
	Convert(ctx context.Context, info *model.Conversion) error
}
