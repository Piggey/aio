package aio

type groupOpt func(g *group)

func WithGroupEndpointOptions(endpointOpts ...endpointOpt) groupOpt {
	return func(g *group) {
		g.endpointOpts = append(g.endpointOpts, endpointOpts...)
	}
}

func WithGroupEndpoints(endpoints ...endpoint) groupOpt {
	return func(g *group) {
		g.endpoints = append(g.endpoints, endpoints...)
	}
}
