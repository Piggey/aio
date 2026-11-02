package middleware

import (
	"net/http"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/Piggey/aio"
)

type CORSConfig struct {
	// AllowedOrigins lists origins allowed to make cross-origin requests, "*" allows any.
	// When empty, CORS is disabled and no headers are set.
	AllowedOrigins []string
	// AllowedMethods defaults to GET, HEAD, POST, PUT, PATCH and DELETE.
	AllowedMethods []string
	// AllowedHeaders defaults to the headers requested in Access-Control-Request-Headers.
	AllowedHeaders   []string
	ExposedHeaders   []string
	AllowCredentials bool
	// MaxAge is how long the preflight response may be cached, 0 leaves it to the browser.
	MaxAge time.Duration
}

var defaultCORSMethods = []string{
	http.MethodGet,
	http.MethodHead,
	http.MethodPost,
	http.MethodPut,
	http.MethodPatch,
	http.MethodDelete,
}

// CORS sets the CORS headers and answers preflight requests with 204 No Content,
// without calling the rest of the chain.
func CORS(cfg CORSConfig) aio.Middleware {
	if len(cfg.AllowedMethods) == 0 {
		cfg.AllowedMethods = defaultCORSMethods
	}

	return func(next aio.HandlerFunc) aio.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) error {
			origin := r.Header.Get("Origin")
			requestMethod := r.Header.Get("Access-Control-Request-Method")

			if r.Method == http.MethodOptions && requestMethod != "" {
				requestHeaders := r.Header.Get("Access-Control-Request-Headers")
				headers := preflightHeaders(cfg, origin, requestMethod, requestHeaders)
				addHeaders(w.Header(), headers)
				w.WriteHeader(http.StatusNoContent)
				return nil
			}

			headers := corsHeaders(cfg, origin)
			addHeaders(w.Header(), headers)

			return next(w, r)
		}
	}
}

func corsHeaders(cfg CORSConfig, origin string) http.Header {
	headers := http.Header{}
	if len(cfg.AllowedOrigins) == 0 {
		return headers
	}

	headers.Add("Vary", "Origin")
	if !isOriginAllowed(cfg.AllowedOrigins, origin) {
		return headers
	}

	allowOrigin := allowOriginValue(cfg, origin)
	headers.Set("Access-Control-Allow-Origin", allowOrigin)
	if cfg.AllowCredentials {
		headers.Set("Access-Control-Allow-Credentials", "true")
	}
	if len(cfg.ExposedHeaders) > 0 {
		headers.Set("Access-Control-Expose-Headers", strings.Join(cfg.ExposedHeaders, ", "))
	}

	return headers
}

func preflightHeaders(cfg CORSConfig, origin string, requestMethod string, requestHeaders string) http.Header {
	headers := http.Header{}
	if len(cfg.AllowedOrigins) == 0 {
		return headers
	}

	headers.Add("Vary", "Origin")
	headers.Add("Vary", "Access-Control-Request-Method")
	headers.Add("Vary", "Access-Control-Request-Headers")
	if !isOriginAllowed(cfg.AllowedOrigins, origin) || !slices.Contains(cfg.AllowedMethods, requestMethod) {
		return headers
	}

	allowOrigin := allowOriginValue(cfg, origin)
	headers.Set("Access-Control-Allow-Origin", allowOrigin)
	headers.Set("Access-Control-Allow-Methods", strings.Join(cfg.AllowedMethods, ", "))

	allowHeaders := requestHeaders
	if len(cfg.AllowedHeaders) > 0 {
		allowHeaders = strings.Join(cfg.AllowedHeaders, ", ")
	}
	if allowHeaders != "" {
		headers.Set("Access-Control-Allow-Headers", allowHeaders)
	}

	if cfg.AllowCredentials {
		headers.Set("Access-Control-Allow-Credentials", "true")
	}
	if cfg.MaxAge > 0 {
		maxAge := strconv.Itoa(int(cfg.MaxAge.Seconds()))
		headers.Set("Access-Control-Max-Age", maxAge)
	}

	return headers
}

func isOriginAllowed(allowedOrigins []string, origin string) bool {
	if origin == "" {
		return false
	}

	return slices.ContainsFunc(allowedOrigins, func(allowed string) bool {
		return allowed == "*" || strings.EqualFold(allowed, origin)
	})
}

// allowOriginValue echoes the origin when credentials are allowed, because browsers reject "*" with credentials.
func allowOriginValue(cfg CORSConfig, origin string) string {
	if slices.Contains(cfg.AllowedOrigins, "*") && !cfg.AllowCredentials {
		return "*"
	}

	return origin
}

func addHeaders(dst http.Header, src http.Header) {
	for key, values := range src {
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}
