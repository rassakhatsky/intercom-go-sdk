package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestDecode_Success(t *testing.T) {
	type contact struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	r := &Result{
		StatusCode: 200,
		Body:       []byte(`{"id":"1","name":"Alice"}`),
	}

	got, err := Decode[contact](r)
	if err != nil {
		t.Fatalf("Decode returned error: %v", err)
	}
	if got.ID != "1" {
		t.Errorf("ID = %q, want %q", got.ID, "1")
	}
	if got.Name != "Alice" {
		t.Errorf("Name = %q, want %q", got.Name, "Alice")
	}
}

func TestDecode_EmptyBodyNon2xx(t *testing.T) {
	type contact struct {
		ID string `json:"id"`
	}

	r := &Result{
		StatusCode: 400,
		Body:       nil,
	}

	got, err := Decode[contact](r)
	if err == nil {
		t.Fatal("expected error for empty body on non-2xx, got nil")
	}
	if got != nil {
		t.Errorf("got = %v, want nil on error", got)
	}
}

func TestDecode_EmptyBody2xx(t *testing.T) {
	type contact struct {
		ID string `json:"id"`
	}

	for _, code := range []int{200, 202, 204} {
		t.Run(fmt.Sprintf("HTTP_%d", code), func(t *testing.T) {
			r := &Result{
				StatusCode: code,
				Body:       nil,
			}

			got, err := Decode[contact](r)
			if err != nil {
				t.Fatalf("Decode returned error: %v", err)
			}
			if got == nil {
				t.Fatal("Decode returned nil, want zero-value pointer")
			}
			if got.ID != "" {
				t.Errorf("ID = %q, want empty", got.ID)
			}
		})
	}
}

func TestDecode_NoContent(t *testing.T) {
	r := &Result{
		StatusCode: http.StatusNoContent,
		Body:       nil,
	}

	type contact struct {
		ID string `json:"id"`
	}

	got, err := Decode[contact](r)
	if err != nil {
		t.Fatalf("Decode returned error: %v", err)
	}
	if got == nil {
		t.Fatal("Decode returned nil, want zero-value pointer")
	}
}

func TestDecode_NilResult(t *testing.T) {
	type contact struct {
		ID string `json:"id"`
	}

	got, err := Decode[contact](nil)
	if err == nil {
		t.Fatal("expected error for nil result, got nil")
	}
	if got != nil {
		t.Errorf("got = %v, want nil on error", got)
	}
}

func TestDecode_MalformedJSON(t *testing.T) {
	r := &Result{
		StatusCode: 200,
		Body:       []byte(`{not json`),
	}

	type contact struct {
		ID string `json:"id"`
	}

	_, err := Decode[contact](r)
	if err == nil {
		t.Fatal("expected error for malformed JSON")
	}
}

func TestBuildResult_200(t *testing.T) {
	body := []byte(`{"id":"1"}`)
	resp := &http.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": {"application/json"}},
		Request:    httptest.NewRequest(http.MethodGet, "https://api.intercom.io/contacts/1", nil),
	}

	r := BuildResult(resp, body)

	if r.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", r.StatusCode)
	}
	if string(r.Body) != `{"id":"1"}` {
		t.Errorf("Body = %q, want %q", string(r.Body), `{"id":"1"}`)
	}
	if r.Error != nil {
		t.Errorf("Error = %v, want nil", r.Error)
	}
	if r.URL != "https://api.intercom.io/contacts/1" {
		t.Errorf("URL = %q, want %q", r.URL, "https://api.intercom.io/contacts/1")
	}
}

func TestBuildResult_404(t *testing.T) {
	body := []byte(`{"type":"error.list","errors":[{"code":"not_found","message":"Not found"}]}`)
	resp := &http.Response{
		StatusCode: 404,
		Header:     http.Header{},
		Request:    httptest.NewRequest(http.MethodGet, "https://api.intercom.io/contacts/999", nil),
	}

	r := BuildResult(resp, body)

	if r.StatusCode != 404 {
		t.Errorf("StatusCode = %d, want 404", r.StatusCode)
	}
	if r.Error == nil {
		t.Fatal("Error is nil, want non-nil")
	}
	if r.Error.Code != "not_found" {
		t.Errorf("Error.Code = %q, want %q", r.Error.Code, "not_found")
	}
	if r.Error.Message != "Not found" {
		t.Errorf("Error.Message = %q, want %q", r.Error.Message, "Not found")
	}
}

func TestBuildResult_404_JSONWithoutErrorsArray(t *testing.T) {
	body := []byte(`{"type":"error","message":"Resource not found"}`)
	resp := &http.Response{
		StatusCode: 404,
		Header:     http.Header{},
		Request:    httptest.NewRequest(http.MethodGet, "https://api.intercom.io/contacts/999", nil),
	}

	r := BuildResult(resp, body)

	if r.Error == nil {
		t.Fatal("Error is nil, want non-nil")
	}
	if r.Error.Message != string(body) {
		t.Errorf("Error.Message = %q, want raw body %q", r.Error.Message, string(body))
	}
}

func TestBuildResult_500_NonJSON(t *testing.T) {
	body := []byte(`Internal Server Error`)
	resp := &http.Response{
		StatusCode: 500,
		Header:     http.Header{},
		Request:    httptest.NewRequest(http.MethodGet, "https://api.intercom.io/contacts", nil),
	}

	r := BuildResult(resp, body)

	if r.StatusCode != 500 {
		t.Errorf("StatusCode = %d, want 500", r.StatusCode)
	}
	if r.Error == nil {
		t.Fatal("Error is nil, want non-nil")
	}
	if r.Error.Message != "Internal Server Error" {
		t.Errorf("Error.Message = %q, want %q", r.Error.Message, "Internal Server Error")
	}
}

func TestBuildResult_NilRequest(t *testing.T) {
	body := []byte(`{"id":"1"}`)
	resp := &http.Response{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": {"application/json"}},
		Request:    nil,
	}

	r := BuildResult(resp, body)

	if r.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", r.StatusCode)
	}
	if r.URL != "" {
		t.Errorf("URL = %q, want empty string", r.URL)
	}
	if string(r.Body) != `{"id":"1"}` {
		t.Errorf("Body = %q, want %q", string(r.Body), `{"id":"1"}`)
	}
	if r.Error != nil {
		t.Errorf("Error = %v, want nil", r.Error)
	}
}

func TestErrorResult_Error(t *testing.T) {
	tests := []struct {
		name string
		err  ErrorResult
		want string
	}{
		{
			name: "with code and message",
			err: ErrorResult{
				Code:    "not_found",
				Message: "Resource not found",
				Errors:  []ErrorDetail{{Code: "not_found", Message: "Resource not found"}},
			},
			want: "not_found: Resource not found",
		},
		{
			name: "message only no code",
			err: ErrorResult{
				Message: "Bad Gateway",
			},
			want: "Bad Gateway",
		},
		{
			name: "empty",
			err:  ErrorResult{},
			want: "unknown API error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.err.Error()
			if got != tt.want {
				t.Errorf("Error() = %q, want %q", got, tt.want)
			}
		})
	}
}
