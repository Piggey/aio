package aio

import (
	"github.com/Piggey/aio/internal/openapi"
)

type oapi struct {
	isEnabled   bool
	uiPath      string
	openAPIPath string
	openAPI     openapi.OpenAPI
}
