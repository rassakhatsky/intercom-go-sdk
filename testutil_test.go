package intercom

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

// setup returns a Client configured to talk to a local httptest.Server,
// the ServeMux the server routes through (so tests can register handlers),
// and a teardown function the caller must defer.
func setup() (client *Client, mux *http.ServeMux, teardown func()) {
	mux = http.NewServeMux()
	server := httptest.NewServer(mux)
	client = NewClient("test-token", WithBaseURL(server.URL+"/"))
	return client, mux, server.Close
}

// testMethod checks that the incoming request uses the expected HTTP method.
func testMethod(t *testing.T, r *http.Request, want string) {
	t.Helper()
	if got := r.Method; got != want {
		t.Errorf("Request method = %v, want %v", got, want)
	}
}

// testHeader checks that the incoming request has the expected header value.
func testHeader(t *testing.T, r *http.Request, header, want string) {
	t.Helper()
	if got := r.Header.Get(header); got != want {
		t.Errorf("Header %v = %q, want %q", header, got, want)
	}
}
