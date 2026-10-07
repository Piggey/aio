package aio

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"

	"github.com/Piggey/aio/internal/binder"
	"github.com/Piggey/aio/internal/openapi"
)

type App struct {
	binder *binder.Binder
	oapi   oapi

	middlewares []func(http.Handler) http.Handler
	endpoints   []endpoint
	groups      []Group
}

func New(name string, version string, appOpts ...appOpt) *App {
	app := App{
		binder: nil,
		oapi: oapi{
			isEnabled:   true,
			uiPath:      "/swagger/ui",
			openAPIPath: "/swagger/json",
			openAPI: openapi.OpenAPI{
				OpenAPI: openapi.OpenAPIVersion,
				Info: openapi.Info{
					Title:   name,
					Version: version,
				},
			},
		},
	}

	for _, appOpt := range appOpts {
		appOpt(&app)
	}

	if app.binder == nil {
		app.binder = binder.New()
	}

	return &app
}

func (app *App) ErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	pd, ok := errors.AsType[ProblemDetails](err)
	if !ok {
		pd = ProblemDetails{
			Title:  http.StatusText(http.StatusInternalServerError),
			Status: http.StatusInternalServerError,
			Detail: nil,
		}
	}

	_ = app.binder.BindOut(w, r, pd.Status, pd)
}

func (app *App) Use(middleware func(http.Handler) http.Handler) {
	app.middlewares = append(app.middlewares, middleware)
}

func (app *App) Group(path string) *Group {
	g := Group{
		app:         app,
		parent:      nil,
		path:        path,
		middlewares: []func(http.Handler) http.Handler{},
	}

	copy(g.middlewares, app.middlewares)
	app.groups = append(app.groups, g)

	return &g
}

func (app *App) Get[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error)) {
	app.Register(http.MethodGet, path, handlerFunc)
}

func (app *App) Put[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error)) {
	app.Register(http.MethodPut, path, handlerFunc)
}

func (app *App) Post[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error)) {
	app.Register(http.MethodPost, path, handlerFunc)
}

func (app *App) Delete[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error)) {
	app.Register(http.MethodDelete, path, handlerFunc)
}

func (app *App) Options[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error)) {
	app.Register(http.MethodOptions, path, handlerFunc)
}

func (app *App) Head[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error)) {
	app.Register(http.MethodHead, path, handlerFunc)
}

func (app *App) Patch[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error)) {
	app.Register(http.MethodPatch, path, handlerFunc)
}

func (app *App) Trace[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error)) {
	app.Register(http.MethodTrace, path, handlerFunc)
}

func (app *App) Query[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error)) {
	app.Register("QUERY", path, handlerFunc)
}

func (app *App) Register[In, Out any](method string, path string, handlerFunc func(context.Context, In) (Out, error)) {
	app.addEndpoint(method, path, app.wrapHandlerFunc(handlerFunc), app.middlewares)
}

func (app *App) Run(addr string) error {
	return http.ListenAndServe(addr, app.intoMux())
}

func (app *App) RunTLS(addr string, certFile string, keyFile string) error {
	return http.ListenAndServeTLS(addr, certFile, keyFile, app.intoMux())
}

func (app *App) addEndpoint(method string, path string, handler http.Handler, middlewares []func(http.Handler) http.Handler) {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}

	pattern := fmt.Sprintf("%s %s", method, path)

	ep := endpoint{
		pattern: pattern,
		handler: handler,
	}

	app.endpoints = append(app.endpoints, ep)
	slog.Debug("registered endpoint", slog.String("pattern", pattern))
}

func (app *App) wrapHandlerFunc[In, Out any](handlerFunc func(context.Context, In) (Out, error)) http.Handler {
	if app.oapi.isEnabled {
		app.oapi.schemaTypes = append(app.oapi.schemaTypes, reflect.TypeFor[In](), reflect.TypeFor[Out]())
	}

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var in In
		err := app.binder.BindIn(r, &in)
		if err != nil {
			app.ErrorHandler(w, r, ErrBadRequest)
			return
		}

		out, err := handlerFunc(r.Context(), in)
		if err != nil {
			app.ErrorHandler(w, r, ErrInternalServerError)
			return
		}

		err = app.binder.BindOut(w, r, http.StatusOK, out)
		if err != nil {
			app.ErrorHandler(w, r, ErrInternalServerError)
			return
		}
	})
}

func (app *App) intoMux() *http.ServeMux {
	mux := http.NewServeMux()
	for _, ep := range app.endpoints {
		mux.Handle(ep.pattern, ep.handler)
	}

	return mux
}
