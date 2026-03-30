package intercom

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"strings"
	"testing"

	"github.com/rassakhatsky/intercom-go-sdk/api"
)

func TestNewClient_Defaults(t *testing.T) {
	c := NewClient("tok")

	if c.token != "tok" {
		t.Errorf("token = %q, want %q", c.token, "tok")
	}
	if c.baseURL != defaultBaseURL {
		t.Errorf("baseURL = %q, want %q", c.baseURL, defaultBaseURL)
	}
	if c.httpClient == nil {
		t.Error("httpClient should not be nil")
	}
	if c.httpClient == http.DefaultClient {
		t.Error("httpClient should not be http.DefaultClient")
	}
	if c.logger == nil {
		t.Error("logger should not be nil")
	}
}

func TestNewClient_WithHTTPClient(t *testing.T) {
	custom := &http.Client{}
	c := NewClient("tok", WithHTTPClient(custom))

	if c.httpClient != custom {
		t.Error("expected custom HTTP client")
	}
}

func TestNewClient_WithLogger(t *testing.T) {
	l := slog.New(slog.NewTextHandler(&bytes.Buffer{}, nil))
	c := NewClient("tok", WithLogger(l))

	if c.logger != l {
		t.Error("expected custom logger")
	}
}

func TestNewClient_WithBaseURL(t *testing.T) {
	c := NewClient("tok", WithBaseURL("https://custom.example.com/"))

	if c.baseURL != "https://custom.example.com/" {
		t.Errorf("baseURL = %q, want %q", c.baseURL, "https://custom.example.com/")
	}
}

func TestNewClient_WithBaseURL_TrailingSlash(t *testing.T) {
	c := NewClient("tok", WithBaseURL("https://custom.example.com"))

	if c.baseURL != "https://custom.example.com/" {
		t.Errorf("baseURL = %q, want trailing slash", c.baseURL)
	}
}

func TestNewRequest_Headers(t *testing.T) {
	c := NewClient("my-token")

	req, err := c.NewRequest(http.MethodGet, "contacts", nil)
	if err != nil {
		t.Fatalf("NewRequest returned error: %v", err)
	}

	testHeader(t, req, "Authorization", "Bearer my-token")
	testHeader(t, req, "Intercom-Version", "2.15")
	testHeader(t, req, "Accept", "application/json")

	// Content-Type should not be set for bodyless requests
	if ct := req.Header.Get("Content-Type"); ct != "" {
		t.Errorf("Content-Type = %q for bodyless request, want empty", ct)
	}

	ua := req.Header.Get("User-Agent")
	if ua == "" {
		t.Error("User-Agent header should not be empty")
	}
}

func TestNewRequest_Headers_WithBody(t *testing.T) {
	c := NewClient("my-token")

	req, err := c.NewRequest(http.MethodPost, "contacts", map[string]string{"name": "Alice"})
	if err != nil {
		t.Fatalf("NewRequest returned error: %v", err)
	}

	testHeader(t, req, "Content-Type", "application/json")
}

func TestNewRequest_Body(t *testing.T) {
	c := NewClient("tok")

	body := map[string]string{"name": "Alice"}
	req, err := c.NewRequest(http.MethodPost, "contacts", body)
	if err != nil {
		t.Fatalf("NewRequest returned error: %v", err)
	}

	b, _ := io.ReadAll(req.Body)
	var got map[string]string
	if err := json.Unmarshal(b, &got); err != nil {
		t.Fatalf("failed to unmarshal body: %v", err)
	}
	if got["name"] != "Alice" {
		t.Errorf("body name = %q, want %q", got["name"], "Alice")
	}
}

func TestNewRequest_NilBody(t *testing.T) {
	c := NewClient("tok")

	req, err := c.NewRequest(http.MethodGet, "contacts", nil)
	if err != nil {
		t.Fatalf("NewRequest returned error: %v", err)
	}
	if req.Body != nil && req.Body != http.NoBody {
		t.Error("expected nil body for GET request with nil input")
	}
}

func TestDo_DecodesJSON(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	type payload struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	mux.HandleFunc("/contacts/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{"id":"123","name":"Alice"}`))
	})

	req, _ := client.NewRequest(http.MethodGet, "contacts/123", nil)
	var got payload
	_, err := client.Do(context.Background(), req, &got)
	if err != nil {
		t.Fatalf("Do returned error: %v", err)
	}
	if got.ID != "123" || got.Name != "Alice" {
		t.Errorf("Do decoded = %+v, want {ID:123, Name:Alice}", got)
	}
}

func TestDo_NilV(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})

	req, _ := client.NewRequest(http.MethodDelete, "contacts/123", nil)
	_, err := client.Do(context.Background(), req, nil)
	if err != nil {
		t.Fatalf("Do returned error: %v", err)
	}
}

func TestDo_HTTPError(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/999", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		w.Write([]byte(`{"type":"error.list","errors":[{"code":"not_found","message":"not found"}]}`))
	})

	req, _ := client.NewRequest(http.MethodGet, "contacts/999", nil)
	_, err := client.Do(context.Background(), req, nil)
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	if !strings.Contains(err.Error(), "not_found") {
		t.Errorf("error = %v, want to contain 'not_found'", err)
	}
}

func TestDo_LogsRequest(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	var buf bytes.Buffer
	client.logger = slog.New(slog.NewTextHandler(&buf, &slog.HandlerOptions{Level: slog.LevelDebug}))

	mux.HandleFunc("/contacts", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.Write([]byte(`{}`))
	})

	req, _ := client.NewRequest(http.MethodGet, "contacts", nil)
	client.Do(context.Background(), req, nil)

	if buf.Len() == 0 {
		t.Error("expected logger to receive messages during Do")
	}
}

func TestAddQueryOptions(t *testing.T) {
	type opts struct {
		Page  int    `url:"page,omitempty"`
		Name  string `url:"name,omitempty"`
		Empty string `url:"empty,omitempty"`
	}

	got, err := api.AddQueryOptions("contacts", &opts{Page: 2, Name: "alice"})
	if err != nil {
		t.Fatalf("addQueryOptions error: %v", err)
	}

	// Order of query params may vary, so check both
	if !strings.Contains(got, "page=2") {
		t.Errorf("expected page=2 in %q", got)
	}
	if !strings.Contains(got, "name=alice") {
		t.Errorf("expected name=alice in %q", got)
	}
	if strings.Contains(got, "empty=") {
		t.Errorf("expected empty to be omitted in %q", got)
	}
}

func TestAddQueryOptions_NilOpts(t *testing.T) {
	got, err := api.AddQueryOptions("contacts", nil)
	if err != nil {
		t.Fatalf("addQueryOptions error: %v", err)
	}
	if got != "contacts" {
		t.Errorf("got = %q, want %q", got, "contacts")
	}
}

func TestAddQueryOptions_PointerFields(t *testing.T) {
	type opts struct {
		Active *bool   `url:"active,omitempty"`
		Name   *string `url:"name,omitempty"`
	}

	active := true
	name := "bob"
	got, err := api.AddQueryOptions("contacts", &opts{Active: &active, Name: &name})
	if err != nil {
		t.Fatalf("addQueryOptions error: %v", err)
	}
	if !strings.Contains(got, "active=true") {
		t.Errorf("expected active=true in %q", got)
	}
	if !strings.Contains(got, "name=bob") {
		t.Errorf("expected name=bob in %q", got)
	}

	// nil pointers should be omitted
	got, err = api.AddQueryOptions("contacts", &opts{})
	if err != nil {
		t.Fatalf("addQueryOptions error: %v", err)
	}
	if got != "contacts" {
		t.Errorf("expected no query params for nil pointers, got %q", got)
	}
}

func TestAddQueryOptions_SkipTag(t *testing.T) {
	type opts struct {
		Page   int    `url:"page,omitempty"`
		Secret string `url:"-"`
	}

	got, err := api.AddQueryOptions("contacts", &opts{Page: 1, Secret: "hidden"})
	if err != nil {
		t.Fatalf("addQueryOptions error: %v", err)
	}
	if !strings.Contains(got, "page=1") {
		t.Errorf("expected page=1 in %q", got)
	}
	if strings.Contains(got, "hidden") || strings.Contains(got, "Secret") {
		t.Errorf("expected Secret to be skipped, got %q", got)
	}
}

func TestAddQueryOptions_NonStruct(t *testing.T) {
	_, err := api.AddQueryOptions("contacts", "not-a-struct")
	if err == nil {
		t.Error("expected error for non-struct input")
	}
}

func TestClient_AllAccessorsNonNil(t *testing.T) {
	c := NewClient("tok")

	accessors := []struct {
		name string
		val  any
	}{
		{"AIContent", c.AIContent()},
		{"FinVoice", c.FinVoice()},
		{"Segments", c.Segments()},
		{"Tags", c.Tags()},
		{"ExportReporting", c.ExportReporting()},
		{"DataExport", c.DataExport()},
		{"News", c.News()},
		{"Messages", c.Messages()},
		{"Emails", c.Emails()},
		{"SubscriptionTypes", c.SubscriptionTypes()},
		{"Brands", c.Brands()},
		{"IPAllowlist", c.IPAllowlist()},
		{"CustomChannelEvents", c.CustomChannelEvents()},
		{"Jobs", c.Jobs()},
		{"Notes", c.Notes()},
		{"DataEvents", c.DataEvents()},
		{"DataAttributes", c.DataAttributes()},
		{"CustomObjects", c.CustomObjects()},
		{"Companies", c.Companies()},
		{"Admins", c.Admins()},
		{"Teams", c.Teams()},
		{"AwayStatusReasons", c.AwayStatusReasons()},
		{"Calls", c.Calls()},
		{"PhoneCallRedirects", c.PhoneCallRedirects()},
		{"HelpCenter", c.HelpCenter()},
		{"Articles", c.Articles()},
		{"InternalArticles", c.InternalArticles()},
		{"Contacts", c.Contacts()},
		{"Visitors", c.Visitors()},
		{"Conversations", c.Conversations()},
		{"Tickets", c.Tickets()},
		{"TicketTypes", c.TicketTypes()},
		{"TicketStates", c.TicketStates()},
		{"Workflows", c.Workflows()},
	}

	for _, a := range accessors {
		if a.val == nil {
			t.Errorf("%s() returned nil, want non-nil service", a.name)
		}
	}
}

func TestNewClient_EmptyToken_Panics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for empty token, got none")
		}
		msg, ok := r.(string)
		if !ok {
			t.Fatalf("expected string panic, got %T: %v", r, r)
		}
		if !strings.Contains(msg, "token") {
			t.Errorf("panic message = %q, want it to mention 'token'", msg)
		}
	}()
	NewClient("")
}

func TestNewClient_WhitespaceToken_Panics(t *testing.T) {
	defer func() {
		r := recover()
		if r == nil {
			t.Fatal("expected panic for whitespace-only token, got none")
		}
	}()
	NewClient("   ")
}

func TestNewClient_ValidToken_NoPanic(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("unexpected panic for valid token: %v", r)
		}
	}()
	c := NewClient("valid-token")
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

// Ensure *slog.Logger satisfies the Logger interface at compile time.
var _ Logger = slog.Default()
