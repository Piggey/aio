package main

import (
	"context"
	"log"
	"net/http"

	"github.com/Piggey/aio"
)

func main() {
	s := aio.NewServer(
		"simple-api", "1.0.0",
		aio.WithOpenAPIOptions(
			aio.WithOpenAPIDescription("desc"),
		),
		aio.WithGlobalEndpointOptions(),
		aio.WithEndpoint(
			http.MethodPost, "/test", handleTest,
		),
		aio.WithGroup(
			"/v1",
			aio.WithGroupEndpointOptions(),
			aio.WithGroupEndpoints(
			),
		),
		aio.WithEndpoint(
			http.MethodGet, "/ping", func(ctx context.Context, _ any) (string, error) {
				return "pong", nil
			},
		),
	)

	err := s.Run(":8080")
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
