package aio

import (
	"net/http"
)

// HandlerFunc is an http.HandlerFunc that returns an error instead of writing it.
// The error is turned into a response by App.ErrorHandler.
type HandlerFunc func(w http.ResponseWriter, r *http.Request) error

type Middleware func(next HandlerFunc) HandlerFunc
