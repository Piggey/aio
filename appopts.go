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

// WithEndpointOptions applies endpointOpts to every endpoint of the App, including the ones in groups.
// They are applied before the options passed when registering the endpoint.
func WithEndpointOptions(endpointOpts ...endpointOpt) appOpt {
	return func(app *App) {
		app.endpointOpts = append(app.endpointOpts, endpointOpts...)
	}
}

func WithValidator(validate *validator.Validate) appOpt {
	return func(app *App) {
		app.binder = binder.New(binder.WithValidator(validate))
	}
}
