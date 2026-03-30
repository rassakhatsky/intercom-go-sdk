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

func TestLinkedObjectList_JSONRoundTrip(t *testing.T) {
	l := LinkedObjectList{
		Type:       "list",
		Data:       []any{"ticket_1", float64(42)},
		TotalCount: 2,
		HasMore:    true,
	}
	data, err := json.Marshal(l)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got LinkedObjectList
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Type != l.Type || got.TotalCount != l.TotalCount || got.HasMore != l.HasMore {
		t.Fatalf("got %+v, want %+v", got, l)
	}
	if len(got.Data) != 2 {
		t.Fatalf("expected 2 data items, got %d", len(got.Data))
	}
}

func TestLinkedObjectList_EmptyData(t *testing.T) {
	raw := `{"type":"list","data":[],"total_count":0,"has_more":false}`
	var l LinkedObjectList
	if err := json.Unmarshal([]byte(raw), &l); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if l.Type != "list" || l.TotalCount != 0 || l.HasMore != false || len(l.Data) != 0 {
		t.Fatalf("unexpected: %+v", l)
	}
}

func TestLinkedObjectList_OmitEmpty(t *testing.T) {
	// Zero-value LinkedObjectList should still marshal required fields
	l := LinkedObjectList{Type: "list"}
	data, err := json.Marshal(l)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(data)
	// data should be null (nil slice), total_count and has_more should be present as zero values
	if !contains(s, `"type":"list"`) {
		t.Fatalf("expected type field, got %s", s)
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
