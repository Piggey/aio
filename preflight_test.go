package aio

import (
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAllowHeader(t *testing.T) {
	tests := []struct {
		name string
		eps  []endpoint
		want string
	}{
		{
			name: "jedna metoda z dodanym OPTIONS",
			eps:  []endpoint{{method: http.MethodGet}},
			want: "GET, OPTIONS",
		},
		{
			name: "metody posortowane bez duplikatow",
			eps:  []endpoint{{method: http.MethodPost}, {method: http.MethodGet}, {method: http.MethodPost}},
			want: "GET, OPTIONS, POST",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, allowHeader(tt.eps))
		})
	}
}
