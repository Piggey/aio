package aio

import (
	"context"
	"net/http"
	pathpkg "path"
	"slices"
)

type Group struct {
	app         *App
	parent      *Group
	path        string
	middlewares []Middleware
}

func (g *Group) Group(path string) *Group {
	child := Group{
		app:         g.app,
		parent:      g,
		path:        pathpkg.Join(g.path, path),
		middlewares: slices.Clone(g.middlewares),
	}

	g.app.groups = append(g.app.groups, child)

	return &child
}

func (g *Group) Use(middleware Middleware) {
	g.middlewares = append(g.middlewares, middleware)
}

func (g *Group) Get[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error), opts ...endpointOpt) {
	g.Register(http.MethodGet, path, handlerFunc, opts...)
}

func (g *Group) Put[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error), opts ...endpointOpt) {
	g.Register(http.MethodPut, path, handlerFunc, opts...)
}

func (g *Group) Post[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error), opts ...endpointOpt) {
	g.Register(http.MethodPost, path, handlerFunc, opts...)
}

func (g *Group) Delete[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error), opts ...endpointOpt) {
	g.Register(http.MethodDelete, path, handlerFunc, opts...)
}

func (g *Group) Options[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error), opts ...endpointOpt) {
	g.Register(http.MethodOptions, path, handlerFunc, opts...)
}

func (g *Group) Head[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error), opts ...endpointOpt) {
	g.Register(http.MethodHead, path, handlerFunc, opts...)
}

func (g *Group) Patch[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error), opts ...endpointOpt) {
	g.Register(http.MethodPatch, path, handlerFunc, opts...)
}

func (g *Group) Trace[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error), opts ...endpointOpt) {
	g.Register(http.MethodTrace, path, handlerFunc, opts...)
}

func (g *Group) Query[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error), opts ...endpointOpt) {
	g.Register("QUERY", path, handlerFunc, opts...)
}

func (g *Group) Register[In, Out any](method string, path string, handlerFunc func(context.Context, In) (Out, error), opts ...endpointOpt) {
	g.app.addEndpoint(method, pathpkg.Join(g.path, path), handlerFunc, g.middlewares, opts...)
}
