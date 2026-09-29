package aio

import "net/http"

type serverOpt func(s *Server)

func WithOpenAPIOptions(oapiOpts ...openapiOpt) serverOpt {
	return func(s *Server) {
		for _, oapiOpt := range oapiOpts {
			oapiOpt(s.oapi)
		}
	}
}

func WithGlobalEndpointOptions(endpointOpts ...endpointOpt) serverOpt {
	return func(s *Server) {
		for _, endpointOpt := range endpointOpts {
			endpointOpt(s)
		}
	}
}

func WithEndpoint[In any, Out any](method, path string, handlerFunc HandlerFunc[In, Out], endpointOpts ...endpointOpt) serverOpt {
	return func(s *Server) {
		endpoint := endpoint{
			method:  method,
			path:    path,
			handler: wrapHandlerFunc(s, handlerFunc),
		}

		s.endpoints = append(s.endpoints, endpoint)
	}
}

func wrapHandlerFunc[In, Out any](s *Server, handlerFunc HandlerFunc[In, Out]) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var in In
		err := s.binder.Bind(r, &in)
		if err != nil {
			return
		}

		out, err := handlerFunc(r.Context(), in)
		if err != nil {
			return
		}

		err = s.binder.BindOut(r, w, out)
	}
}

func WithGroup(path string, groupOpts ...groupOpt) serverOpt {
	return func(s *Server) {
		panic("impl")
	}
}
