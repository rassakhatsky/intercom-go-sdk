package intercom

import (
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestBuildResult_PopulatesFields(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("X-Request-Id", "abc123")
		w.WriteHeader(200)
		w.Write([]byte(`{"id":"1"}`))
	}
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	r := buildResult(resp, body)

	if r.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", r.StatusCode)
	}
	if got := r.Header.Get("X-Request-Id"); got != "abc123" {
		t.Errorf("Header X-Request-Id = %q, want %q", got, "abc123")
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
			name: "with errors",
			err: ErrorResult{
				Type:      "error.list",
				RequestID: "req-123",
				Code:      "not_found",
				Message:   "Resource not found",
				Errors:    []ErrorDetail{{Code: "not_found", Message: "Resource not found"}},
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
			name: "empty errors",
			err: ErrorResult{
				Type: "error.list",
			},
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

func TestBuildResult_404_PopulatesError(t *testing.T) {
	handler := func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(404)
		w.Write([]byte(`{"type":"error.list","request_id":"req-456","errors":[{"code":"not_found","message":"Resource not found"}]}`))
	}
	ts := httptest.NewServer(http.HandlerFunc(handler))
	defer ts.Close()

	resp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	r := buildResult(resp, body)

	if r.StatusCode != 404 {
		t.Errorf("StatusCode = %d, want 404", r.StatusCode)
	}
	if r.Error == nil {
		t.Fatal("Error is nil, want non-nil")
	}
	if r.Error.RequestID != "req-456" {
		t.Errorf("Error.RequestID = %q, want %q", r.Error.RequestID, "req-456")
	}
	if len(r.Error.Errors) != 1 {
		t.Fatalf("Error.Errors len = %d, want 1", len(r.Error.Errors))
	}
	if r.Error.Errors[0].Code != "not_found" {
		t.Errorf("Error.Errors[0].Code = %q, want %q", r.Error.Errors[0].Code, "not_found")
	}
}

func TestResult_SuccessAccess(t *testing.T) {
	type testData struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	r := &Result{
		StatusCode: 200,
		Header:     http.Header{"Content-Type": {"application/json"}},
		URL:        "https://api.intercom.io/contacts/1",
		Body:       []byte(`{"id":"1","name":"Alice"}`),
	}

	if r.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", r.StatusCode)
	}

	data, err := Decode[testData](r)
	if err != nil {
		t.Fatalf("Decode returned error: %v", err)
	}
	if data.ID != "1" {
		t.Errorf("Data.ID = %q, want %q", data.ID, "1")
	}
	if data.Name != "Alice" {
		t.Errorf("Data.Name = %q, want %q", data.Name, "Alice")
	}
	if r.Error != nil {
		t.Errorf("Error = %v, want nil", r.Error)
	}
}

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

func TestDecode_EmptyBodyNon204(t *testing.T) {
	type contact struct {
		ID string `json:"id"`
	}

	r := &Result{
		StatusCode: 200,
		Body:       nil,
	}

	got, err := Decode[contact](r)
	if err == nil {
		t.Fatal("expected error for empty body on non-204, got nil")
	}
	if got != nil {
		t.Errorf("got = %v, want nil on error", got)
	}
}

func TestDecode_EmptyBody204(t *testing.T) {
	type contact struct {
		ID string `json:"id"`
	}

	r := &Result{
		StatusCode: http.StatusNoContent,
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

	r := buildResult(resp, body)

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

	r := buildResult(resp, body)

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

	r := buildResult(resp, body)

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

	r := buildResult(resp, body)

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
