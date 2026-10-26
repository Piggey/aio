package middleware

import (
	"net/http"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestIsOriginAllowed(t *testing.T) {
	tests := []struct {
		name           string
		allowedOrigins []string
		origin         string
		want           bool
	}{
		{
			name:           "origin z listy",
			allowedOrigins: []string{"https://a.pl", "https://b.pl"},
			origin:         "https://b.pl",
			want:           true,
		},
		{
			name:           "origin spoza listy",
			allowedOrigins: []string{"https://a.pl"},
			origin:         "https://evil.pl",
			want:           false,
		},
		{
			name:           "porownanie bez wielkosci liter",
			allowedOrigins: []string{"https://A.pl"},
			origin:         "https://a.pl",
			want:           true,
		},
		{
			name:           "gwiazdka dopuszcza kazdy origin",
			allowedOrigins: []string{"*"},
			origin:         "https://evil.pl",
			want:           true,
		},
		{
			name:           "brak naglowka Origin",
			allowedOrigins: []string{"*"},
			origin:         "",
			want:           false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, isOriginAllowed(tt.allowedOrigins, tt.origin))
		})
	}
}

func TestCorsHeaders(t *testing.T) {
	tests := []struct {
		name   string
		cfg    CORSConfig
		origin string
		want   http.Header
	}{
		{
			name:   "brak originow wylacza CORS",
			cfg:    CORSConfig{},
			origin: "https://a.pl",
			want:   http.Header{},
		},
		{
			name:   "niedozwolony origin dostaje tylko Vary",
			cfg:    CORSConfig{AllowedOrigins: []string{"https://a.pl"}},
			origin: "https://evil.pl",
			want:   http.Header{"Vary": {"Origin"}},
		},
		{
			name:   "dozwolony origin jest odbijany",
			cfg:    CORSConfig{AllowedOrigins: []string{"https://a.pl"}},
			origin: "https://a.pl",
			want: http.Header{
				"Vary":                        {"Origin"},
				"Access-Control-Allow-Origin": {"https://a.pl"},
			},
		},
		{
			name:   "gwiazdka bez credentials daje *",
			cfg:    CORSConfig{AllowedOrigins: []string{"*"}},
			origin: "https://a.pl",
			want: http.Header{
				"Vary":                        {"Origin"},
				"Access-Control-Allow-Origin": {"*"},
			},
		},
		{
			name: "gwiazdka z credentials odbija origin",
			cfg: CORSConfig{
				AllowedOrigins:   []string{"*"},
				AllowCredentials: true,
				ExposedHeaders:   []string{"X-Total", "X-Page"},
			},
			origin: "https://a.pl",
			want: http.Header{
				"Vary":                             {"Origin"},
				"Access-Control-Allow-Origin":      {"https://a.pl"},
				"Access-Control-Allow-Credentials": {"true"},
				"Access-Control-Expose-Headers":    {"X-Total, X-Page"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, corsHeaders(tt.cfg, tt.origin))
		})
	}
}

func TestPreflightHeaders(t *testing.T) {
	vary := []string{"Origin", "Access-Control-Request-Method", "Access-Control-Request-Headers"}

	tests := []struct {
		name           string
		cfg            CORSConfig
		origin         string
		requestMethod  string
		requestHeaders string
		want           http.Header
	}{
		{
			name:          "brak originow wylacza CORS",
			cfg:           CORSConfig{AllowedMethods: defaultCORSMethods},
			origin:        "https://a.pl",
			requestMethod: http.MethodGet,
			want:          http.Header{},
		},
		{
			name:          "niedozwolony origin dostaje tylko Vary",
			cfg:           CORSConfig{AllowedOrigins: []string{"https://a.pl"}, AllowedMethods: defaultCORSMethods},
			origin:        "https://evil.pl",
			requestMethod: http.MethodGet,
			want:          http.Header{"Vary": vary},
		},
		{
			name:          "niedozwolona metoda dostaje tylko Vary",
			cfg:           CORSConfig{AllowedOrigins: []string{"https://a.pl"}, AllowedMethods: []string{http.MethodGet}},
			origin:        "https://a.pl",
			requestMethod: http.MethodDelete,
			want:          http.Header{"Vary": vary},
		},
		{
			name:           "naglowki zadania sa odbijane gdy brak AllowedHeaders",
			cfg:            CORSConfig{AllowedOrigins: []string{"https://a.pl"}, AllowedMethods: []string{http.MethodGet, http.MethodPost}},
			origin:         "https://a.pl",
			requestMethod:  http.MethodPost,
			requestHeaders: "Content-Type, Authorization",
			want: http.Header{
				"Vary":                         vary,
				"Access-Control-Allow-Origin":  {"https://a.pl"},
				"Access-Control-Allow-Methods": {"GET, POST"},
				"Access-Control-Allow-Headers": {"Content-Type, Authorization"},
			},
		},
		{
			name: "pelna konfiguracja",
			cfg: CORSConfig{
				AllowedOrigins:   []string{"https://a.pl"},
				AllowedMethods:   []string{http.MethodGet},
				AllowedHeaders:   []string{"Content-Type"},
				AllowCredentials: true,
				MaxAge:           10 * time.Minute,
			},
			origin:         "https://a.pl",
			requestMethod:  http.MethodGet,
			requestHeaders: "Authorization",
			want: http.Header{
				"Vary":                             vary,
				"Access-Control-Allow-Origin":      {"https://a.pl"},
				"Access-Control-Allow-Methods":     {"GET"},
				"Access-Control-Allow-Headers":     {"Content-Type"},
				"Access-Control-Allow-Credentials": {"true"},
				"Access-Control-Max-Age":           {"600"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := preflightHeaders(tt.cfg, tt.origin, tt.requestMethod, tt.requestHeaders)
			assert.Equal(t, tt.want, got)
		})
	}
}
