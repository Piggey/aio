package main

import (
	"context"
	"log"
	"log/slog"
	"net/http"

	"github.com/Piggey/aio"
	"github.com/go-playground/validator/v10"
)

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	app := aio.New(
		"my cool api", "v1.0.0",
		aio.WithOpenAPIOptions(),
		aio.WithCustomValidator(validator.New()),
	)

	app.Get("/test", func(ctx context.Context, _ struct{}) (string, error) {
		log.Println("test")
		return "test", nil
	})

	kutas := app.Group("/kutas")

	kutas.Use(func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			log.Println("i sie zesralo")
			next.ServeHTTP(w, r)
			log.Println("prrrrt")
		})
	})

	kutas.Get("/huj", func(ctx context.Context, _ struct{}) (string, error) {
		log.Println("huj")
		return "", aio.ErrBadRequest
	})

	err := app.Run(":8080")
	if err != nil {
		log.Panicf("app.ListenAndServe: %v", err)
	}
}
