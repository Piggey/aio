package aio

import (
	"reflect"

	"github.com/Piggey/aio/internal/openapi"
)

type oapi struct {
	isEnabled   bool
	uiPath      string
	openAPIPath string
	schemaTypes []reflect.Type
	openAPI     openapi.OpenAPI
}
