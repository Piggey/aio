package aio

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

type testIn struct {
	Name string `json:"name" validate:"required"`
}

type testOut struct {
	Name string `json:"name"`
}

func errorMiddleware(err error) Middleware {
	return func(next HandlerFunc) HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) error {
			return err
		}
	}
}

func TestErrorHandling(t *testing.T) {
	errDomain := errors.New("domain error")

	tests := []struct {
		name        string
		body        string
		handlerErr  error
		middlewares []Middleware
		wantStatus  int
	}{
		{
			name:       "poprawne zadanie zwraca 200",
			body:       `{"name":"x"}`,
			wantStatus: http.StatusOK,
		},
		{
			name:       "blad bindowania zwraca 400",
			body:       `{"name":`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "blad walidacji zwraca 400",
			body:       `{"name":""}`,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "ProblemDetails z handlera jest zachowany",
			body:       `{"name":"x"}`,
			handlerErr: ErrBadRequest,
			wantStatus: http.StatusBadRequest,
		},
		{
			name:       "opakowany ProblemDetails z handlera jest zachowany",
			body:       `{"name":"x"}`,
			handlerErr: errors.Join(errDomain, newProblemDetails(http.StatusConflict)),
			wantStatus: http.StatusConflict,
		},
		{
			name:       "nieznany blad z handlera zwraca 500",
			body:       `{"name":"x"}`,
			handlerErr: errDomain,
			wantStatus: http.StatusInternalServerError,
		},
		{
			name:        "blad z middleware'a jest obslugiwany przez ErrorHandler",
			body:        `{"name":"x"}`,
			middlewares: []Middleware{errorMiddleware(newProblemDetails(http.StatusUnauthorized))},
			wantStatus:  http.StatusUnauthorized,
		},
		{
			name: "blad z handlera przechodzi przez middleware",
			body: `{"name":"x"}`,
			middlewares: []Middleware{func(next HandlerFunc) HandlerFunc {
				return next
			}},
			handlerErr: newProblemDetails(http.StatusForbidden),
			wantStatus: http.StatusForbidden,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			app := New("test", "v1")
			app.Post("/test", func(ctx context.Context, in testIn) (testOut, error) {
				return testOut{Name: in.Name}, tt.handlerErr
			}, WithMiddlewares(tt.middlewares...))

			req := httptest.NewRequest(http.MethodPost, "/test", strings.NewReader(tt.body))
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			app.intoMux().ServeHTTP(rec, req)

			assert.Equal(t, tt.wantStatus, rec.Code)
			if tt.wantStatus != http.StatusOK {
				var pd ProblemDetails
				err := json.NewDecoder(rec.Body).Decode(&pd)
				assert.NoError(t, err)
				assert.Equal(t, tt.wantStatus, pd.Status)
			}
		})
	}
}
