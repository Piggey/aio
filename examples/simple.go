package main

import (
	"context"
	"log"
	"net/http"

	"github.com/Piggey/aio"
)

func main() {
	s := aio.NewServer(
		aio.WithEndpoint(
			http.MethodPost, "/test", handleTest,
		),
		aio.WithEndpoint(
			http.MethodGet, "/ping", func(ctx context.Context, _ any) (string, error) {
				return "pong", nil
			},
		),
	)

	err := s.Run("localhost:8080")
	if err != nil {
		log.Panicf("s.Run: %v", err)
	}
}

type (
	testRequest struct {
		Name string `json:"name" validate:"required"`
	}

	testResponse struct {
		Value int `json:"value"`
	}
)

func handleTest(ctx context.Context, testReq testRequest) ([]testResponse, error) {
	return []testResponse{
		{
			Value: len(testReq.Name),
		},
	}, nil
}
