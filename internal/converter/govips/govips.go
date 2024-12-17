package govips

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/chistyakoviv/converter/internal/config"
	"github.com/chistyakoviv/converter/internal/converter"
	"github.com/chistyakoviv/converter/internal/file"
	"github.com/chistyakoviv/converter/internal/lib/mapper"
	"github.com/chistyakoviv/converter/internal/lib/slogger"
	"github.com/davidbyttow/govips/v2/vips"
)

const (
	filePermissions = 0644
)

type conv struct {
	logger *slog.Logger
}

func NewImageConverter(logger *slog.Logger, cfg *config.Config) converter.ImageConverter {
	vipsLogger := func(messageDomain string, verbosity vips.LogLevel, message string) {
		var messageLevelDescription string
		var loggerFn func(msg string, args ...any)
		switch verbosity {
		case vips.LogLevelError:
			messageLevelDescription = "error"
			loggerFn = logger.Error
		case vips.LogLevelCritical:
			messageLevelDescription = "critical"
			loggerFn = logger.Error
		case vips.LogLevelWarning:
			messageLevelDescription = "warning"
			loggerFn = logger.Warn
		case vips.LogLevelMessage:
			messageLevelDescription = "message"
			loggerFn = logger.Info
		case vips.LogLevelInfo:
			messageLevelDescription = "info"
			loggerFn = logger.Info
		case vips.LogLevelDebug:
			messageLevelDescription = "debug"
			loggerFn = logger.Debug
		}

		loggerFn("govips",
			slog.Attr{
				Key:   "domain",
				Value: slog.StringValue(messageDomain),
			},
			slog.Attr{
				Key:   "level",
				Value: slog.StringValue(messageLevelDescription),
			},
			slog.Attr{
				Key:   "message",
				Value: slog.StringValue(message),
			})
	}

	var logLevel vips.LogLevel
	switch cfg.Env {
	case config.EnvProd:
		logLevel = vips.LogLevelWarning
	default:
		logLevel = vips.LogLevelDebug
	}

	vips.LoggingSettings(vipsLogger, logLevel)
	// See config example https://github.com/davidbyttow/govips/blob/master/examples/image/bench_test.go
	// Set default values as defined in the source at https://github.com/davidbyttow/govips/blob/4188b98393e95247efbae5011a9b8b98fbcd329d/vips/govips.go#L17
	conf := &vips.Config{
		ConcurrencyLevel: 1,
		MaxCacheFiles:    0,
		MaxCacheMem:      50 * 1024 * 1024,
		MaxCacheSize:     100,
		// ReportLeaks:      false,
		// CacheTrace:       false,
		// CollectStats:     false,
	}
	if cfg.Image.Threads > 0 {
		conf.ConcurrencyLevel = cfg.Image.Threads
	}
	vips.Startup(conf)

	return &conv{
		logger: logger,
	}
}

func (c *conv) Convert(from string, to string, conf converter.ConversionConfig) error {
	const op = "govips.Convert"

	logger := c.logger.With(slog.String("op", op))
	ext := file.Ext(to)

	toTmp := file.ToTmpFilePath(to)

	var err error
	switch ext {
	case "jpg", "jpeg":
		err = c.toJpeg(from, toTmp, conf)
	case "png":
		err = c.toPng(from, toTmp, conf)
	case "webp":
		err = c.toWebp(from, toTmp, conf)
	case "avif":
		err = c.toAvif(from, toTmp, conf)
	default:
		return wrapError(fmt.Errorf("unsupported format: %s", ext))
	}

	if err != nil {
		logger.Error("error:", slogger.Err(err))
		return wrapError(err)
	}

	if err = os.Remove(to); err != nil && !os.IsNotExist(err) {
		return wrapError(fmt.Errorf("failed to remove old file: %w", err))
	}

	if err = os.Rename(toTmp, to); err != nil {
		return wrapError(fmt.Errorf("failed to rename tmp file: %w", err))
	}

	return nil
}

func (c *conv) toJpeg(from string, to string, conf converter.ConversionConfig) error {
	image, err := vips.NewImageFromFile(from)
	if err != nil {
		return err
	}

	ep := vips.NewJpegExportParams()

	mapper.MapToStruct(conf, ep)

	imageBytes, _, err := image.ExportJpeg(ep)
	if err != nil {
		return err
	}

	err = os.WriteFile(to, imageBytes, filePermissions)
	if err != nil {
		return err
	}

	return nil
}

func (c *conv) toPng(from string, to string, conf converter.ConversionConfig) error {
	image, err := vips.NewImageFromFile(from)
	if err != nil {
		return err
	}

	ep := vips.NewPngExportParams()

	mapper.MapToStruct(conf, ep)

	imageBytes, _, err := image.ExportPng(ep)
	if err != nil {
		return err
	}

	err = os.WriteFile(to, imageBytes, filePermissions)
	if err != nil {
		return err
	}

	return nil
}

func (c *conv) toWebp(from string, to string, conf converter.ConversionConfig) error {
	image, err := vips.NewImageFromFile(from)
	if err != nil {
		return err
	}

	ep := vips.NewWebpExportParams()

	mapper.MapToStruct(conf, ep)

	imageBytes, _, err := image.ExportWebp(ep)
	if err != nil {
		return err
	}

	err = os.WriteFile(to, imageBytes, filePermissions)
	if err != nil {
		return err
	}

	return nil
}

func (c *conv) toAvif(from string, to string, conf converter.ConversionConfig) error {
	image, err := vips.NewImageFromFile(from)
	if err != nil {
		return err
	}

	ep := vips.NewAvifExportParams()

	mapper.MapToStruct(conf, ep)

	imageBytes, _, err := image.ExportAvif(ep)
	if err != nil {
		return err
	}

	err = os.WriteFile(to, imageBytes, filePermissions)
	if err != nil {
		return err
	}

	return nil
}

func (c *conv) Shutdown() {
	vips.Shutdown()
}

func wrapError(err error) error {
	return fmt.Errorf("govips: %w", err)
}
