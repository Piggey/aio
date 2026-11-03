package aio

import (
	"context"
	"net/http"
	pathpkg "path"
	"reflect"
	"slices"
)

type Group struct {
	app          *App
	parent       *Group
	path         string
	middlewares  []Middleware
	endpointOpts []endpointOpt
}

func (g *Group) Group(path string) *Group {
	child := Group{
		app:          g.app,
		parent:       g,
		path:         pathpkg.Join(g.path, path),
		middlewares:  slices.Clone(g.middlewares),
		endpointOpts: slices.Clone(g.endpointOpts),
	}

	g.app.groups = append(g.app.groups, child)
	return &child
}

// UseMiddleware adds middlewares to the endpoints and child groups registered after the call.
func (g *Group) UseMiddleware(middlewares ...Middleware) {
	g.middlewares = append(g.middlewares, middlewares...)
}

// UseEndpointOptions adds endpointOpts to the endpoints and child groups registered after the call.
// They are applied before the options passed when registering the endpoint.
func (g *Group) UseEndpointOptions(endpointOpts ...endpointOpt) {
	g.endpointOpts = append(g.endpointOpts, endpointOpts...)
}

func (g *Group) Get[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error), endpointOpts ...endpointOpt) {
	g.Register(http.MethodGet, path, handlerFunc, endpointOpts...)
}

func (g *Group) Put[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error), endpointOpts ...endpointOpt) {
	g.Register(http.MethodPut, path, handlerFunc, endpointOpts...)
}

func (g *Group) Post[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error), endpointOpts ...endpointOpt) {
	g.Register(http.MethodPost, path, handlerFunc, endpointOpts...)
}

func (g *Group) Delete[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error), endpointOpts ...endpointOpt) {
	g.Register(http.MethodDelete, path, handlerFunc, endpointOpts...)
}

func (g *Group) Options[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error), endpointOpts ...endpointOpt) {
	g.Register(http.MethodOptions, path, handlerFunc, endpointOpts...)
}

func (g *Group) Head[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error), endpointOpts ...endpointOpt) {
	g.Register(http.MethodHead, path, handlerFunc, endpointOpts...)
}

func (g *Group) Patch[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error), endpointOpts ...endpointOpt) {
	g.Register(http.MethodPatch, path, handlerFunc, endpointOpts...)
}

func (g *Group) Trace[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error), endpointOpts ...endpointOpt) {
	g.Register(http.MethodTrace, path, handlerFunc, endpointOpts...)
}

func (g *Group) Query[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error), endpointOpts ...endpointOpt) {
	g.Register("QUERY", path, handlerFunc, endpointOpts...)
}

func (g *Group) Register[In, Out any](method string, path string, handlerFunc func(context.Context, In) (Out, error), endpointOpts ...endpointOpt) {
	fullpath := pathpkg.Join(g.path, path)

	inType, outType, err := parseInOutTypes[In, Out](method, fullpath)
	if err != nil {
		panic(err)
	}

	opts := slices.Concat(g.endpointOpts, endpointOpts)

	g.app.addEndpoint(method, fullpath, g.app.wrapHandlerFunc(handlerFunc), inType, outType, g.middlewares, opts...)
}

func (g *Group) GetStd(path string, handlerFunc HandlerFunc, endpointOpts ...endpointOpt) {
	g.RegisterStd(http.MethodGet, path, handlerFunc, endpointOpts...)
}

func (g *Group) PutStd(path string, handlerFunc HandlerFunc, endpointOpts ...endpointOpt) {
	g.RegisterStd(http.MethodPut, path, handlerFunc, endpointOpts...)
}

func (g *Group) PostStd(path string, handlerFunc HandlerFunc, endpointOpts ...endpointOpt) {
	g.RegisterStd(http.MethodPost, path, handlerFunc, endpointOpts...)
}

func (g *Group) DeleteStd(path string, handlerFunc HandlerFunc, endpointOpts ...endpointOpt) {
	g.RegisterStd(http.MethodDelete, path, handlerFunc, endpointOpts...)
}

func (g *Group) OptionsStd(path string, handlerFunc HandlerFunc, endpointOpts ...endpointOpt) {
	g.RegisterStd(http.MethodOptions, path, handlerFunc, endpointOpts...)
}

func (g *Group) HeadStd(path string, handlerFunc HandlerFunc, endpointOpts ...endpointOpt) {
	g.RegisterStd(http.MethodHead, path, handlerFunc, endpointOpts...)
}

func (g *Group) PatchStd(path string, handlerFunc HandlerFunc, endpointOpts ...endpointOpt) {
	g.RegisterStd(http.MethodPatch, path, handlerFunc, endpointOpts...)
}

func (g *Group) TraceStd(path string, handlerFunc HandlerFunc, endpointOpts ...endpointOpt) {
	g.RegisterStd(http.MethodTrace, path, handlerFunc, endpointOpts...)
}

func (g *Group) QueryStd(path string, handlerFunc HandlerFunc, endpointOpts ...endpointOpt) {
	g.RegisterStd("QUERY", path, handlerFunc, endpointOpts...)
}

func (g *Group) RegisterStd(method string, path string, handlerFunc HandlerFunc, endpointOpts ...endpointOpt) {
	emptyType := reflect.TypeFor[Empty]()

	fullpath := pathpkg.Join(g.path, path)
	opts := slices.Concat(g.endpointOpts, endpointOpts)

	g.app.addEndpoint(method, fullpath, handlerFunc, emptyType, emptyType, g.middlewares, opts...)
}
