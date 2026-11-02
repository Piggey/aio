package aio

import (
	"net/http"
)

// handleFallbacks makes mux answer with ProblemDetails instead of its plain text 404 and 405 responses.
// Every path gets a pattern without a method, which answers 405 - patterns with a method are more specific,
// so they still win for the registered methods. The catch-all "/" answers 404.
// Both go through the App middlewares, so they are logged, have CORS headers etc.
func (app *App) handleFallbacks(mux *http.ServeMux, allowByPath map[string]string) {
	notFound := applyMiddlewares(errorHandler(ErrNotFound), app.middlewares)
	mux.Handle("/", app.intoHTTPHandler(notFound))

	for path, allow := range allowByPath {
		methodNotAllowed := applyMiddlewares(methodNotAllowedHandler(allow), app.middlewares)
		handleIgnoringConflict(mux, path, app.intoHTTPHandler(methodNotAllowed))
	}
}

// handleIgnoringConflict skips patterns conflicting with already registered ones, e.g. /a/{name} after /a/{id}
// or "/" after the catch-all - the registered pattern already matches the same requests.
func handleIgnoringConflict(mux *http.ServeMux, pattern string, handler http.Handler) {
	defer func() {
		_ = recover()
	}()

	mux.Handle(pattern, handler)
}

func methodNotAllowedHandler(allow string) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		w.Header().Set("Allow", allow)
		return ErrMethodNotAllowed
	}
}

func errorHandler(err error) HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) error {
		return err
	}
}

func applyMiddlewares(handler HandlerFunc, middlewares []Middleware) HandlerFunc {
	for _, middleware := range middlewares {
		handler = middleware(handler)
	}

	return handler
}
