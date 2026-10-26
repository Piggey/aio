package aio

import (
	"net/http"
	"reflect"
)

type endpoint struct {
	method      string
	path        string
	handler     http.Handler
	middlewares []func(http.Handler) http.Handler

	// openAPI
	summary          string
	description      string
	tags             []string
	inType           reflect.Type
	outType          reflect.Type
	errorStatusCodes []int
	deprecated       bool
}
