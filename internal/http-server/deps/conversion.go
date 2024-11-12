package deps

import (
	"context"
	"log/slog"

	"github.com/go-playground/validator/v10"
	"github.com/chistyakoviv/converter/internal/service"
	"github.com/chistyakoviv/converter/internal/http-server/request"
)

type ConversionDeps struct {
	Ctx               context.Context
	Logger            *slog.Logger
	Validator         *validator.Validate
	ConversionService service.ConversionService
	Request           request.ConversionRequest
}