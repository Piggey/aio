package aio

type ParameterLocation string

const (
	ParameterLocationPath   ParameterLocation = "path"
	ParameterLocationQuery  ParameterLocation = "query"
	ParameterLocationHeader ParameterLocation = "header"
	ParameterLocationCookie ParameterLocation = "cookie"
)
