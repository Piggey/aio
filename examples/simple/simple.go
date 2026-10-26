package main

import (
	"context"
	"fmt"
	"log"
	"log/slog"
	"net/http"

	"github.com/Piggey/aio"
	"github.com/Piggey/aio/middleware"
	"github.com/go-playground/validator/v10"
)

type MyCustomEnum string

const (
	MyCustomEnumFirst  MyCustomEnum = "first"
	MyCustomEnumSecond MyCustomEnum = "second"
	MyCustomEnumThird  MyCustomEnum = "third"
)

var MyCustomEnumValues = []MyCustomEnum{MyCustomEnumFirst, MyCustomEnumSecond, MyCustomEnumThird}

func main() {
	slog.SetLogLoggerLevel(slog.LevelDebug)

	app := aio.New(
		"my cool api", "v1.0.0",
		aio.WithOpenAPIOptions(
			aio.WithAPIKeySecurity("apiKey", "api_key", aio.ParameterLocationQuery),
			aio.WithEnumType(MyCustomEnumValues),
		),
		aio.WithValidator(validator.New()),
	)

	app.Use(middleware.Logger(nil))

	app.Get("/test", func(ctx context.Context, _ struct{}) (struct{}, error) {
		log.Println("test")
		return struct{}{}, nil
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

	kutas.Use(func(next aio.HandlerFunc) aio.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) error {
			log.Println("i sie zesralo")
			err := next(w, r)
			log.Println("prrrrt", err)
			return err
		}
	})

	kutas.Get("/huj", func(ctx context.Context, _ struct{}) (struct{}, error) {
		return struct{}{}, aio.Error(http.StatusBadRequest, "nie podoba mi sie ten request", nil)
	})

	err := app.Run(":8080")
	if err != nil {
		log.Panicf("app.ListenAndServe: %v", err)
	}
}

type (
	TestRequest struct {
		Limit        int          `query:"limit"`
		Offset       int          `query:"offset"`
		Name         string       `json:"name"`
		Age          int          `json:"age"`
		MyCustomEnum MyCustomEnum `json:"myCustomEnum"`
	}

	TestResponse struct {
		Success bool `json:"success"`
	}
)
