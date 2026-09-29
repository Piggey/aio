package aio

import (
	"net/http"

	"github.com/Piggey/aio/internal/openapi"
	"github.com/Piggey/binder"
)

type App struct {
	mux    *http.ServeMux
	binder *binder.Binder
	oapi   openapi.OpenAPI

	middlewares []http.Handler
	endpoints   []endpoint
	groups      []Group
}

func New(name string, appOpts ...appOpt) *App {
	app := App{
		mux:    http.NewServeMux(),
		binder: nil,
		oapi: openapi.OpenAPI{
			OpenAPI: openapi.OpenAPIVersion,
			Info: openapi.Info{
				Title: name,
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

func (app *App) Use(middleware http.Handler) {
	app.middlewares = append(app.middlewares, middleware)
}

func (app *App) Group(path string) *Group {
	group := Group{}
	app.groups = append(app.groups, group)
	return &group
}

func (app *App) Get[In, Out any](path string, handlerFunc HandlerFunc[In, Out]) {
	app.Register(http.MethodGet, path, handlerFunc)
}

func (app *App) Put[In, Out any](path string, handlerFunc HandlerFunc[In, Out]) {
	app.Register(http.MethodPut, path, handlerFunc)
}

func (app *App) Post[In, Out any](path string, handlerFunc HandlerFunc[In, Out]) {
	app.Register(http.MethodPost, path, handlerFunc)
}

func (app *App) Delete[In, Out any](path string, handlerFunc HandlerFunc[In, Out]) {
	app.Register(http.MethodDelete, path, handlerFunc)
}

func (app *App) Options[In, Out any](path string, handlerFunc HandlerFunc[In, Out]) {
	app.Register(http.MethodOptions, path, handlerFunc)
}

func (app *App) Head[In, Out any](path string, handlerFunc HandlerFunc[In, Out]) {
	app.Register(http.MethodHead, path, handlerFunc)
}

func (app *App) Patch[In, Out any](path string, handlerFunc HandlerFunc[In, Out]) {
	app.Register(http.MethodPatch, path, handlerFunc)
}

func (app *App) Trace[In, Out any](path string, handlerFunc HandlerFunc[In, Out]) {
	app.Register(http.MethodTrace, path, handlerFunc)
}

func (app *App) Query[In, Out any](path string, handlerFunc HandlerFunc[In, Out]) {
	app.Register("QUERY", path, handlerFunc)
}

func (app *App) Register[In, Out any](method string, path string, handlerFunc HandlerFunc[In, Out]) {
	pattern := method + " " + path
	app.mux.HandleFunc(pattern, wrapHandlerFunc(handlerFunc))
}

func wrapHandlerFunc[In, Out any](handlerFunc HandlerFunc[In, Out]) func(http.ResponseWriter, *http.Request) {
	panic("unimplemented")
}

func (app *App) ListenAndServe(addr string) error {
	return http.ListenAndServe(addr, s.mux)
}

func (app *App) ListenAndServeTLS(addr string, certFile string, keyFile string) error {
	return http.ListenAndServeTLS(addr, certFile, keyFile, app.mux)
}
