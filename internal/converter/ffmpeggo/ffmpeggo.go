package ffmpeggo

import (
	"fmt"
	"log/slog"
	"os"

	"github.com/chistyakoviv/converter/internal/config"
	"github.com/chistyakoviv/converter/internal/converter"
	"github.com/chistyakoviv/converter/internal/file"
	"github.com/chistyakoviv/converter/internal/lib/slogger"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

type conv struct {
	cfg    *config.Config
	logger *slog.Logger
}

func NewVideoConverter(cfg *config.Config, logger *slog.Logger) converter.VideoConverter {
	return &conv{
		cfg:    cfg,
		logger: logger,
	}
}

func (c *conv) Convert(from string, to string, conf converter.ConversionConfig) error {
	const op = "ffmpeg-go.Convert"

	logger := c.logger.With(slog.String("op", op))

	// Build args
	args := ffmpeg.KwArgs{}

	for key, value := range conf {
		args[key] = value
	}

	args["threads"] = c.cfg.Video.Threads

	tmpFile := file.ToTmpFilePath(to)

	// Build and run the FFmpeg command
	err := ffmpeg.Input(from).
		Output(tmpFile, args).
		OverWriteOutput(). // Overwrite the output file if it already exists
		Run()

	if err != nil {
		logger.Error("failed to convert video", slog.String("from", from), slog.String("to", to), slogger.Err(err))
		return fmt.Errorf("failed to convert video: %w", err)
	}

	if err := os.Remove(to); err != nil && !os.IsNotExist(err) {
		logger.Error("failed to remove old file", slog.String("path", to), slogger.Err(err))
		return fmt.Errorf("failed to remove old file '%s': %w", to, err)
	}

	if err := os.Rename(tmpFile, to); err != nil {
		logger.Error("failed to rename tmp file", slog.String("from", tmpFile), slog.String("to", to), slogger.Err(err))
		return fmt.Errorf("failed to rename tmp file '%s' to '%s': %w", tmpFile, to, err)
	}

	return nil
}

func (c *conv) Shutdown() {}
