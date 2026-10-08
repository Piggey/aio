package aio

import (
	"net/http"
	"reflect"
)

type endpoint struct {
	method  string
	path    string
	handler http.Handler
	inType  reflect.Type
	outType reflect.Type
}
