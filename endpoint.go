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
