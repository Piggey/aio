package main

import (
	"context"
	"log"

	"github.com/Piggey/aio"
	"github.com/go-playground/validator/v10"
)

func main() {
	app := aio.New(
		"my cool api",
		aio.WithOpenAPIOptions(
			aio.WithOpenAPIVersion("1.0.0"),
		),
		aio.WithCustomValidator(validator.New()),
	)

	app.Get("/test", func(ctx context.Context, _ any) (string, error) {
		return "test", nil
	})

	err := app.ListenAndServe(":3000")
	if err != nil {
		log.Panicf("app.ListenAndServe: %v", err)
	}
}
