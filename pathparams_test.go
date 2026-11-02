package aio

import (
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPathParams(t *testing.T) {
	tests := []struct {
		name string
		path string
		want []string
	}{
		{
			name: "brak parametrow",
			path: "/users",
			want: []string{},
		},
		{
			name: "kilka parametrow",
			path: "/users/{userID}/orders/{orderID}",
			want: []string{"userID", "orderID"},
		},
		{
			name: "parametr wielosegmentowy bez kropek",
			path: "/files/{path...}",
			want: []string{"path"},
		},
		{
			name: "{$} nie jest parametrem",
			path: "/users/{id}/{$}",
			want: []string{"id"},
		},
		{
			name: "klamry w srodku segmentu sa pomijane",
			path: "/users/a{id}b",
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, pathParams(tt.path))
		})
	}
}

func TestMissingPathParams(t *testing.T) {
	type withID struct {
		ID   int    `path:"id"`
		Name string `query:"name"`
	}

	type withPrivateID struct {
		id int `path:"id"`
	}

	tests := []struct {
		name   string
		path   string
		inType reflect.Type
		want   []string
	}{
		{
			name:   "wszystkie parametry sa w strukturze",
			path:   "/users/{id}",
			inType: reflect.TypeFor[withID](),
			want:   []string{},
		},
		{
			name:   "brakujacy parametr",
			path:   "/users/{id}/orders/{orderID}",
			inType: reflect.TypeFor[withID](),
			want:   []string{"orderID"},
		},
		{
			name:   "pusta struktura",
			path:   "/users/{id}",
			inType: reflect.TypeFor[struct{}](),
			want:   []string{"id"},
		},
		{
			name:   "prywatne pole sie nie liczy",
			path:   "/users/{id}",
			inType: reflect.TypeFor[withPrivateID](),
			want:   []string{"id"},
		},
		{
			name:   "tag query o tej samej nazwie sie nie liczy",
			path:   "/users/{name}",
			inType: reflect.TypeFor[withID](),
			want:   []string{"name"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, checkPathParams(tt.path, tt.inType))
		})
	}
}
