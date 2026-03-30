package aio

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/Piggey/binder"
)

type Server struct{
	binder *binder.Binder
	mux *http.ServeMux
}

func NewServer(serverOpts ...serverOpt) *Server {
	s := &Server{
		binder: binder.New(),
		mux: http.NewServeMux(),
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

func WithEndpoint[In any, Out any](method, path string, handler func(context.Context, In) (Out, error)) serverOpt {
	return func(s *Server) {
		pattern := fmt.Sprint(method, " ", path)
		s.mux.HandleFunc(pattern, func(w http.ResponseWriter, r *http.Request) {
			var in In
			err := s.binder.Bind(r, &in)
			if err != nil {
				w.Write([]byte(err.Error()))
				return 
			}

			out, err := handler(r.Context(), in)
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
