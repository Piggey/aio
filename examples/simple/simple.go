package main

import (
	"context"
	"log"

	"github.com/Piggey/aio"
	"github.com/go-playground/validator/v10"
)

func main() {
	app := aio.New(
		"my cool api", "v1.0.0",
		aio.WithOpenAPIOptions(),
		aio.WithCustomValidator(validator.New()),
	)

	app.Get("/test", func(ctx context.Context, _ any) (string, error) {
		return "test", aio.ErrBadRequest
	})

	err := app.Run(":3000")
	if err != nil {
		log.Panicf("app.ListenAndServe: %v", err)
	}
}
