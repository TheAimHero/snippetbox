package main

import (
	"bytes"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/TheAimHero/sb/internal/assert"
)

func TestSecureHeaders(t *testing.T) {
	rr := httptest.NewRecorder()
	r, err := http.NewRequest(http.MethodGet, "/", nil)
	if err != nil {
		t.Fatal(err)
	}
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte("OK"))
	})
	secureHeaders(next).ServeHTTP(rr, r)
	tt := []struct {
		key      string
		expected string
	}{
		{key: "Content-Security-Policy", expected: "default-src 'self'; style-src 'self' fonts.googleapis.com; font-src fonts.gstatic.com"},
		{key: "Referrer-Policy", expected: "origin-when-cross-origin"},
		{key: "X-Content-Type-Options", expected: "nosniff"},
		{key: "X-Frame-Options", expected: "deny"},
		{key: "X-XSS-Protection", expected: "0"},
	}
	rs := rr.Result()
	for _, tt := range tt {
		assert.Equal(t, rs.Header.Get(tt.key), tt.expected)
	}
	assert.Equal(t, rs.StatusCode, http.StatusOK)
	defer rs.Body.Close()
	body, err := io.ReadAll(rs.Body)
	if err != nil {
		t.Fatal(err)
	}
	bytes.TrimSpace(body)
	assert.Equal(t, string(body), "OK")
}
