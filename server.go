package aio

import (
	"net/http"

	"github.com/Piggey/aio/internal/openapi"
	"github.com/Piggey/binder"
)

type Server struct {
	binder *binder.Binder
	mux    *http.ServeMux
	oapi   *openapi.OpenAPI

	middlewares []http.Handler
	endpoints   []endpoint
	groups      []group
}

func NewServer(title, version string, serverOpts ...serverOpt) *Server {
	s := &Server{
		binder: nil,
		mux:    http.NewServeMux(),
		oapi: &openapi.OpenAPI{
			OpenAPI: openapi.OpenAPIVersion,
			Info: openapi.Info{
				Title:   title,
				Version: version,
			},
		},
	}

	for _, serverOpt := range serverOpts {
		serverOpt(s)
	}

	if s.binder == nil {
		s.binder = binder.New()
	}

	return s
}

func (s *Server) Run(addr string) error {
	return http.ListenAndServe(addr, s.mux)
}

func (s *Server) RunTLS(addr string, certFile string, keyFile string) error {
	return http.ListenAndServeTLS(addr, certFile, keyFile, s.mux)
}
