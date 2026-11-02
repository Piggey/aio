package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsAbortHandler(t *testing.T) {
	tests := []struct {
		name string
		rec  any
		want bool
	}{
		{
			name: "http.ErrAbortHandler",
			rec:  http.ErrAbortHandler,
			want: true,
		},
		{
			name: "opakowany http.ErrAbortHandler",
			rec:  fmt.Errorf("handler: %w", http.ErrAbortHandler),
			want: true,
		},
		{
			name: "inny blad",
			rec:  errors.New("boom"),
			want: false,
		},
		{
			name: "string z trescia ErrAbortHandler",
			rec:  http.ErrAbortHandler.Error(),
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isAbortHandler(tt.rec))
		})
	}
}

func TestPanicError_Error(t *testing.T) {
	tests := []struct {
		name  string
		value any
		want  string
	}{
		{
			name:  "string",
			value: "boom",
			want:  "panic: boom",
		},
		{
			name:  "blad",
			value: errors.New("boom"),
			want:  "panic: boom",
		},
		{
			name:  "liczba",
			value: 42,
			want:  "panic: 42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, PanicError{Value: tt.value}.Error())
		})
	}
}
