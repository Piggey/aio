package aio

// OpenAPITyper overrides the OpenAPI data type (e.g. "string", "integer") of a custom type.
// The method is called on a zero value of the type.
type OpenAPITyper interface {
	OpenAPIType() string
}

// OpenAPIFormatter overrides the OpenAPI data format (e.g. "uuid", "date", "email") of a custom type.
// The method is called on a zero value of the type.
type OpenAPIFormatter interface {
	OpenAPIFormat() string
}
