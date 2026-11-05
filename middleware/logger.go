package middleware

import (
	"log/slog"
	"net/http"
	"time"

	"github.com/Piggey/aio"
)

// Logger logs every incoming request using logger (slog.Default() when nil).
// Requests that end with an error are logged at ERROR level, the rest at INFO.
func Logger(logger *slog.Logger) aio.Middleware {
	if logger == nil {
		logger = slog.Default()
	}

	return func(next aio.HandlerFunc) aio.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) error {
			attrs := []slog.Attr{
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
			}

			start := time.Now()

			err := next(w, r)
			duration := time.Since(start)
			attrs = append(attrs, slog.Duration("duration", duration))
			if err != nil {
				attrs = append(attrs, slog.String("err", err.Error()))
				logger.LogAttrs(r.Context(), slog.LevelError, "request", attrs...)
				return err
			}

			logger.LogAttrs(r.Context(), slog.LevelInfo, "request", attrs...)
			return err
		}
	}
}
