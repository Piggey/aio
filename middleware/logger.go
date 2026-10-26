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
			start := time.Now()
			sr := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			attrs := []slog.Attr{
				slog.String("method", r.Method),
				slog.String("path", r.URL.Path),
			}

			defer func() {
				attrs = append(attrs, slog.Duration("duration", time.Since(start)))
				attrs = append(attrs, slog.Int("status", sr.status))
			}()

			level := slog.LevelInfo
			err := next(sr, r)
			if err != nil {
				level = slog.LevelError
				attrs = append(attrs, slog.String("error", err.Error()))
			}

			logger.LogAttrs(r.Context(), level, "request", attrs...)
			return err
		}
	}
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (sr *statusRecorder) WriteHeader(status int) {
	sr.status = status
	sr.ResponseWriter.WriteHeader(status)
}

func (sr *statusRecorder) Unwrap() http.ResponseWriter {
	return sr.ResponseWriter
}
