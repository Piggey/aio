package middleware

import (
	"bytes"
	"encoding/json"
	"errors"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Piggey/aio"
)

func TestLogger(t *testing.T) {
	tests := []struct {
		name       string
		handler    aio.HandlerFunc
		wantLevel  string
		wantStatus int
		wantError  string
	}{
		{
			name: "udane zadanie logowane jako INFO",
			handler: func(w http.ResponseWriter, r *http.Request) error {
				return nil
			},
			wantLevel:  "INFO",
			wantStatus: http.StatusOK,
		},
		{
			name: "status ustawiony przez handler jest logowany",
			handler: func(w http.ResponseWriter, r *http.Request) error {
				w.WriteHeader(http.StatusCreated)
				return nil
			},
			wantLevel:  "INFO",
			wantStatus: http.StatusCreated,
		},
		{
			name: "ProblemDetails logowany jako ERROR ze swoim statusem",
			handler: func(w http.ResponseWriter, r *http.Request) error {
				return aio.ErrNotFound
			},
			wantLevel:  "ERROR",
			wantStatus: http.StatusNotFound,
			wantError:  "Not Found",
		},
		{
			name: "nieznany blad logowany jako ERROR ze statusem 500",
			handler: func(w http.ResponseWriter, r *http.Request) error {
				return errors.New("boom")
			},
			wantLevel:  "ERROR",
			wantStatus: http.StatusInternalServerError,
			wantError:  "boom",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			buf := bytes.Buffer{}
			logger := slog.New(slog.NewJSONHandler(&buf, nil))

			req := httptest.NewRequest(http.MethodGet, "/test", nil)
			err := Logger(logger)(tt.handler)(httptest.NewRecorder(), req)
			if tt.wantError != "" {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}

			var entry map[string]any
			err = json.Unmarshal(buf.Bytes(), &entry)
			assert.NoError(t, err)
			assert.Equal(t, tt.wantLevel, entry["level"])
			assert.Equal(t, http.MethodGet, entry["method"])
			assert.Equal(t, "/test", entry["path"])
			assert.EqualValues(t, tt.wantStatus, entry["status"])
			if tt.wantError != "" {
				assert.Equal(t, tt.wantError, entry["error"])
			} else {
				assert.NotContains(t, entry, "error")
			}
		})
	}
}
