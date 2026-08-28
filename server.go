package aio

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Piggey/binder"
)

type Server struct {
	binder *binder.Binder
	mux    *http.ServeMux
	oapi   openapi
}

func NewServer(title, version string, serverOpts ...serverOpt) *Server {
	s := &Server{
		binder: binder.New(),
		mux:    http.NewServeMux(),
	}

	for _, serverOpt := range serverOpts {
		serverOpt(s)
	}

	return s
}

func (s *Server) Run(addr string) error {
	return http.ListenAndServe(addr, s.mux)
}

type serverOpt func(s *Server)

type HandlerFunc[In, Out any] func(context.Context, In) (Out, error)

func WithEndpoint[In any, Out any](method, path string, handlerFunc HandlerFunc[In, Out]) serverOpt {
	return func(s *Server) {
		pattern := fmt.Sprint(method, " ", path)
		s.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
			var in In
			err := s.binder.Bind(r, &in)
			if err != nil {
				w.Write([]byte(err.Error()))
				return
			}

			out, err := handlerFunc(r.Context(), in)
			if err != nil {
				w.Write([]byte(err.Error()))
				return
			}

			err = json.NewEncoder(w).Encode(out)
			if err != nil {
				w.Write([]byte(err.Error()))
				return
			}
		})
	}
}

type oapiOpt func(oapi *openapi)

func WithOpenAPIOptions(oapiOpts ...oapiOpt) serverOpt {
	return func(s *Server) {
		for _, oapiOpt := range oapiOpts {
			oapiOpt(&s.oapi)
		}
	}
}

func WithOpenAPIDescription(desc string) oapiOpt {
	return func(oapi *openapi) {
		oapi.Info.Description = desc
	}
}
