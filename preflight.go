package aio

import (
	"net/http"
	"slices"
	"strings"
)

// preflightHandler answers OPTIONS requests for a path that has no OPTIONS endpoint.
// A CORS preflight is passed through the middlewares of the endpoint named in Access-Control-Request-Method,
// any other OPTIONS request gets 204 No Content with the Allow header.
func preflightHandler(eps []endpoint) http.Handler {
	allow := allowHeader(eps)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		method := r.Header.Get("Access-Control-Request-Method")
		i := slices.IndexFunc(eps, func(ep endpoint) bool {
			return ep.method == method
		})
		if i == -1 {
			w.Header().Set("Allow", allow)
			w.WriteHeader(http.StatusNoContent)
			return
		}

		eps[i].preflight.ServeHTTP(w, r)
	})
}

func allowHeader(eps []endpoint) string {
	methods := []string{http.MethodOptions}
	for _, ep := range eps {
		methods = append(methods, ep.method)
	}

	slices.Sort(methods)
	methods = slices.Compact(methods)

	return strings.Join(methods, ", ")
}

func noContent(w http.ResponseWriter, r *http.Request) error {
	w.WriteHeader(http.StatusNoContent)
	return nil
}
