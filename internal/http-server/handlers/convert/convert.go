package convert

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/chistyakoviv/converter/internal/http-server/converter"
	"github.com/chistyakoviv/converter/internal/http-server/deps"
	"github.com/chistyakoviv/converter/internal/http-server/response"
	resp "github.com/chistyakoviv/converter/internal/lib/http/response"
	"github.com/chistyakoviv/converter/internal/lib/slogger"
	"github.com/chistyakoviv/converter/internal/repository/conversion"
	"github.com/go-chi/render"
)

func New(
	d *deps.ConversionDeps,
	hander http.HandlerFunc,
) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id, err := d.ConversionService.Add(d.Ctx, converter.ToConversionInfoFromRequest(d.Request))
		if errors.Is(err, conversion.ErrPathAlreadyExist) {
			d.Logger.Debug("path already exists", slog.String("path", d.Request.Path))

			render.JSON(w, r, resp.Error("path already exists"))

			return
		}
		if err != nil {
			d.Logger.Error("failed to add file to conversion queue", slogger.Err(err))

			render.JSON(w, r, resp.Error("failed to add file to conversion queue"))

			return
		}

		d.Logger.Debug("file added", slog.Int64("id", id))

		render.JSON(w, r, response.ConversionResponse{
			Response: resp.OK(),
			Id:       id,
		})
	}
}
