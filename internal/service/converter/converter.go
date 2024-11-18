package converter

import (
	"context"
	"fmt"
	"log/slog"
	"os"

	"github.com/chistyakoviv/converter/internal/config"
	"github.com/chistyakoviv/converter/internal/converter"
	"github.com/chistyakoviv/converter/internal/lib/file"
	"github.com/chistyakoviv/converter/internal/model"
)

type serv struct {
	cfg            *config.Config
	logger         *slog.Logger
	imageConverter converter.ImageConverter
}

func NewService(cfg *config.Config, logger *slog.Logger, imageConverter converter.ImageConverter) *serv {
	return &serv{
		cfg:            cfg,
		logger:         logger,
		imageConverter: imageConverter,
	}
}

func (s *serv) Convert(ctx context.Context, info *model.Conversion) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	absolutePath := fmt.Sprintf("%s/%s", wd, info.Fullpath)
	if !file.Exists(absolutePath) {
		return NewConversionError(fmt.Sprintf("file '%s' does not exist", info.Fullpath), ErrFileDoesNotExist)
	}

	path := fmt.Sprintf("%s%s/%s", wd, info.Path, info.Filestem)
	if !info.ReplaceOrigExt {
		path = fmt.Sprintf("%s.%s", path, info.Ext)
	}
	for _, ext := range info.ConvertTo {
		output := fmt.Sprintf("%s.%s", path, ext)
		switch ext {
		case "webp":
			if err := s.imageConverter.ToWebp(absolutePath, output); err != nil {
				return NewConversionError(err.Error(), ErrConversion)
			}
		default:
			return NewConversionError(fmt.Sprintf("conversion to '%s' is not implemented", ext), ErrInvalidConversionFormat)
		}
	}
	return nil
}
