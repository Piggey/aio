package aio

import "github.com/Piggey/aio/internal/openapi"

type securitySchemeOpt func(scheme *openapi.SecurityScheme)

type oauth2FlowOpt func(flow *openapi.OAuth2Flow)

// WithSecuritySchemeDescription sets the description of a security scheme.
func WithSecuritySchemeDescription(desc string) securitySchemeOpt {
	return func(scheme *openapi.SecurityScheme) {
		scheme.Description = desc
	}
}

// WithSecuritySchemeDeprecated marks a security scheme as deprecated.
func WithSecuritySchemeDeprecated() securitySchemeOpt {
	return func(scheme *openapi.SecurityScheme) {
		scheme.Deprecated = true
	}
}

// WithOAuth2MetadataURL sets the URL of the OAuth2 authorization server metadata (RFC 8414).
// Applies only to WithOAuth2Security.
func WithOAuth2MetadataURL(url string) securitySchemeOpt {
	return func(scheme *openapi.SecurityScheme) {
		scheme.OAuth2MetadataURL = url
	}
}

// WithOAuth2ImplicitFlow declares the OAuth2 Implicit flow. Applies only to WithOAuth2Security.
// scopes maps scope names to their short descriptions and may be empty.
func WithOAuth2ImplicitFlow(authorizationURL string, scopes map[string]string, opts ...oauth2FlowOpt) securitySchemeOpt {
	return withOAuth2Flow(func(flows *openapi.OAuth2Flows, flow *openapi.OAuth2Flow) {
		flow.AuthorizationURL = authorizationURL
		flows.Implicit = flow
	}, scopes, opts)
}

// WithOAuth2PasswordFlow declares the OAuth2 Resource Owner Password flow. Applies only to WithOAuth2Security.
// scopes maps scope names to their short descriptions and may be empty.
func WithOAuth2PasswordFlow(tokenURL string, scopes map[string]string, opts ...oauth2FlowOpt) securitySchemeOpt {
	return withOAuth2Flow(func(flows *openapi.OAuth2Flows, flow *openapi.OAuth2Flow) {
		flow.TokenURL = tokenURL
		flows.Password = flow
	}, scopes, opts)
}

// WithOAuth2ClientCredentialsFlow declares the OAuth2 Client Credentials flow. Applies only to WithOAuth2Security.
// scopes maps scope names to their short descriptions and may be empty.
func WithOAuth2ClientCredentialsFlow(tokenURL string, scopes map[string]string, opts ...oauth2FlowOpt) securitySchemeOpt {
	return withOAuth2Flow(func(flows *openapi.OAuth2Flows, flow *openapi.OAuth2Flow) {
		flow.TokenURL = tokenURL
		flows.ClientCredentials = flow
	}, scopes, opts)
}

// WithOAuth2AuthorizationCodeFlow declares the OAuth2 Authorization Code flow. Applies only to WithOAuth2Security.
// scopes maps scope names to their short descriptions and may be empty.
func WithOAuth2AuthorizationCodeFlow(authorizationURL, tokenURL string, scopes map[string]string, opts ...oauth2FlowOpt) securitySchemeOpt {
	return withOAuth2Flow(func(flows *openapi.OAuth2Flows, flow *openapi.OAuth2Flow) {
		flow.AuthorizationURL = authorizationURL
		flow.TokenURL = tokenURL
		flows.AuthorizationCode = flow
	}, scopes, opts)
}

// WithOAuth2DeviceAuthorizationFlow declares the OAuth2 Device Authorization flow (RFC 8628). Applies only to WithOAuth2Security.
// scopes maps scope names to their short descriptions and may be empty.
func WithOAuth2DeviceAuthorizationFlow(deviceAuthorizationURL, tokenURL string, scopes map[string]string, opts ...oauth2FlowOpt) securitySchemeOpt {
	return withOAuth2Flow(func(flows *openapi.OAuth2Flows, flow *openapi.OAuth2Flow) {
		flow.DeviceAuthorizationURL = deviceAuthorizationURL
		flow.TokenURL = tokenURL
		flows.DeviceAuthorization = flow
	}, scopes, opts)
}

// WithOAuth2RefreshURL sets the URL used to obtain refresh tokens in an OAuth2 flow.
func WithOAuth2RefreshURL(url string) oauth2FlowOpt {
	return func(flow *openapi.OAuth2Flow) {
		flow.RefreshURL = url
	}
}

func withOAuth2Flow(set func(flows *openapi.OAuth2Flows, flow *openapi.OAuth2Flow), scopes map[string]string, opts []oauth2FlowOpt) securitySchemeOpt {
	return func(scheme *openapi.SecurityScheme) {
		if scheme.Flows == nil {
			return
		}

		if scopes == nil {
			scopes = map[string]string{} // scopes are required, serialize as {} instead of null
		}

		flow := &openapi.OAuth2Flow{Scopes: scopes}
		for _, opt := range opts {
			opt(flow)
		}

		set(scheme.Flows, flow)
	}
}
