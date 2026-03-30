package intercom

import (
	"net/http"
	"testing"

	"github.com/rassakhatsky/intercom-go-sdk/api"
)

// TestPublicAPITypes verifies that key types are accessible from the public api/ package.
func TestPublicAPITypes(t *testing.T) {
	// Result type
	var _ api.Result

	// Filter type
	var _ api.Filter

	// ErrorResponse type
	var _ api.ErrorResponse

	// Iter generic type
	var _ *api.Iter[any]

	// PagedResult generic type
	var _ api.PagedResult[any]
}

// TestPublicAPIFunctions verifies that key functions are accessible from the public api/ package.
func TestPublicAPIFunctions(t *testing.T) {
	// BuildResult function
	resp := &http.Response{
		StatusCode: 200,
		Header:     http.Header{},
	}
	r := api.BuildResult(resp, []byte(`{}`))
	if r == nil {
		t.Fatal("BuildResult returned nil")
	}

	// ResultError function — nil input should return nil
	errResp := api.ResultError(nil)
	if errResp != nil {
		t.Fatal("ResultError(nil) should return nil")
	}

	// SingleFilterOf function
	f := api.SingleFilterOf("email", api.OpEquals, "test@example.com")
	if f.Field != "email" {
		t.Fatalf("expected field 'email', got %q", f.Field)
	}
}
