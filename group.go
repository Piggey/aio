package aio

import (
	"context"
	"net/http"
	pathpkg "path"
)

type Group struct {
	app         *App
	parent      *Group
	path        string
	middlewares []func(http.Handler) http.Handler
}

func (g *Group) Group(path string) *Group {
	child := Group{
		app:         g.app,
		parent:      g,
		path:        pathpkg.Join(g.path, path),
		middlewares: []func(http.Handler) http.Handler{},
	}

	g.app.groups = append(g.app.groups, child)
	copy(child.middlewares, g.middlewares)

	return &child
}

func (g *Group) Use(middleware func(http.Handler) http.Handler) {
	g.middlewares = append(g.middlewares, middleware)
}

func (g *Group) Get[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error)) {
	g.Register(http.MethodGet, path, handlerFunc)
}

func (g *Group) Put[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error)) {
	g.Register(http.MethodPut, path, handlerFunc)
}

func (g *Group) Post[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error)) {
	g.Register(http.MethodPost, path, handlerFunc)
}

func (g *Group) Delete[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error)) {
	g.Register(http.MethodDelete, path, handlerFunc)
}

func (g *Group) Options[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error)) {
	g.Register(http.MethodOptions, path, handlerFunc)
}

func (g *Group) Head[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error)) {
	g.Register(http.MethodHead, path, handlerFunc)
}

func (g *Group) Patch[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error)) {
	g.Register(http.MethodPatch, path, handlerFunc)
}

func (g *Group) Trace[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error)) {
	g.Register(http.MethodTrace, path, handlerFunc)
}

func (g *Group) Query[In, Out any](path string, handlerFunc func(context.Context, In) (Out, error)) {
	g.Register("QUERY", path, handlerFunc)
}

func (g *Group) Register[In, Out any](method string, path string, handlerFunc func(context.Context, In) (Out, error)) {
	g.app.addEndpoint(method, pathpkg.Join(g.path, path), handlerFunc, g.middlewares)
}
