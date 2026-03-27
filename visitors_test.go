package intercom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestVisitorsService_Get(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/visitors", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testHeader(t, r, "Authorization", "Bearer test-token")
		if got := r.URL.Query().Get("user_id"); got != "visitor-uid-1" {
			t.Errorf("user_id query param = %v, want visitor-uid-1", got)
		}
		fmt.Fprint(w, `{
			"type":"visitor",
			"id":"530370b477ad7120001d",
			"user_id":"visitor-uid-1",
			"anonymous":true,
			"email":"jane@example.com",
			"name":"Jane Doe",
			"pseudonym":"Red Duck from Dublin",
			"phone":"555-555-5555",
			"app_id":"hfi1bx4l",
			"created_at":1663597223,
			"updated_at":1663597260,
			"last_request_at":1663597260,
			"session_count":5,
			"unsubscribed_from_emails":false,
			"marked_email_as_spam":false,
			"has_hard_bounced":false,
			"custom_attributes":{"plan":"free"},
			"referrer":"https://www.google.com/",
			"utm_source":"Intercom",
			"utm_medium":"email",
			"utm_campaign":"intercom-link",
			"utm_content":"banner",
			"utm_term":"messenger",
			"avatar":{"type":"avatar","image_url":"https://example.com/avatar.png"},
			"tags":{"type":"tag.list","tags":[]},
			"segments":{"type":"segment.list","segments":[]},
			"social_profiles":{"type":"social_profile.list","social_profiles":[]},
			"companies":{"type":"company.list","companies":[]}
		}`)
	})

	ctx := context.Background()
	visitor, err := client.Visitors.Get(ctx, "visitor-uid-1")
	if err != nil {
		t.Fatalf("Visitors.Get returned error: %v", err)
	}
	if visitor.Type != "visitor" {
		t.Errorf("Type = %v, want visitor", visitor.Type)
	}
	if visitor.ID != "530370b477ad7120001d" {
		t.Errorf("ID = %v, want 530370b477ad7120001d", visitor.ID)
	}
	if visitor.UserID != "visitor-uid-1" {
		t.Errorf("UserID = %v, want visitor-uid-1", visitor.UserID)
	}
	if !visitor.Anonymous {
		t.Error("Anonymous = false, want true")
	}
	if visitor.Email != "jane@example.com" {
		t.Errorf("Email = %v, want jane@example.com", visitor.Email)
	}
	if visitor.Name != "Jane Doe" {
		t.Errorf("Name = %v, want Jane Doe", visitor.Name)
	}
	if visitor.Pseudonym != "Red Duck from Dublin" {
		t.Errorf("Pseudonym = %v, want Red Duck from Dublin", visitor.Pseudonym)
	}
	if visitor.Phone != "555-555-5555" {
		t.Errorf("Phone = %v, want 555-555-5555", visitor.Phone)
	}
	if visitor.SessionCount != 5 {
		t.Errorf("SessionCount = %d, want 5", visitor.SessionCount)
	}
	if visitor.Referrer != "https://www.google.com/" {
		t.Errorf("Referrer = %v, want https://www.google.com/", visitor.Referrer)
	}
	if visitor.UTMSource != "Intercom" {
		t.Errorf("UTMSource = %v, want Intercom", visitor.UTMSource)
	}
	if visitor.Avatar.ImageURL != "https://example.com/avatar.png" {
		t.Errorf("Avatar.ImageURL = %v, want https://example.com/avatar.png", visitor.Avatar.ImageURL)
	}
	if visitor.CustomAttributes["plan"] != "free" {
		t.Errorf("CustomAttributes[plan] = %v, want free", visitor.CustomAttributes["plan"])
	}
}

func TestVisitorsService_Get_NotFound(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/visitors", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{
			"type":"error.list",
			"request_id":"req-123",
			"errors":[{"code":"not_found","message":"Visitor Not Found"}]
		}`)
	})

	ctx := context.Background()
	_, err := client.Visitors.Get(ctx, "nonexistent")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("Expected IsNotFound, got: %v", err)
	}
}

func TestVisitorsService_Update(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/visitors", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		testHeader(t, r, "Authorization", "Bearer test-token")

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["id"] != "530370b477ad7120001d" {
			t.Errorf("body id = %v, want 530370b477ad7120001d", body["id"])
		}
		if body["name"] != "Updated Name" {
			t.Errorf("body name = %v, want Updated Name", body["name"])
		}
		attrs, ok := body["custom_attributes"].(map[string]any)
		if !ok || attrs["plan"] != "premium" {
			t.Errorf("body custom_attributes.plan = %v, want premium", attrs["plan"])
		}

		fmt.Fprint(w, `{
			"type":"visitor",
			"id":"530370b477ad7120001d",
			"user_id":"visitor-uid-1",
			"name":"Updated Name",
			"custom_attributes":{"plan":"premium"},
			"created_at":1663597223,
			"updated_at":1663597999
		}`)
	})

	ctx := context.Background()
	visitor, err := client.Visitors.Update(ctx, &UpdateVisitorRequest{
		ID:   "530370b477ad7120001d",
		Name: "Updated Name",
		CustomAttributes: map[string]any{
			"plan": "premium",
		},
	})
	if err != nil {
		t.Fatalf("Visitors.Update returned error: %v", err)
	}
	if visitor.Name != "Updated Name" {
		t.Errorf("Name = %v, want Updated Name", visitor.Name)
	}
	if visitor.UpdatedAt != 1663597999 {
		t.Errorf("UpdatedAt = %v, want 1663597999", visitor.UpdatedAt)
	}
}

func TestVisitorsService_Update_ByUserID(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/visitors", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["user_id"] != "visitor-uid-1" {
			t.Errorf("body user_id = %v, want visitor-uid-1", body["user_id"])
		}

		fmt.Fprint(w, `{
			"type":"visitor",
			"id":"530370b477ad7120001d",
			"user_id":"visitor-uid-1",
			"name":"Updated Via UserID"
		}`)
	})

	ctx := context.Background()
	visitor, err := client.Visitors.Update(ctx, &UpdateVisitorRequest{
		UserID: "visitor-uid-1",
		Name:   "Updated Via UserID",
	})
	if err != nil {
		t.Fatalf("Visitors.Update returned error: %v", err)
	}
	if visitor.Name != "Updated Via UserID" {
		t.Errorf("Name = %v, want Updated Via UserID", visitor.Name)
	}
}

func TestVisitorsService_Convert(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/visitors/convert", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		testHeader(t, r, "Authorization", "Bearer test-token")

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["type"] != "user" {
			t.Errorf("body type = %v, want user", body["type"])
		}
		visitor, ok := body["visitor"].(map[string]any)
		if !ok {
			t.Fatal("body visitor is missing or not an object")
		}
		if visitor["id"] != "530370b477ad7120001d" {
			t.Errorf("body visitor.id = %v, want 530370b477ad7120001d", visitor["id"])
		}
		user, ok := body["user"].(map[string]any)
		if !ok {
			t.Fatal("body user is missing or not an object")
		}
		if user["email"] != "jane@example.com" {
			t.Errorf("body user.email = %v, want jane@example.com", user["email"])
		}

		fmt.Fprint(w, `{
			"type":"contact",
			"id":"converted-contact-id",
			"role":"user",
			"email":"jane@example.com",
			"name":"Jane Doe"
		}`)
	})

	ctx := context.Background()
	contact, err := client.Visitors.Convert(ctx, &ConvertVisitorRequest{
		Type: "user",
		Visitor: ConvertVisitorIdentifier{
			ID: "530370b477ad7120001d",
		},
		User: ConvertVisitorUser{
			Email: "jane@example.com",
		},
	})
	if err != nil {
		t.Fatalf("Visitors.Convert returned error: %v", err)
	}
	if contact.Type != "contact" {
		t.Errorf("Type = %v, want contact", contact.Type)
	}
	if contact.ID != "converted-contact-id" {
		t.Errorf("ID = %v, want converted-contact-id", contact.ID)
	}
	if contact.Role != "user" {
		t.Errorf("Role = %v, want user", contact.Role)
	}
	if contact.Email != "jane@example.com" {
		t.Errorf("Email = %v, want jane@example.com", contact.Email)
	}
}

func TestVisitorsService_GetRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/visitors", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		if got := r.URL.Query().Get("user_id"); got != "visitor-uid-1" {
			t.Errorf("user_id query param = %v, want visitor-uid-1", got)
		}
		w.Header().Set("X-Request-Id", "req-visitor-get")
		fmt.Fprint(w, `{"type":"visitor","id":"530370b477ad7120001d","user_id":"visitor-uid-1","name":"Jane Doe"}`)
	})

	ctx := context.Background()
	result, err := client.Visitors.GetRaw(ctx, "visitor-uid-1")
	if err != nil {
		t.Fatalf("GetRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-visitor-get" {
		t.Errorf("Header X-Request-Id = %q, want req-visitor-get", result.Header.Get("X-Request-Id"))
	}
	if result.Error != nil {
		t.Errorf("Result.Error = %v, want nil", result.Error)
	}
	visitor, err := ParseVisitorGetResult(result)
	if err != nil {
		t.Fatalf("ParseVisitorGetResult returned error: %v", err)
	}
	if visitor.ID != "530370b477ad7120001d" {
		t.Errorf("ID = %v, want 530370b477ad7120001d", visitor.ID)
	}
	if visitor.Name != "Jane Doe" {
		t.Errorf("Name = %v, want Jane Doe", visitor.Name)
	}
}

func TestVisitorsService_GetRaw_NotFound(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/visitors", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"Visitor Not Found"}]}`)
	})

	ctx := context.Background()
	result, err := client.Visitors.GetRaw(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("GetRaw returned Go error: %v", err)
	}
	if result.Error == nil {
		t.Fatal("Result.Error is nil, want non-nil")
	}
	if result.Error.Code != "not_found" {
		t.Errorf("Error.Code = %q, want not_found", result.Error.Code)
	}
	if result.StatusCode != 404 {
		t.Errorf("StatusCode = %d, want 404", result.StatusCode)
	}
}

func TestVisitorsService_UpdateRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/visitors", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		w.Header().Set("X-Request-Id", "req-visitor-update")
		fmt.Fprint(w, `{"type":"visitor","id":"530370b477ad7120001d","name":"Updated Name"}`)
	})

	ctx := context.Background()
	result, err := client.Visitors.UpdateRaw(ctx, &UpdateVisitorRequest{
		ID:   "530370b477ad7120001d",
		Name: "Updated Name",
	})
	if err != nil {
		t.Fatalf("UpdateRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	visitor, err := ParseVisitorUpdateResult(result)
	if err != nil {
		t.Fatalf("ParseVisitorUpdateResult returned error: %v", err)
	}
	if visitor.Name != "Updated Name" {
		t.Errorf("Name = %v, want Updated Name", visitor.Name)
	}
}

func TestVisitorsService_ConvertRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/visitors/convert", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.Header().Set("X-Request-Id", "req-visitor-convert")
		fmt.Fprint(w, `{"type":"contact","id":"converted-contact-id","role":"user","email":"jane@example.com"}`)
	})

	ctx := context.Background()
	result, err := client.Visitors.ConvertRaw(ctx, &ConvertVisitorRequest{
		Type:    "user",
		Visitor: ConvertVisitorIdentifier{ID: "530370b477ad7120001d"},
		User:    ConvertVisitorUser{Email: "jane@example.com"},
	})
	if err != nil {
		t.Fatalf("ConvertRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-visitor-convert" {
		t.Errorf("Header X-Request-Id = %q, want req-visitor-convert", result.Header.Get("X-Request-Id"))
	}
	contact, err := ParseVisitorConvertResult(result)
	if err != nil {
		t.Fatalf("ParseVisitorConvertResult returned error: %v", err)
	}
	if contact.ID != "converted-contact-id" {
		t.Errorf("ID = %v, want converted-contact-id", contact.ID)
	}
	if contact.Role != "user" {
		t.Errorf("Role = %v, want user", contact.Role)
	}
}

func TestVisitorsService_Convert_ToLead(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/visitors/convert", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)

		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["type"] != "lead" {
			t.Errorf("body type = %v, want lead", body["type"])
		}

		fmt.Fprint(w, `{
			"type":"contact",
			"id":"converted-lead-id",
			"role":"lead",
			"email":"visitor@example.com"
		}`)
	})

	ctx := context.Background()
	contact, err := client.Visitors.Convert(ctx, &ConvertVisitorRequest{
		Type: "lead",
		Visitor: ConvertVisitorIdentifier{
			UserID: "visitor-uid-1",
		},
		User: ConvertVisitorUser{
			UserID: "external-user-id",
		},
	})
	if err != nil {
		t.Fatalf("Visitors.Convert returned error: %v", err)
	}
	if contact.Role != "lead" {
		t.Errorf("Role = %v, want lead", contact.Role)
	}
}

func TestVisitorsService_Convert_Unauthorized(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/visitors/convert", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{
			"type":"error.list",
			"request_id":"req-456",
			"errors":[{"code":"unauthorized","message":"Access Token Invalid"}]
		}`)
	})

	ctx := context.Background()
	_, err := client.Visitors.Convert(ctx, &ConvertVisitorRequest{
		Type:    "user",
		Visitor: ConvertVisitorIdentifier{ID: "some-id"},
		User:    ConvertVisitorUser{Email: "test@example.com"},
	})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !IsUnauthorized(err) {
		t.Errorf("Expected IsUnauthorized, got: %v", err)
	}
}
