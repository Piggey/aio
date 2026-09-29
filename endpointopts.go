package aio

import (
	"context"
	"net/http"
)

type endpointOpt func(e *endpoint)

type HandlerFunc[In, Out any] func(context.Context, In) (Out, error)

func WithMiddleware(handler http.Handler) endpointOpt {
	return func(e *endpoint) {
		e.middlewares = append(e.middlewares, handler)
	}
}
