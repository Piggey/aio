package aio

import (
	"reflect"
)

type endpoint struct {
	method      string
	path        string
	handler     HandlerFunc
	middlewares []Middleware
	// preflight runs the same middlewares as handler, but ends with 204 No Content instead of the handler
	preflight HandlerFunc

	// openAPI
	summary          string
	description      string
	tags             []string
	inType           reflect.Type
	outType          reflect.Type
	errorStatusCodes []int
	deprecated       bool
	security         []map[string][]string
}
