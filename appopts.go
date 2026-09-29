package aio

import (
	"github.com/Piggey/binder"
	"github.com/go-playground/validator/v10"
)

type appOpt func(app *App)

func WithOpenAPIOptions(openapiOpts ...openapiOpt) appOpt {
	return func(app *App) {
		for _, openapiOpt := range openapiOpts {
			openapiOpt(&app.oapi)
		}
	}
}

func WithCustomValidator(validate *validator.Validate) appOpt {
	return func(app *App) {
		app.binder = binder.New(binder.WithCustomValidatorInstance(validate))
	}
}
