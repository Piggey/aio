package aio

import (
	"reflect"

	"github.com/Piggey/aio/internal/openapi"
)

type oapiOpt func(oapi *oapi)

func WithoutOpenAPI() oapiOpt {
	return func(oapi *oapi) {
		oapi.isEnabled = false
	}
}

func WithUIPath(uiPath string) oapiOpt {
	return func(oapi *oapi) {
		oapi.uiPath = uiPath
	}
}

func WithOpenAPIPath(openAPIPath string) oapiOpt {
	return func(oapi *oapi) {
		oapi.schemaPath = openAPIPath
	}
}

func WithOpenAPIDescription(desc string) oapiOpt {
	return func(oapi *oapi) {
		oapi.openAPI.Info.Description = desc
	}
}

func WithOpenAPIContact(name, url, email string) oapiOpt {
	return func(oapi *oapi) {
		contact := openapi.Contact{
			Name:  name,
			URL:   url,
			Email: email,
		}

		oapi.openAPI.Info.Contact = &contact
	}
}

func WithOpenAPILicense(name, url string) oapiOpt {
	return func(oapi *oapi) {
		licence := openapi.License{
			Name: name,
			URL:  url,
		}

		oapi.openAPI.Info.License = &licence
	}
}

// WithEnumType declares T as an enum with the given values.
// Every occurrence of T in request and response types is described
// in the OpenAPI document as a reference to a component schema listing the values.
func WithEnumType[T ~string | ~int | ~uint](values []T) oapiOpt {
	return func(oapi *oapi) {
		enumValues := make([]any, len(values))
		for i, value := range values {
			enumValues[i] = value
		}

		oapi.enums[reflect.TypeFor[T]()] = enumValues
	}
}

// WithAPIKeySecurity declares an API key security scheme named name,
// with the key passed in the parameter paramName located in in
// (ParameterLocationHeader, ParameterLocationQuery or ParameterLocationCookie).
func WithAPIKeySecurity(name, paramName string, in ParameterLocation, opts ...securitySchemeOpt) oapiOpt {
	return func(oapi *oapi) {
		oapi.addSecurityScheme(name, openapi.SecurityScheme{
			Type: openapi.SecuritySchemeTypeAPIKey,
			Name: paramName,
			In:   in,
		}, opts)
	}
}

// WithHTTPSecurity declares an HTTP authentication security scheme named name.
// scheme is an HTTP authentication scheme from the IANA registry (e.g. "basic", "bearer", "digest").
func WithHTTPSecurity(name, scheme string, opts ...securitySchemeOpt) oapiOpt {
	return func(oapi *oapi) {
		oapi.addSecurityScheme(name, openapi.SecurityScheme{
			Type:   openapi.SecuritySchemeTypeHTTP,
			Scheme: scheme,
		}, opts)
	}
}

// WithBasicSecurity declares an HTTP Basic authentication security scheme named name.
func WithBasicSecurity(name string, opts ...securitySchemeOpt) oapiOpt {
	return WithHTTPSecurity(name, "basic", opts...)
}

// WithBearerSecurity declares an HTTP Bearer authentication security scheme named name.
// bearerFormat is an optional hint of the token format (e.g. "JWT").
func WithBearerSecurity(name, bearerFormat string, opts ...securitySchemeOpt) oapiOpt {
	return func(oapi *oapi) {
		oapi.addSecurityScheme(name, openapi.SecurityScheme{
			Type:         openapi.SecuritySchemeTypeHTTP,
			Scheme:       "bearer",
			BearerFormat: bearerFormat,
		}, opts)
	}
}

// WithMutualTLSSecurity declares a mutual TLS security scheme named name.
func WithMutualTLSSecurity(name string, opts ...securitySchemeOpt) oapiOpt {
	return func(oapi *oapi) {
		oapi.addSecurityScheme(name, openapi.SecurityScheme{
			Type: openapi.SecuritySchemeTypeMutualTLS,
		}, opts)
	}
}

// WithOAuth2Security declares an OAuth2 security scheme named name.
// The supported flows are declared with the WithOAuth2*Flow options.
func WithOAuth2Security(name string, opts ...securitySchemeOpt) oapiOpt {
	return func(oapi *oapi) {
		oapi.addSecurityScheme(name, openapi.SecurityScheme{
			Type:  openapi.SecuritySchemeTypeOAuth2,
			Flows: &openapi.OAuth2Flows{},
		}, opts)
	}
}

// WithOpenIDConnectSecurity declares an OpenID Connect security scheme named name.
// url points to the OpenID Connect Discovery document.
func WithOpenIDConnectSecurity(name, url string, opts ...securitySchemeOpt) oapiOpt {
	return func(oapi *oapi) {
		oapi.addSecurityScheme(name, openapi.SecurityScheme{
			Type:             openapi.SecuritySchemeTypeOpenIDConnect,
			OpenIDConnectURL: url,
		}, opts)
	}
}
