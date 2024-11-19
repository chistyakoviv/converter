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

	src := fmt.Sprintf("%s/%s", wd, info.Fullpath)
	if !file.Exists(src) {
		return NewConversionError(fmt.Sprintf("file '%s' does not exist", info.Fullpath), ErrFileDoesNotExist)
	}

	destPrefix := fmt.Sprintf("%s%s/%s", wd, info.Path, info.Filestem)
	if !info.ReplaceOrigExt {
		destPrefix = fmt.Sprintf("%s.%s", destPrefix, info.Ext)
	}
	for _, ext := range info.ConvertTo {
		dest := fmt.Sprintf("%s.%s", destPrefix, ext)
		switch ext {
		case "webp":
			if err := s.imageConverter.ToWebp(src, dest); err != nil {
				return NewConversionError(err.Error(), ErrConversion)
			}
		default:
			return NewConversionError(fmt.Sprintf("conversion to '%s' is not implemented", ext), ErrInvalidConversionFormat)
		}
	}
	return nil
}
