package aio

type endpointOpt func(ep *Endpoint)

// WithMiddlewares adds middlewares applied only to this endpoint.
// They run after the App and Group middlewares, right before the handler.
func WithMiddlewares(middlewares ...Middleware) endpointOpt {
	return func(ep *Endpoint) {
		for _, middleware := range middlewares {
			ep.middlewares = append(ep.middlewares, toHTTPMiddleware(middleware))
		}
	}
}

func WithSummary(summary string) endpointOpt {
	return func(ep *Endpoint) {
		ep.summary = summary
	}
}

func WithDescription(description string) endpointOpt {
	return func(ep *Endpoint) {
		ep.description = description
	}
}

func WithTags(tags ...string) endpointOpt {
	return func(ep *Endpoint) {
		ep.tags = append(ep.tags, tags...)
	}
}

// WithSecurity declares in OpenAPI that the endpoint requires the security scheme named name,
// declared before with one of the With...Security options. scopes are needed only for OAuth2 and OpenID Connect.
// Every call adds an alternative - the client has to satisfy only one of them.
func WithSecurity(name string, scopes ...string) endpointOpt {
	return func(ep *Endpoint) {
		requirement := map[string][]string{
			name: append([]string{}, scopes...), // scopes are required, serialize as [] instead of null
		}
		ep.security = append(ep.security, requirement)
	}
}

// WithErrorStatusCodes declares HTTP status codes the endpoint may return an error with.
// Each of them is described in OpenAPI as a ProblemDetails response.
func WithErrorStatusCodes(statusCodes ...int) endpointOpt {
	return func(ep *Endpoint) {
		ep.errorStatusCodes = append(ep.errorStatusCodes, statusCodes...)
	}
}

func WithDeprecated() endpointOpt {
	return func(ep *Endpoint) {
		ep.deprecated = true
	}
}
