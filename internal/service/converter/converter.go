package converter

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/chistyakoviv/converter/internal/config"
	"github.com/chistyakoviv/converter/internal/converter"
	"github.com/chistyakoviv/converter/internal/file"
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
	src, err := info.AbsoluteSourcePath()
	s.logger.Debug("convert", slog.String("src", src))
	if err != nil {
		return service.NewConverterError(err.Error(), service.ErrUnableToConvertFile)
	}

	if !file.Exists(src) {
		return service.NewConverterError(fmt.Sprintf("file '%s' does not exist", src), service.ErrFileDoesNotExist)
	}

	for _, entry := range info.ConvertTo {
		dest, err := info.AbsoluteDestinationPath(entry)
		if err != nil {
			return service.NewConverterError(err.Error(), service.ErrUnableToConvertFile)
		}
		var filetypeErr error
		var imageOk, videoOk bool
		if imageOk, filetypeErr = file.IsImage(info.Fullpath); imageOk {
			mergedConf := converter.MergeConfigs(s.imageConfigs[entry.Ext], entry.ConvConf)
			if err := s.imageConverter.Convert(src, dest, mergedConf); err != nil {
				return service.NewConverterError(err.Error(), service.ErrUnableToConvertFile)
			}
		}
		if filetypeErr != nil {
			return service.NewConverterError(filetypeErr.Error(), service.ErrInvalidConversionFormat)
		}
		if videoOk, filetypeErr = file.IsVideo(info.Fullpath); videoOk {
			mergedConf := converter.MergeConfigs(s.videoConfigs[entry.Ext], entry.ConvConf)
			if err := s.videoConverter.Convert(src, dest, mergedConf); err != nil {
				return service.NewConverterError(err.Error(), service.ErrUnableToConvertFile)
			}
		}
		if filetypeErr != nil {
			return service.NewConverterError(filetypeErr.Error(), service.ErrInvalidConversionFormat)
		}
		if !imageOk && !videoOk {
			return service.NewConverterError(fmt.Sprintf("the file is not an image or video: %s", info.Fullpath), service.ErrWrongSourceFile)
		}
	}
	return nil
}
