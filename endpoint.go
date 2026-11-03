package aio

import (
	"reflect"
)

type Endpoint struct {
	method      string
	path        string
	handler     HandlerFunc
	middlewares []Middleware

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
