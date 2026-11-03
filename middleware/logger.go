package middleware

import (
	"errors"
	"log/slog"
	"net/http"
	"time"

	"github.com/Piggey/aio"
)

// Logger logs every incoming request with its status code using logger (slog.Default() when nil).
// Requests with status >= 400 are logged at ERROR level, the rest at INFO.
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
			sr := &statusRecorder{ResponseWriter: w, status: http.StatusOK}

			err := next(sr, r)
			duration := time.Since(start)
			status := responseStatus(sr.status, err)

			attrs = append(attrs, slog.Int("status", status), slog.Duration("duration", duration))
			if err != nil {
				attrs = append(attrs, slog.String("err", err.Error()))
			}

			level := statusLevel(status)
			logger.LogAttrs(r.Context(), level, "request", attrs...)

			return err
		}
	}
}

// responseStatus returns the status code the client gets.
// The error response is written by App.ErrorHandler after the whole chain, so sr never sees its status.
func responseStatus(written int, err error) int {
	if err == nil {
		return written
	}

	pd, ok := errors.AsType[aio.ProblemDetails](err)
	if !ok {
		return http.StatusInternalServerError
	}

	return pd.Status
}

func statusLevel(status int) slog.Level {
	if status >= 400 {
		return slog.LevelError
	}

	return slog.LevelInfo
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
