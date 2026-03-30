package api

import (
	"strings"
	"testing"
)

func TestAddQueryOptions(t *testing.T) {
	type opts struct {
		Page int    `url:"page,omitempty"`
		Name string `url:"name,omitempty"`
	}

	got, err := AddQueryOptions("contacts", &opts{Page: 2, Name: "alice"})
	if err != nil {
		t.Fatalf("AddQueryOptions error: %v", err)
	}

	if !strings.Contains(got, "page=2") {
		t.Errorf("got %q, want page=2 in query", got)
	}
	if !strings.Contains(got, "name=alice") {
		t.Errorf("got %q, want name=alice in query", got)
	}
}

func TestAddQueryOptions_NilOpts(t *testing.T) {
	got, err := AddQueryOptions("contacts", nil)
	if err != nil {
		t.Fatalf("AddQueryOptions error: %v", err)
	}
	if got != "contacts" {
		t.Errorf("got %q, want %q", got, "contacts")
	}
}

func TestAddQueryOptions_PointerFields(t *testing.T) {
	type opts struct {
		Active *bool   `url:"active,omitempty"`
		Name   *string `url:"name,omitempty"`
	}

	active := true
	name := "alice"
	got, err := AddQueryOptions("contacts", &opts{Active: &active, Name: &name})
	if err != nil {
		t.Fatalf("AddQueryOptions error: %v", err)
	}
	if !strings.Contains(got, "active=true") {
		t.Errorf("got %q, want active=true in query", got)
	}
	if !strings.Contains(got, "name=alice") {
		t.Errorf("got %q, want name=alice in query", got)
	}

	// All nil — should return path unchanged
	got, err = AddQueryOptions("contacts", &opts{})
	if err != nil {
		t.Fatalf("AddQueryOptions error: %v", err)
	}
	if got != "contacts" {
		t.Errorf("got %q, want %q", got, "contacts")
	}
}

func TestAddQueryOptions_SkipTag(t *testing.T) {
	type opts struct {
		Page   int    `url:"page,omitempty"`
		Secret string `url:"-"`
	}

	got, err := AddQueryOptions("contacts", &opts{Page: 1, Secret: "hidden"})
	if err != nil {
		t.Fatalf("AddQueryOptions error: %v", err)
	}
	if strings.Contains(got, "hidden") {
		t.Errorf("got %q, should not contain untagged field value", got)
	}
	if !strings.Contains(got, "page=1") {
		t.Errorf("got %q, want page=1 in query", got)
	}
}

func TestAddQueryOptions_NonStruct(t *testing.T) {
	_, err := AddQueryOptions("contacts", "not-a-struct")
	if err == nil {
		t.Error("expected error for non-struct opts, got nil")
	}
}
