package aio

import "github.com/Piggey/aio/internal/openapi"

type oapiOpt func(oapi *oapi)

func WithoutOpenAPI() oapiOpt {
	return func(oapi *oapi) {
		oapi.isEnabled = false
	}
}

func WithCustomUIPath(uiPath string) oapiOpt {
	return func(oapi *oapi) {
		oapi.uiPath = uiPath
	}
}

func WithCustomOpenAPIPath(openAPIPath string) oapiOpt {
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

// WithAPIKeyHeaderSecurity declares an API key security scheme named name,
// with the key passed in the header headerName.
func WithAPIKeyHeaderSecurity(name, headerName string, opts ...securitySchemeOpt) oapiOpt {
	return withAPIKeySecurity(name, headerName, openapi.ParameterLocationHeader, opts)
}

// WithAPIKeyQuerySecurity declares an API key security scheme named name,
// with the key passed in the query parameter paramName.
func WithAPIKeyQuerySecurity(name, paramName string, opts ...securitySchemeOpt) oapiOpt {
	return withAPIKeySecurity(name, paramName, openapi.ParameterLocationQuery, opts)
}

// WithAPIKeyCookieSecurity declares an API key security scheme named name,
// with the key passed in the cookie cookieName.
func WithAPIKeyCookieSecurity(name, cookieName string, opts ...securitySchemeOpt) oapiOpt {
	return withAPIKeySecurity(name, cookieName, openapi.ParameterLocationCookie, opts)
}

func withAPIKeySecurity(name, paramName string, in openapi.ParameterLocation, opts []securitySchemeOpt) oapiOpt {
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
