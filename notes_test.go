package intercom

import (
	"context"
	"fmt"
	"net/http"
	"testing"
)

func TestNotesService_Get(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/notes/34", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testHeader(t, r, "Authorization", "Bearer test-token")
		fmt.Fprint(w, `{
			"type": "note",
			"id": "34",
			"created_at": 1733846617,
			"contact": {
				"type": "contact",
				"id": "6762f2591bb69f9f2193bc1f"
			},
			"author": {
				"type": "admin",
				"id": "991267864",
				"name": "Ciaran Lee",
				"email": "admin@email.com"
			},
			"body": "<p>This is a note.</p>"
		}`)
	})

	ctx := context.Background()
	note, err := client.Notes.Get(ctx, "34")
	if err != nil {
		t.Fatalf("Notes.Get returned error: %v", err)
	}
	if note.Type != "note" {
		t.Errorf("Note.Type = %v, want note", note.Type)
	}
	if note.ID != "34" {
		t.Errorf("Note.ID = %v, want 34", note.ID)
	}
	if note.CreatedAt != 1733846617 {
		t.Errorf("Note.CreatedAt = %v, want 1733846617", note.CreatedAt)
	}
	if note.Body != "<p>This is a note.</p>" {
		t.Errorf("Note.Body = %v, want <p>This is a note.</p>", note.Body)
	}
	if note.Contact == nil {
		t.Fatal("Note.Contact is nil")
	}
	if note.Contact.Type != "contact" {
		t.Errorf("Note.Contact.Type = %v, want contact", note.Contact.Type)
	}
	if note.Contact.ID != "6762f2591bb69f9f2193bc1f" {
		t.Errorf("Note.Contact.ID = %v, want 6762f2591bb69f9f2193bc1f", note.Contact.ID)
	}
	if note.Author == nil {
		t.Fatal("Note.Author is nil")
	}
	if note.Author.Type != "admin" {
		t.Errorf("Note.Author.Type = %v, want admin", note.Author.Type)
	}
	if note.Author.ID != "991267864" {
		t.Errorf("Note.Author.ID = %v, want 991267864", note.Author.ID)
	}
	if note.Author.Name != "Ciaran Lee" {
		t.Errorf("Note.Author.Name = %v, want Ciaran Lee", note.Author.Name)
	}
	if note.Author.Email != "admin@email.com" {
		t.Errorf("Note.Author.Email = %v, want admin@email.com", note.Author.Email)
	}
}

func TestNotesService_Get_NotFound(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/notes/nonexistent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{
			"type": "error.list",
			"request_id": "req-123",
			"errors": [{"code": "not_found", "message": "Note not found"}]
		}`)
	})

	ctx := context.Background()
	_, err := client.Notes.Get(ctx, "nonexistent")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("Expected IsNotFound, got: %v", err)
	}
}

func TestNotesService_GetRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/notes/34", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-note")
		fmt.Fprint(w, `{"type":"note","id":"34","created_at":1733846617,"body":"<p>This is a note.</p>","contact":{"type":"contact","id":"abc123"},"author":{"type":"admin","id":"991267864","name":"Ciaran Lee","email":"admin@email.com"}}`)
	})

	ctx := context.Background()
	result, err := client.Notes.GetRaw(ctx, "34")
	if err != nil {
		t.Fatalf("GetRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-note" {
		t.Errorf("Header X-Request-Id = %q, want req-note", result.Header.Get("X-Request-Id"))
	}
	if result.Error != nil {
		t.Errorf("Result.Error = %v, want nil", result.Error)
	}
	note, err := ParseNoteGetResult(result)
	if err != nil {
		t.Fatalf("ParseNoteGetResult returned error: %v", err)
	}
	if note.ID != "34" {
		t.Errorf("Data.ID = %v, want 34", note.ID)
	}
	if note.Body != "<p>This is a note.</p>" {
		t.Errorf("Data.Body = %v, want <p>This is a note.</p>", note.Body)
	}
}
