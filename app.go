package aio

import (
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"reflect"
	"slices"
	"strings"
	"time"

	"github.com/Piggey/aio/internal/binder"
)

type Empty struct{}

type App struct {
	binder *binder.Binder
	oapi   oapi
	debug  bool

	middlewares  []Middleware
	endpointOpts []endpointOpt
	endpoints    []Endpoint
	groups       []Group
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
	pd := problemDetailsFor(err, app.debug)
	_ = app.binder.BindOut(w, r, pd.Status, pd)
}

func problemDetailsFor(err error, debug bool) ProblemDetails {
	pd, ok := errors.AsType[ProblemDetails](err)
	if !ok {
		pd = ProblemDetails{
			Title:  http.StatusText(http.StatusInternalServerError),
			Status: http.StatusInternalServerError,
			Detail: err.Error(),
		}
	}

	if !debug {
		pd.Detail = ""
	}

	return pd
}

// UseMiddleware adds middlewares to the endpoints and groups registered after the call.
func (app *App) UseMiddleware(middlewares ...Middleware) {
	app.middlewares = append(app.middlewares, middlewares...)
}

// UseEndpointOptions adds endpointOpts to the endpoints and groups registered after the call.
// They are applied before the options passed when registering the endpoint.
func (app *App) UseEndpointOptions(endpointOpts ...endpointOpt) {
	app.endpointOpts = append(app.endpointOpts, endpointOpts...)
}

func (app *App) Group(path string) *Group {
	g := Group{
		app:          app,
		parent:       nil,
		path:         path,
		middlewares:  slices.Clone(app.middlewares),
		endpointOpts: slices.Clone(app.endpointOpts),
	}

	app.groups = append(app.groups, g)
	return &g
}

func (app *App) Get[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error), endpointOpts ...endpointOpt) {
	app.Register(http.MethodGet, path, handlerFunc, endpointOpts...)
}

func (app *App) Put[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error), endpointOpts ...endpointOpt) {
	app.Register(http.MethodPut, path, handlerFunc, endpointOpts...)
}

func (app *App) Post[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error), endpointOpts ...endpointOpt) {
	app.Register(http.MethodPost, path, handlerFunc, endpointOpts...)
}

func (app *App) Delete[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error), endpointOpts ...endpointOpt) {
	app.Register(http.MethodDelete, path, handlerFunc, endpointOpts...)
}

func (app *App) Options[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error), endpointOpts ...endpointOpt) {
	app.Register(http.MethodOptions, path, handlerFunc, endpointOpts...)
}

func (app *App) Head[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error), endpointOpts ...endpointOpt) {
	app.Register(http.MethodHead, path, handlerFunc, endpointOpts...)
}

func (app *App) Patch[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error), endpointOpts ...endpointOpt) {
	app.Register(http.MethodPatch, path, handlerFunc, endpointOpts...)
}

func (app *App) Trace[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error), endpointOpts ...endpointOpt) {
	app.Register(http.MethodTrace, path, handlerFunc, endpointOpts...)
}

func (app *App) Query[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error), endpointOpts ...endpointOpt) {
	app.Register("QUERY", path, handlerFunc, endpointOpts...)
}

func (app *App) Register[In, Out any](method string, path string, handlerFunc func(context.Context, In) (Out, error), endpointOpts ...endpointOpt) {
	inType, outType, err := parseInOutTypes[In, Out](method, path)
	if err != nil {
		panic(err)
	}

	opts := slices.Concat(app.endpointOpts, endpointOpts)

	app.addEndpoint(method, path, app.wrapHandlerFunc(handlerFunc), inType, outType, app.middlewares, opts...)
}

func (app *App) GetStd(path string, handlerFunc HandlerFunc, endpointOpts ...endpointOpt) {
	app.RegisterStd(http.MethodGet, path, handlerFunc, endpointOpts...)
}

func (app *App) PutStd(path string, handlerFunc HandlerFunc, endpointOpts ...endpointOpt) {
	app.RegisterStd(http.MethodPut, path, handlerFunc, endpointOpts...)
}

func (app *App) PostStd(path string, handlerFunc HandlerFunc, endpointOpts ...endpointOpt) {
	app.RegisterStd(http.MethodPost, path, handlerFunc, endpointOpts...)
}

func (app *App) DeleteStd(path string, handlerFunc HandlerFunc, endpointOpts ...endpointOpt) {
	app.RegisterStd(http.MethodDelete, path, handlerFunc, endpointOpts...)
}

func (app *App) OptionsStd(path string, handlerFunc HandlerFunc, endpointOpts ...endpointOpt) {
	app.RegisterStd(http.MethodOptions, path, handlerFunc, endpointOpts...)
}

func (app *App) HeadStd(path string, handlerFunc HandlerFunc, endpointOpts ...endpointOpt) {
	app.RegisterStd(http.MethodHead, path, handlerFunc, endpointOpts...)
}

func (app *App) PatchStd(path string, handlerFunc HandlerFunc, endpointOpts ...endpointOpt) {
	app.RegisterStd(http.MethodPatch, path, handlerFunc, endpointOpts...)
}

func (app *App) TraceStd(path string, handlerFunc HandlerFunc, endpointOpts ...endpointOpt) {
	app.RegisterStd(http.MethodTrace, path, handlerFunc, endpointOpts...)
}

func (app *App) QueryStd(path string, handlerFunc HandlerFunc, endpointOpts ...endpointOpt) {
	app.RegisterStd("QUERY", path, handlerFunc, endpointOpts...)
}

func (app *App) RegisterStd(method string, path string, handlerFunc HandlerFunc, endpointOpts ...endpointOpt) {
	emptyType := reflect.TypeFor[Empty]()

	opts := slices.Concat(app.endpointOpts, endpointOpts)

	app.addEndpoint(method, path, handlerFunc, emptyType, emptyType, app.middlewares, opts...)
}

func (app *App) Run(addr string) error {
	start := time.Now()
	mux := app.intoMux()
	slog.Debug("generated HTTP ServeMux", slog.Duration("duration", time.Since(start)))

	return http.ListenAndServe(addr, mux)
}

func (app *App) RunTLS(addr string, certFile string, keyFile string) error {
	start := time.Now()
	mux := app.intoMux()
	slog.Debug("generated HTTP ServeMux", slog.Duration("duration", time.Since(start)))

	return http.ListenAndServeTLS(addr, certFile, keyFile, mux)
}

func (app *App) addEndpoint(method string, path string, handlerFunc HandlerFunc, inType reflect.Type, outType reflect.Type, middlewares []Middleware, endpointOpts ...endpointOpt) {
	ep := Endpoint{
		method:      method,
		path:        path,
		handler:     handlerFunc,
		middlewares: middlewares,
		inType:      inType,
		outType:     outType,
	}

	for _, opt := range endpointOpts {
		opt(&ep)
	}

	app.endpoints = append(app.endpoints, ep)
	slog.Debug("registered endpoint", slog.String("method", method), slog.String("path", path))
}

func (app *App) wrapHandlerFunc[In, Out any](handlerFunc func(context.Context, In) (Out, error)) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		var in In
		err := app.binder.BindIn(r, &in)
		if err != nil {
			return Error(http.StatusBadRequest, "Bad Request", err)
		}

		out, err := handlerFunc(r.Context(), in)
		if err != nil {
			return err
		}

		err = app.binder.BindOut(w, r, http.StatusOK, out)
		if err != nil {
			return Error(http.StatusInternalServerError, "Response write failed", err)
		}

		return nil
	}
}

func (app *App) intoMux() *http.ServeMux {
	mux := http.NewServeMux()
	methodsByPath := make(map[string][]string, len(app.endpoints))

	for _, ep := range app.endpoints {
		pattern := fmt.Sprintf("%s %s", ep.method, ep.path)

		handler := ep.handler
		for _, mw := range ep.middlewares {
			handler = mw(handler)
		}

		mux.Handle(pattern, app.intoHTTPHandler(handler))
		methodsByPath[ep.path] = append(methodsByPath[ep.path], ep.method)

		if app.oapi.isEnabled {
			app.oapi.addEndpoint(ep)
		}
	}

	for path, methods := range methodsByPath {
		hasOptions := slices.ContainsFunc(methods, func(method string) bool {
			return method == http.MethodOptions
		})
		if hasOptions {
			continue
		}

		mux.Handle(fmt.Sprintf("%s %s", http.MethodOptions, path), preflightHandler(methods))
	}

	if app.oapi.isEnabled {
		mux.Handle(fmt.Sprintf("%s %s", http.MethodGet, app.oapi.schemaPath), app.oapi.schemaHandler())
		mux.Handle(fmt.Sprintf("%s %s", http.MethodGet, app.oapi.uiPath), app.oapi.uiHandler())
	}

	return mux
}

func (app *App) intoHTTPHandler(handler HandlerFunc) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		err := handler(w, r)
		if err != nil {
			app.ErrorHandler(w, r, err)
		}
	})
}

func parseInOutTypes[In, Out any](method string, path string) (inType reflect.Type, outType reflect.Type, err error) {
	inType = reflect.TypeFor[In]()
	if inType.Kind() != reflect.Struct {
		return inType, outType, fmt.Errorf("aio: In type has to be a struct")
	}

	missing := checkPathParams(path, inType)
	if len(missing) > 0 {
		return inType, outType, fmt.Errorf("aio: %s %s: In type %s has no field with path tag for path params: %s", method, path, inType, strings.Join(missing, ", "))
	}

	outType = reflect.TypeFor[Out]()
	if outType.Kind() != reflect.Struct {
		return inType, outType, fmt.Errorf("aio: Out type has to be a struct")
	}

	return inType, outType, nil
}

func preflightHandler(methods []string) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Allow", strings.Join(methods, ", "))
		w.WriteHeader(http.StatusNoContent)
	})
}
