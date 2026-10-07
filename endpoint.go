package aio

import (
	"net/http"
)

type endpoint struct {
	pattern string
	handler http.Handler
}

