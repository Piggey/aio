package aio

import (
	"context"
	"net/http"
)

// HandlerFunc is an http.HandlerFunc that returns an error instead of writing it.
// The error is turned into a response by App.ErrorHandler.
type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

type Middleware func(next HandlerFunc) HandlerFunc

// http.Handler can't return an error, so between the layers of the chain
// it is passed through a slot put in the request context by App.handleErrors.
type errorSlotKey struct{}

// toHTTPHandler puts the error returned by handler into the request's error slot.
func toHTTPHandler(handler HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := handler(w, r)
		setRequestError(r, err)
	})
}

// toHTTPMiddleware lets middleware see the error of the rest of the chain as the error returned by next.
func toHTTPMiddleware(middleware Middleware) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		handler := middleware(func(w http.ResponseWriter, r *http.Request) error {
			next.ServeHTTP(w, r)
			return takeRequestError(r)
		})

		return toHTTPHandler(handler)
	}
}

// handleErrors is the outermost layer of the chain, it responds with the error left in the slot.
func (app *App) handleErrors(handler http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var err error
		ctx := context.WithValue(r.Context(), errorSlotKey{}, &err)
		handler.ServeHTTP(w, r.WithContext(ctx))

		if err != nil {
			app.ErrorHandler(w, r, err)
		}
	})
}

func setRequestError(r *http.Request, err error) {
	slot, ok := r.Context().Value(errorSlotKey{}).(*error)
	if ok {
		*slot = err
	}
}

// takeRequestError returns the error from the slot and clears it, so it isn't reported twice.
func takeRequestError(r *http.Request) error {
	slot, ok := r.Context().Value(errorSlotKey{}).(*error)
	if !ok {
		return nil
	}

	err := *slot
	*slot = nil

	return err
}
