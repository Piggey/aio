package aio

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToHTTPMiddleware(t *testing.T) {
	errHandler := errors.New("handler error")
	errMiddleware := errors.New("middleware error")

	passThrough := func(next HandlerFunc) HandlerFunc {
		return next
	}

	tests := []struct {
		name        string
		handlerErr  error
		middleware  Middleware
		wantSeenErr error
		wantErr     error
	}{
		{
			name:        "brak bledu",
			middleware:  passThrough,
			wantSeenErr: nil,
			wantErr:     nil,
		},
		{
			name:        "middleware widzi blad handlera i go przekazuje",
			handlerErr:  errHandler,
			middleware:  passThrough,
			wantSeenErr: errHandler,
			wantErr:     errHandler,
		},
		{
			name:       "middleware podmienia blad handlera",
			handlerErr: errHandler,
			middleware: func(next HandlerFunc) HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) error {
					_ = next(w, r)
					return errMiddleware
				}
			},
			wantSeenErr: errHandler,
			wantErr:     errMiddleware,
		},
		{
			name:       "middleware polyka blad handlera",
			handlerErr: errHandler,
			middleware: func(next HandlerFunc) HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) error {
					_ = next(w, r)
					return nil
				}
			},
			wantSeenErr: errHandler,
			wantErr:     nil,
		},
		{
			name:       "blad przechodzi gdy middleware podmienia kontekst",
			handlerErr: errHandler,
			middleware: func(next HandlerFunc) HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) error {
					ctx := context.WithValue(r.Context(), struct{}{}, "x")
					return next(w, r.WithContext(ctx))
				}
			},
			wantSeenErr: errHandler,
			wantErr:     errHandler,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var seenErr error
			// spy sits between the handler and the tested middleware, outer passes the error one more layer up
			spy := func(next HandlerFunc) HandlerFunc {
				return func(w http.ResponseWriter, r *http.Request) error {
					seenErr = next(w, r)
					return seenErr
				}
			}

			handler := toHTTPHandler(func(w http.ResponseWriter, r *http.Request) error {
				return tt.handlerErr
			})
			handler = toHTTPMiddleware(spy)(handler)
			handler = toHTTPMiddleware(tt.middleware)(handler)
			handler = toHTTPMiddleware(passThrough)(handler)

			var err error
			ctx := context.WithValue(context.Background(), errorSlotKey{}, &err)
			req, reqErr := http.NewRequestWithContext(ctx, http.MethodGet, "/", nil)
			assert.NoError(t, reqErr)

			handler.ServeHTTP(nil, req)

			assert.Equal(t, tt.wantSeenErr, seenErr)
			assert.Equal(t, tt.wantErr, err)
		})
	}
}
