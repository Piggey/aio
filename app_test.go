package aio

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestProblemDetailsFor(t *testing.T) {
	errDomain := errors.New("domain error")
	conflict := Error(http.StatusConflict, "conflict", errDomain)

	tests := []struct {
		name string
		err  error
		want ProblemDetails
	}{
		{
			name: "ProblemDetails jest zwracany bez zmian",
			err:  ErrNotFound,
			want: ProblemDetails{Title: "Not Found", Status: http.StatusNotFound},
		},
		{
			name: "ProblemDetails opakowany przez fmt.Errorf jest odnajdywany",
			err:  fmt.Errorf("app.binder.BindIn: %w", ErrBadRequest),
			want: ProblemDetails{Title: "Bad Request", Status: http.StatusBadRequest},
		},
		{
			name: "ProblemDetails w errors.Join jest odnajdywany",
			err:  errors.Join(errDomain, conflict),
			want: conflict,
		},
		{
			name: "nieznany blad daje 500",
			err:  errDomain,
			want: ProblemDetails{Title: "Internal Server Error", Status: http.StatusInternalServerError},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, problemDetailsFor(tt.err))
		})
	}
}
