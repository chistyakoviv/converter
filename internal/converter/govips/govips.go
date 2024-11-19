package govips

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/chistyakoviv/converter/internal/config"
	"github.com/chistyakoviv/converter/internal/converter"
	"github.com/chistyakoviv/converter/internal/lib/slogger"
	"github.com/davidbyttow/govips/v2/vips"
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
	vips.Startup(nil)

	return &conv{
		logger: logger,
	}
}

func (c *conv) ToWebp(from string, to string, conf converter.ConversionConfig) error {
	const op = "govips.ToWebp"

	logger := c.logger.With(slog.String("op", op))

	image1, err := vips.NewImageFromFile(from)
	if err != nil {
		logger.Error("error:", slogger.Err(err))
		return wrapError(err)
	}

	ep := vips.NewDefaultWEBPExportParams()
	image1bytes, _, err := image1.Export(ep)
	if err != nil {
		logger.Error("error:", slogger.Err(err))
		return wrapError(err)
	}

	err = os.WriteFile(to, image1bytes, 0644)
	if err != nil {
		logger.Error("error:", slogger.Err(err))
		return wrapError(err)
	}
	return nil
}

func (c *conv) Shutdown() {
	vips.Shutdown()
}

func wrapError(err error) error {
	return fmt.Errorf("govips: %w", err)
}
