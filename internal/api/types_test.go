package api

import (
	"encoding/json"
	"testing"
)

func TestAuthor_JSONRoundTrip(t *testing.T) {
	a := Author{
		Type:  "admin",
		ID:    "123",
		Name:  "Alice",
		Email: "alice@example.com",
	}
	data, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got Author
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got != a {
		t.Fatalf("got %+v, want %+v", got, a)
	}
}

func TestAuthor_OmitEmpty(t *testing.T) {
	a := Author{Type: "bot", ID: "456"}
	data, err := json.Marshal(a)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(data)
	if contains(s, "name") || contains(s, "email") {
		t.Fatalf("expected name/email omitted, got %s", s)
	}
}

func TestAuthor_UnmarshalMinimal(t *testing.T) {
	raw := `{"type":"admin","id":"1"}`
	var a Author
	if err := json.Unmarshal([]byte(raw), &a); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if a.Type != "admin" || a.ID != "1" {
		t.Fatalf("unexpected: %+v", a)
	}
	if a.Name != "" || a.Email != "" {
		t.Fatalf("expected zero Name/Email, got %+v", a)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && searchString(s, sub)
}

func searchString(s, sub string) bool {
	for i := 0; i <= len(s)-len(sub); i++ {
		if s[i:i+len(sub)] == sub {
			return true
		}
	}
	return false
}
