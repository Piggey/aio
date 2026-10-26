package aio

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"slices"

	"github.com/Piggey/aio/internal/binder"
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
		oapi: defaultOAPI(name, version),
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
		middlewares: slices.Clone(app.middlewares),
	}

	app.groups = append(app.groups, g)

	return &g
}

func (app *App) Get[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error), opts ...endpointOpt) {
	app.Register(http.MethodGet, path, handlerFunc, opts...)
}

func (app *App) Put[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error), opts ...endpointOpt) {
	app.Register(http.MethodPut, path, handlerFunc, opts...)
}

func (app *App) Post[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error), opts ...endpointOpt) {
	app.Register(http.MethodPost, path, handlerFunc, opts...)
}

func (app *App) Delete[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error), opts ...endpointOpt) {
	app.Register(http.MethodDelete, path, handlerFunc, opts...)
}

func (app *App) Options[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error), opts ...endpointOpt) {
	app.Register(http.MethodOptions, path, handlerFunc, opts...)
}

func (app *App) Head[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error), opts ...endpointOpt) {
	app.Register(http.MethodHead, path, handlerFunc, opts...)
}

func (app *App) Patch[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error), opts ...endpointOpt) {
	app.Register(http.MethodPatch, path, handlerFunc, opts...)
}

func (app *App) Trace[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error), opts ...endpointOpt) {
	app.Register(http.MethodTrace, path, handlerFunc, opts...)
}

func (app *App) Query[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error), opts ...endpointOpt) {
	app.Register("QUERY", path, handlerFunc, opts...)
}

func (app *App) Register[In, Out any](method string, path string, handlerFunc func(context.Context, In) (Out, error), opts ...endpointOpt) {
	app.addEndpoint(method, path, handlerFunc, app.middlewares, opts...)
}

func (app *App) Run(addr string) error {
	return http.ListenAndServe(addr, app.intoMux())
}

func (app *App) RunTLS(addr string, certFile string, keyFile string) error {
	return http.ListenAndServeTLS(addr, certFile, keyFile, app.intoMux())
}

func (app *App) addEndpoint[In, Out any](method string, path string, handlerFunc func(context.Context, In) (Out, error), middlewares []func(http.Handler) http.Handler, opts ...endpointOpt) {
	inType := reflect.TypeFor[In]()
	if inType.Kind() != reflect.Struct {
		panic("aio: In type has to be a struct")
	}

	outType := reflect.TypeFor[Out]()
	if outType.Kind() != reflect.Struct {
		panic("aio: Out type has to be a struct")
	}

	ep := endpoint{
		method:  method,
		path:    path,
		inType:  inType,
		outType: outType,
	}

	for _, opt := range opts {
		opt(&ep)
	}

	handler := app.wrapHandlerFunc(handlerFunc)

	for _, middleware := range ep.middlewares {
		handler = middleware(handler)
	}

	for _, middleware := range middlewares {
		handler = middleware(handler)
	}

	ep.handler = handler

	app.endpoints = append(app.endpoints, ep)
	slog.Debug("registered endpoint", slog.String("method", method), slog.String("path", path))
}

func (app *App) wrapHandlerFunc[In, Out any](handlerFunc func(context.Context, In) (Out, error)) http.Handler {
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
		pattern := fmt.Sprintf("%s %s", ep.method, ep.path)
		mux.Handle(pattern, ep.handler)

		if app.oapi.isEnabled {
			app.oapi.addEndpoint(ep)
		}
	}

	if app.oapi.isEnabled {
		mux.Handle(fmt.Sprintf("%s %s", http.MethodGet, app.oapi.schemaPath), app.oapi.schemaHandler())
		mux.Handle(fmt.Sprintf("%s %s", http.MethodGet, app.oapi.uiPath), app.oapi.uiHandler())
	}

	return mux
}
