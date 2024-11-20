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
	"github.com/chistyakoviv/converter/internal/service"
)

type serv struct {
	cfg            *config.Config
	logger         *slog.Logger
	imageConverter converter.ImageConverter
	videoConverter converter.VideoConverter
	imageConfigs   map[string]converter.ConversionConfig
	videoConfigs   map[string]converter.ConversionConfig
}

func NewService(
	cfg *config.Config,
	logger *slog.Logger,
	imageConverter converter.ImageConverter,
	videoConverter converter.VideoConverter,
) (service.ConverterService, error) {
	imageConfigs := make(map[string]converter.ConversionConfig)
	videoConfigs := make(map[string]converter.ConversionConfig)
	for _, entry := range cfg.Image.DefaultFormats {
		imageConfigs[entry.Ext] = entry.ConvConf
	}
	for _, entry := range cfg.Video.DefaultFormats {
		videoConfigs[entry.Ext] = entry.ConvConf
	}
	return &serv{
		cfg:            cfg,
		logger:         logger,
		imageConverter: imageConverter,
		videoConverter: videoConverter,
		imageConfigs:   imageConfigs,
		videoConfigs:   videoConfigs,
	}, nil
}

func (s *serv) Convert(ctx context.Context, info *model.Conversion) error {
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	src := fmt.Sprintf("%s%s", wd, info.Fullpath)
	if !file.Exists(src) {
		return NewConversionError(fmt.Sprintf("file '%s' does not exist", info.Fullpath), ErrFileDoesNotExist)
	}

	destPrefix := fmt.Sprintf("%s%s/%s", wd, info.Path, info.Filestem)
	for _, entry := range info.ConvertTo {
		dest := destPrefix
		boolOk := false
		replaceOrigExt := false
		if value, ok := entry.Optional["replace_orig_ext"]; ok {
			if replaceOrigExt, boolOk = value.(bool); boolOk {
				replaceOrigExt = true
			}
		}
		if !replaceOrigExt {
			dest = fmt.Sprintf("%s.%s", dest, info.Ext)
		}
		dest = fmt.Sprintf("%s.%s", dest, entry.Ext)
		switch entry.Ext {
		case "webp":
			mergedConf := converter.MergeConfigs(s.imageConfigs[entry.Ext], entry.ConvConf)
			if err := s.imageConverter.ToWebp(src, dest, mergedConf); err != nil {
				return NewConversionError(err.Error(), ErrConversion)
			}
		case "webm":
			mergedConf := converter.MergeConfigs(s.videoConfigs[entry.Ext], entry.ConvConf)
			if err := s.videoConverter.ToWebm(src, dest, mergedConf); err != nil {
				return NewConversionError(err.Error(), ErrConversion)
			}
		default:
			return NewConversionError(fmt.Sprintf("conversion to '%s' is not implemented", entry.Ext), ErrInvalidConversionFormat)
		}
	}
	return nil
}
