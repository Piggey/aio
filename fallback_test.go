package aio

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestHandleIgnoringConflict(t *testing.T) {
	tests := []struct {
		name        string
		patterns    []string
		method      string
		target      string
		wantPattern string
	}{
		{
			name:        "wzorzec z metoda wygrywa ze sciezka bez metody",
			patterns:    []string{"GET /users", "/users", "/"},
			method:      http.MethodGet,
			target:      "/users",
			wantPattern: "GET /users",
		},
		{
			name:        "inna metoda trafia do sciezki bez metody",
			patterns:    []string{"GET /users", "/users", "/"},
			method:      http.MethodDelete,
			target:      "/users",
			wantPattern: "/users",
		},
		{
			name:        "HEAD trafia do wzorca GET",
			patterns:    []string{"GET /users", "/users", "/"},
			method:      http.MethodHead,
			target:      "/users",
			wantPattern: "GET /users",
		},
		{
			name:        "niezarejestrowana sciezka trafia do catch-all",
			patterns:    []string{"GET /users", "/users", "/"},
			method:      http.MethodGet,
			target:      "/orders",
			wantPattern: "/",
		},
		{
			name:        "konflikt nazw parametrow nie panikuje, wygrywa pierwszy wzorzec",
			patterns:    []string{"/users/{id}", "/users/{name}"},
			method:      http.MethodGet,
			target:      "/users/5",
			wantPattern: "/users/{id}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mux := http.NewServeMux()
			for _, pattern := range tt.patterns {
				handleIgnoringConflict(mux, pattern, http.NotFoundHandler())
			}

			req, err := http.NewRequest(tt.method, tt.target, nil)
			assert.NoError(t, err)

			_, pattern := mux.Handler(req)
			assert.Equal(t, tt.wantPattern, pattern)
		})
	}
}

func TestApplyMiddlewares(t *testing.T) {
	calls := []string{}
	middleware := func(name string) Middleware {
		return func(next HandlerFunc) HandlerFunc {
			return func(w http.ResponseWriter, r *http.Request) error {
				calls = append(calls, name)
				return next(w, r)
			}
		}
	}

	handler := applyMiddlewares(errorHandler(ErrNotFound), []Middleware{middleware("first"), middleware("second")})
	err := handler(nil, nil)

	assert.Equal(t, ErrNotFound, err)
	// last registered is the outermost one, like in addEndpoint
	assert.Equal(t, []string{"second", "first"}, calls)
}
