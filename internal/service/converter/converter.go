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
	cfg                 *config.Config
	logger              *slog.Logger
	imageConverter      converter.ImageConverter
	videoConverter      converter.VideoConverter
	defaultImageFormats converter.ConversionFormats
	defaultVideoFormats converter.ConversionFormats
}

func NewService(
	cfg *config.Config,
	logger *slog.Logger,
	imageConverter converter.ImageConverter,
	videoConverter converter.VideoConverter,
) (service.ConverterService, error) {
	defaultImageFormats, err := converter.ParseFormats(cfg.Image.DefaultFormats)
	if err != nil {
		return nil, fmt.Errorf("failed to parse default image formats: %w", err)
	}
	defaultVideoFormats, err := converter.ParseFormats(cfg.Video.DefaultFormats)
	if err != nil {
		return nil, fmt.Errorf("failed to parse default video formats: %w", err)
	}

	return &serv{
		cfg:                 cfg,
		logger:              logger,
		imageConverter:      imageConverter,
		videoConverter:      videoConverter,
		defaultImageFormats: defaultImageFormats,
		defaultVideoFormats: defaultVideoFormats,
	}, nil
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
	for _, entry := range info.ConvertTo {
		ext, conf, err := converter.ParseFormat(entry)
		if err != nil {
			return NewConversionError(err.Error(), ErrInvalidConversionFormat)
		}
		dest := fmt.Sprintf("%s.%s", destPrefix, ext)
		switch ext {
		case "webp":
			mergedConf := converter.MergeConfigs(s.defaultImageFormats[ext], conf)
			if err := s.imageConverter.ToWebp(src, dest, mergedConf); err != nil {
				return NewConversionError(err.Error(), ErrConversion)
			}
		case "mp4":
			mergedConf := converter.MergeConfigs(s.defaultVideoFormats[ext], conf)
			if err := s.videoConverter.ToWebm(src, dest, mergedConf); err != nil {
				return NewConversionError(err.Error(), ErrConversion)
			}
		default:
			return NewConversionError(fmt.Sprintf("conversion to '%s' is not implemented", ext), ErrInvalidConversionFormat)
		}
	}
	return nil
}
