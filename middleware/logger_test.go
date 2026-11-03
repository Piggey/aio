package middleware

import (
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Piggey/aio"
)

func TestResponseStatus(t *testing.T) {
	tests := []struct {
		name    string
		written int
		err     error
		want    int
	}{
		{
			name:    "bez bledu zwraca zapisany status",
			written: http.StatusCreated,
			want:    http.StatusCreated,
		},
		{
			name:    "bez bledu zwraca zapisany status bledu",
			written: http.StatusNotFound,
			want:    http.StatusNotFound,
		},
		{
			name:    "ProblemDetails wygrywa z zapisanym statusem",
			written: http.StatusOK,
			err:     aio.ErrConflict,
			want:    http.StatusConflict,
		},
		{
			name:    "opakowany ProblemDetails",
			written: http.StatusOK,
			err:     fmt.Errorf("serwis: %w", aio.ErrNotFound),
			want:    http.StatusNotFound,
		},
		{
			name:    "nieznany blad daje 500",
			written: http.StatusOK,
			err:     errors.New("boom"),
			want:    http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, responseStatus(tt.written, tt.err))
		})
	}
}

func TestStatusLevel(t *testing.T) {
	tests := []struct {
		name   string
		status int
		want   slog.Level
	}{
		{name: "200", status: http.StatusOK, want: slog.LevelInfo},
		{name: "304", status: http.StatusNotModified, want: slog.LevelInfo},
		{name: "399", status: 399, want: slog.LevelInfo},
		{name: "400", status: http.StatusBadRequest, want: slog.LevelError},
		{name: "500", status: http.StatusInternalServerError, want: slog.LevelError},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, statusLevel(tt.status))
		})
	}
}
