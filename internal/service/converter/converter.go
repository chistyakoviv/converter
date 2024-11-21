package converter

import (
	"context"
	"fmt"
	"log/slog"
	"os"

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
	wd, err := os.Getwd()
	if err != nil {
		return err
	}

	if !file.Exists(info.Fullpath) {
		return NewConversionError(fmt.Sprintf("file '%s' does not exist", info.Fullpath), ErrFileDoesNotExist)
	}

	src := fmt.Sprintf("%s%s", wd, info.Fullpath)
	destPrefix := fmt.Sprintf("%s%s/%s", wd, info.Path, info.Filestem)
	for _, entry := range info.ConvertTo {
		dest := destPrefix
		var isReplaceOrigExtBool bool
		var replaceOrigExt bool
		if value, ok := entry.Optional["replace_orig_ext"]; ok {
			if replaceOrigExt, isReplaceOrigExtBool = value.(bool); isReplaceOrigExtBool {
				replaceOrigExt = true
			}
		}
		if !replaceOrigExt {
			dest = fmt.Sprintf("%s.%s", dest, info.Ext)
		}
		dest = fmt.Sprintf("%s.%s", dest, entry.Ext)
		var filetypeErr error
		var imageOk bool
		var videoOk bool
		if imageOk, filetypeErr = file.IsImage(info.Fullpath); imageOk {
			mergedConf := converter.MergeConfigs(s.imageConfigs[entry.Ext], entry.ConvConf)
			if err := s.imageConverter.Convert(src, dest, mergedConf); err != nil {
				return NewConversionError(err.Error(), ErrUnableToConvertFile)
			}
		}
		if filetypeErr != nil {
			return NewConversionError(filetypeErr.Error(), ErrInvalidConversionFormat)
		}
		if videoOk, filetypeErr = file.IsVideo(info.Fullpath); videoOk {
			mergedConf := converter.MergeConfigs(s.videoConfigs[entry.Ext], entry.ConvConf)
			if err := s.videoConverter.Convert(src, dest, mergedConf); err != nil {
				return NewConversionError(err.Error(), ErrUnableToConvertFile)
			}
		}
		if filetypeErr != nil {
			return NewConversionError(filetypeErr.Error(), ErrInvalidConversionFormat)
		}
		if !imageOk && !videoOk {
			return NewConversionError(fmt.Sprintf("the file is not an image or video: %s", info.Fullpath), ErrWrongSourceFile)
		}
	}
	return nil
}
