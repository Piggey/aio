package aio

type endpointOpt func(ep *endpoint)

func WithSummary(summary string) endpointOpt {
	return func(ep *endpoint) {
		ep.summary = summary
	}
}

func WithDescription(description string) endpointOpt {
	return func(ep *endpoint) {
		ep.description = description
	}
}

func WithTags(tags ...string) endpointOpt {
	return func(ep *endpoint) {
		ep.tags = append(ep.tags, tags...)
	}
}

// WithErrorStatusCodes declares HTTP status codes the endpoint may return an error with.
// Each of them is described in OpenAPI as a ProblemDetails response.
func WithErrorStatusCodes(statusCodes ...int) endpointOpt {
	return func(ep *endpoint) {
		ep.errorStatusCodes = append(ep.errorStatusCodes, statusCodes...)
	}
}

func WithDeprecated() endpointOpt {
	return func(ep *endpoint) {
		ep.deprecated = true
	}
}
