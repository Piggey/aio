package aio

import "net/http"

type Group struct {
	path        string
	middlewares []http.Handler
	endpoints   []endpoint
}
