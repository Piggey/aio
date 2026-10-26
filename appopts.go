package aio

import (
	"github.com/Piggey/aio/internal/binder"
	"github.com/go-playground/validator/v10"
)

type appOpt func(app *App)

func WithOpenAPIOptions(openapiOpts ...oapiOpt) appOpt {
	return func(app *App) {
		for _, openapiOpt := range openapiOpts {
			openapiOpt(&app.oapi)
		}
	}
}

func WithValidator(validate *validator.Validate) appOpt {
	return func(app *App) {
		app.binder = binder.New(binder.WithValidator(validate))
	}
}
