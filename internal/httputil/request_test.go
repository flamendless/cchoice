package httputil

import (
	"crypto/tls"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRequestIsHTTPS(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		req  *http.Request
		want bool
	}{
		{
			name: "tls connection",
			req:  &http.Request{TLS: &tls.ConnectionState{}},
			want: true,
		},
		{
			name: "forwarded proto https",
			req: func() *http.Request {
				r := &http.Request{Header: http.Header{}}
				r.Header.Set("X-Forwarded-Proto", "https")
				return r
			}(),
			want: true,
		},
		{
			name: "forwarded proto http",
			req: func() *http.Request {
				r := &http.Request{Header: http.Header{}}
				r.Header.Set("X-Forwarded-Proto", "http")
				return r
			}(),
			want: false,
		},
		{
			name: "plain http",
			req:  &http.Request{},
			want: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, RequestIsHTTPS(tt.req))
		})
	}
}
