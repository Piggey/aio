package aio

import (
	"net/http"
)

type endpoint struct {
	method      string
	path        string
	middlewares []http.Handler
	handler     http.Handler
}

func NewEndpoint[In, Out any](method string, path string, handlerFunc HandlerFunc[In, Out], endpointOpts ...endpointOpt) endpoint {
	endpoint := endpoint{
		method: method,
		path: path,
		middlewares: []http.Handler{},
		handler: wrapHandlerFunc(),
	}
}
