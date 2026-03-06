package middleware

// The original work was derived from Goji's middleware, source:
// https://github.com/zenazn/goji/tree/master/web/middleware

import (
	"io"
	"log/slog"
	"net/http"
	"runtime/debug"

	"github.com/chistyakoviv/converter/internal/lib/slogger"
	"github.com/chistyakoviv/converter/internal/lib/stack_parser"
)

func NewRecoverer(
	panicWriter io.Writer,
	stackParser stack_parser.StackParser,
	logger *slog.Logger,
) func(next http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		fn := func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if rvr := recover(); rvr != nil {
					if rvr == http.ErrAbortHandler {
						// we don't recover http.ErrAbortHandler so the response
						// to the client is aborted, this should not be logged
						panic(rvr)
					}
					var writeErr error
					debugStack := debug.Stack()
					out, err := stackParser.Parse(debugStack, rvr)
					if err != nil {
						logger.Error("failed to parse stack", slogger.Err(err))
						_, writeErr = panicWriter.Write(debugStack)
					} else {
						_, writeErr = panicWriter.Write(out)
					}
					if writeErr != nil {
						logger.Error("failed to write stack", slogger.Err(err))
					}

					if r.Header.Get("Connection") != "Upgrade" {
						w.WriteHeader(http.StatusInternalServerError)
					}
				}
			}()

			next.ServeHTTP(w, r)
		}

		return http.HandlerFunc(fn)
	}
}
