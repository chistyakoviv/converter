package ffmpeggo

import (
	"fmt"
	"log/slog"

	"github.com/chistyakoviv/converter/internal/config"
	"github.com/chistyakoviv/converter/internal/converter"
	"github.com/chistyakoviv/converter/internal/lib/slogger"
	ffmpeg "github.com/u2takey/ffmpeg-go"
)

type conf struct {
	cfg    *config.Config
	logger *slog.Logger
}

func NewVideoConverter(cfg *config.Config, logger *slog.Logger) converter.VideoConverter {
	return &conf{
		cfg:    cfg,
		logger: logger,
	}
}

func (c *conf) ToWebm(from string, to string, conf converter.ConversionConfig) error {
	const op = "ffmpeg-go.ToWebm"

	logger := c.logger.With(slog.String("op", op))

	// Build args
	args := ffmpeg.KwArgs{}

	for key, value := range conf {
		args[key] = value
	}

	args["threads"] = c.cfg.Video.Threads

	// Build and run the FFmpeg command
	err := ffmpeg.Input(from).
		Output(to, args).
		OverWriteOutput(). // Overwrite the output file if it already exists
		Run()

	if err != nil {
		logger.Error("failed to convert video", slog.String("from", from), slog.String("to", to), slogger.Err(err))
		return fmt.Errorf("failed to convert video: %w", err)
	}

	return nil
}

func (c *conf) Shutdown() {}
