package main

import (
	"context"
	"fmt"
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

	app.Post("/test", func(ctx context.Context, req TestRequest) (TestResponse, error) {
		if req.Age < 18 {
			return TestResponse{}, fmt.Errorf("chyba cie pojebalo")
		}

		return TestResponse{
			Success: true,
		}, nil
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

type (
	TestRequest struct {
		Limit  int    `query:"limit"`
		Offset int    `query:"offset"`
		Name   string `json:"name"`
		Age    int    `json:"age"`
	}

	TestResponse struct {
		Success bool `json:"success"`
	}
)
