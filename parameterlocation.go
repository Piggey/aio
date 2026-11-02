package aio

import "github.com/Piggey/aio/internal/openapi"

// ParameterLocation is the location of a parameter in an HTTP request.
type ParameterLocation = openapi.ParameterLocation

const (
	ParameterLocationPath   ParameterLocation = openapi.ParameterLocationPath
	ParameterLocationQuery  ParameterLocation = openapi.ParameterLocationQuery
	ParameterLocationHeader ParameterLocation = openapi.ParameterLocationHeader
	ParameterLocationCookie ParameterLocation = openapi.ParameterLocationCookie
)
