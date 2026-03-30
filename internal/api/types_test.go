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

func TestPart_JSONRoundTrip(t *testing.T) {
	p := Part{
		Type:       "conversation_part",
		ID:         "part_1",
		PartType:   "comment",
		Body:       "Hello",
		CreatedAt:  1700000000,
		UpdatedAt:  1700001000,
		NotifiedAt: 1700002000,
		AssignedTo: &Author{Type: "admin", ID: "a1"},
		Author:     &Author{Type: "admin", ID: "a2", Name: "Bob"},
		ExternalID: "ext_1",
		Redacted:   true,
	}
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got Part
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Type != p.Type || got.ID != p.ID || got.PartType != p.PartType ||
		got.Body != p.Body || got.CreatedAt != p.CreatedAt || got.UpdatedAt != p.UpdatedAt ||
		got.NotifiedAt != p.NotifiedAt || got.ExternalID != p.ExternalID || got.Redacted != p.Redacted {
		t.Fatalf("got %+v, want %+v", got, p)
	}
	if got.AssignedTo == nil || got.AssignedTo.ID != "a1" {
		t.Fatalf("AssignedTo mismatch: %+v", got.AssignedTo)
	}
	if got.Author == nil || got.Author.Name != "Bob" {
		t.Fatalf("Author mismatch: %+v", got.Author)
	}
}

func TestPart_WithoutNotifiedAt(t *testing.T) {
	// Tickets don't have NotifiedAt — verify it stays zero and is omitted
	raw := `{"type":"ticket_part","id":"tp_1","part_type":"note","body":"Hi"}`
	var p Part
	if err := json.Unmarshal([]byte(raw), &p); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if p.Type != "ticket_part" || p.ID != "tp_1" || p.NotifiedAt != 0 {
		t.Fatalf("unexpected: %+v", p)
	}
	// Re-marshal: notified_at should be absent
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	if contains(string(data), "notified_at") {
		t.Fatalf("expected notified_at omitted, got %s", string(data))
	}
}

func TestPart_OmitEmpty(t *testing.T) {
	p := Part{Type: "conversation_part", ID: "p1"}
	data, err := json.Marshal(p)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(data)
	for _, field := range []string{"part_type", "body", "created_at", "updated_at", "notified_at", "assigned_to", "author", "external_id", "redacted"} {
		if contains(s, field) {
			t.Fatalf("expected %s omitted, got %s", field, s)
		}
	}
}

func TestContactRef_JSONRoundTrip(t *testing.T) {
	c := ContactRef{
		Type:       "contact",
		ID:         "c1",
		ExternalID: "ext_1",
	}
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got ContactRef
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got != c {
		t.Fatalf("got %+v, want %+v", got, c)
	}
}

func TestContactRef_BackwardCompat(t *testing.T) {
	// Existing usage without ExternalID must still work
	c := ContactRef{Type: "contact", ID: "c2"}
	data, err := json.Marshal(c)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(data)
	if contains(s, "external_id") {
		t.Fatalf("expected external_id omitted, got %s", s)
	}
	// Unmarshal without external_id
	raw := `{"type":"contact","id":"c2"}`
	var got ContactRef
	if err := json.Unmarshal([]byte(raw), &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.ExternalID != "" {
		t.Fatalf("expected empty ExternalID, got %q", got.ExternalID)
	}
}

func TestContactRefList_JSONRoundTrip(t *testing.T) {
	l := ContactRefList{
		Type: "contact_list",
		Contacts: []ContactRef{
			{Type: "contact", ID: "c1", ExternalID: "ext_1"},
			{Type: "contact", ID: "c2"},
		},
	}
	data, err := json.Marshal(l)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got ContactRefList
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Type != l.Type || len(got.Contacts) != 2 {
		t.Fatalf("got %+v, want %+v", got, l)
	}
	if got.Contacts[0].ExternalID != "ext_1" {
		t.Fatalf("expected ext_1, got %q", got.Contacts[0].ExternalID)
	}
	if got.Contacts[1].ExternalID != "" {
		t.Fatalf("expected empty ExternalID, got %q", got.Contacts[1].ExternalID)
	}
}

func TestContactRefList_Empty(t *testing.T) {
	raw := `{"type":"contact_list","contacts":[]}`
	var l ContactRefList
	if err := json.Unmarshal([]byte(raw), &l); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if l.Type != "contact_list" || len(l.Contacts) != 0 {
		t.Fatalf("unexpected: %+v", l)
	}
}

func TestDeleted_JSONRoundTrip(t *testing.T) {
	d := Deleted{
		ID:      "123",
		Object:  "ticket",
		Deleted: true,
	}
	data, err := json.Marshal(d)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got Deleted
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got != d {
		t.Fatalf("got %+v, want %+v", got, d)
	}
}

func TestDeleted_UnmarshalVariousObjects(t *testing.T) {
	for _, obj := range []string{"ticket", "article", "conversation", "company"} {
		raw := `{"id":"456","object":"` + obj + `","deleted":true}`
		var d Deleted
		if err := json.Unmarshal([]byte(raw), &d); err != nil {
			t.Fatalf("unmarshal %s: %v", obj, err)
		}
		if d.ID != "456" || d.Object != obj || !d.Deleted {
			t.Fatalf("unexpected for %s: %+v", obj, d)
		}
	}
}

func TestDeleted_DeletedFalse(t *testing.T) {
	raw := `{"id":"789","object":"ticket","deleted":false}`
	var d Deleted
	if err := json.Unmarshal([]byte(raw), &d); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if d.Deleted {
		t.Fatal("expected Deleted=false")
	}
}

func TestTagRefList_JSONRoundTrip(t *testing.T) {
	l := TagRefList{
		Type: "tag.list",
		Tags: []TagRef{
			{Type: "tag", ID: "t1", Name: "VIP"},
			{Type: "tag", ID: "t2", Name: "Support"},
		},
	}
	data, err := json.Marshal(l)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	var got TagRefList
	if err := json.Unmarshal(data, &got); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if got.Type != l.Type || len(got.Tags) != 2 {
		t.Fatalf("got %+v, want %+v", got, l)
	}
	if got.Tags[0].ID != "t1" || got.Tags[0].Name != "VIP" {
		t.Fatalf("tag 0 mismatch: %+v", got.Tags[0])
	}
	if got.Tags[1].ID != "t2" || got.Tags[1].Name != "Support" {
		t.Fatalf("tag 1 mismatch: %+v", got.Tags[1])
	}
}

func TestTagRefList_Empty(t *testing.T) {
	raw := `{"type":"tag.list","tags":[]}`
	var l TagRefList
	if err := json.Unmarshal([]byte(raw), &l); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if l.Type != "tag.list" || len(l.Tags) != 0 {
		t.Fatalf("unexpected: %+v", l)
	}
}

func TestTagRefList_JSONKey(t *testing.T) {
	// Verify the JSON key is "tags" not "data" (distinct from TagList)
	l := TagRefList{
		Type: "tag.list",
		Tags: []TagRef{{Type: "tag", ID: "t1"}},
	}
	data, err := json.Marshal(l)
	if err != nil {
		t.Fatalf("marshal: %v", err)
	}
	s := string(data)
	if !contains(s, `"tags"`) {
		t.Fatalf("expected 'tags' key, got %s", s)
	}
	if contains(s, `"data"`) {
		t.Fatalf("expected no 'data' key, got %s", s)
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
