package intercom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestContactsService_Get(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testHeader(t, r, "Authorization", "Bearer test-token")
		fmt.Fprint(w, `{"type":"contact","id":"123","role":"user","email":"joe@example.com","name":"Joe"}`)
	})

	ctx := context.Background()
	contact, err := client.Contacts.Get(ctx, "123")
	if err != nil {
		t.Fatalf("Contacts.Get returned error: %v", err)
	}
	if contact.ID != "123" {
		t.Errorf("Contact.ID = %v, want 123", contact.ID)
	}
	if contact.Email != "joe@example.com" {
		t.Errorf("Contact.Email = %v, want joe@example.com", contact.Email)
	}
	if contact.Name != "Joe" {
		t.Errorf("Contact.Name = %v, want Joe", contact.Name)
	}
}

func TestContactsService_Create(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body CreateContactRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.Email != "joe@example.com" {
			t.Errorf("Create body email = %v, want joe@example.com", body.Email)
		}
		if body.Role != "user" {
			t.Errorf("Create body role = %v, want user", body.Role)
		}
		fmt.Fprint(w, `{"type":"contact","id":"456","role":"user","email":"joe@example.com"}`)
	})

	ctx := context.Background()
	contact, err := client.Contacts.Create(ctx, &CreateContactRequest{
		Role:  "user",
		Email: "joe@example.com",
	})
	if err != nil {
		t.Fatalf("Contacts.Create returned error: %v", err)
	}
	if contact.ID != "456" {
		t.Errorf("Contact.ID = %v, want 456", contact.ID)
	}
}

func TestContactsService_Update(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		var body UpdateContactRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.Name != "Joe Updated" {
			t.Errorf("Update body name = %v, want Joe Updated", body.Name)
		}
		fmt.Fprint(w, `{"type":"contact","id":"123","name":"Joe Updated","email":"joe@example.com"}`)
	})

	ctx := context.Background()
	contact, err := client.Contacts.Update(ctx, "123", &UpdateContactRequest{
		Name: "Joe Updated",
	})
	if err != nil {
		t.Fatalf("Contacts.Update returned error: %v", err)
	}
	if contact.Name != "Joe Updated" {
		t.Errorf("Contact.Name = %v, want Joe Updated", contact.Name)
	}
}

func TestContactsService_Delete(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{"type":"contact","id":"123","external_id":"ext-1","deleted":true}`)
	})

	ctx := context.Background()
	deleted, err := client.Contacts.Delete(ctx, "123")
	if err != nil {
		t.Fatalf("Contacts.Delete returned error: %v", err)
	}
	if !deleted.Deleted {
		t.Error("ContactDeleted.Deleted = false, want true")
	}
	if deleted.ID != "123" {
		t.Errorf("ContactDeleted.ID = %v, want 123", deleted.ID)
	}
}

func TestContactsService_List(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"list",
			"data":[{"type":"contact","id":"1","email":"a@b.com"},{"type":"contact","id":"2","email":"c@d.com"}],
			"total_count":2,
			"pages":{"type":"pages","page":1,"per_page":50,"total_pages":1}
		}`)
	})

	ctx := context.Background()
	result, err := client.Contacts.List(ctx, nil)
	if err != nil {
		t.Fatalf("Contacts.List returned error: %v", err)
	}
	if len(result.Data) != 2 {
		t.Errorf("List returned %d contacts, want 2", len(result.Data))
	}
	if result.TotalCount != 2 {
		t.Errorf("TotalCount = %d, want 2", result.TotalCount)
	}
}

func TestContactsService_ListAll(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	callCount := 0
	mux.HandleFunc("/contacts", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		callCount++
		if callCount == 1 {
			fmt.Fprint(w, `{
				"type":"list",
				"data":[{"type":"contact","id":"1"}],
				"total_count":2,
				"pages":{"type":"pages","page":1,"per_page":1,"total_pages":2,"next":{"per_page":1,"starting_after":"cursor1"}}
			}`)
		} else {
			fmt.Fprint(w, `{
				"type":"list",
				"data":[{"type":"contact","id":"2"}],
				"total_count":2,
				"pages":{"type":"pages","page":2,"per_page":1,"total_pages":2}
			}`)
		}
	})

	ctx := context.Background()
	iter := client.Contacts.ListAll(ctx, &ListOptions{PerPage: 1})
	var ids []string
	for iter.Next() {
		ids = append(ids, iter.Current().ID)
	}
	if err := iter.Err(); err != nil {
		t.Fatalf("ListAll returned error: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("ListAll returned %d items, want 2", len(ids))
	}
	if ids[0] != "1" || ids[1] != "2" {
		t.Errorf("ListAll ids = %v, want [1 2]", ids)
	}
}

func TestContactsService_Search(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/search", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body SearchRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.Query.Operator != "AND" {
			t.Errorf("Search query operator = %v, want AND", body.Query.Operator)
		}
		fmt.Fprint(w, `{
			"type":"list",
			"data":[{"type":"contact","id":"1","email":"match@example.com"}],
			"total_count":1,
			"pages":{"type":"pages","page":1,"per_page":5,"total_pages":1}
		}`)
	})

	ctx := context.Background()
	result, err := client.Contacts.Search(ctx, &SearchRequest{
		Query: And(
			SingleFilterOf("created_at", ">", "1306054154"),
		),
		Pagination: &SearchPagination{PerPage: 5},
	})
	if err != nil {
		t.Fatalf("Contacts.Search returned error: %v", err)
	}
	if len(result.Data) != 1 {
		t.Errorf("Search returned %d contacts, want 1", len(result.Data))
	}
}

func TestContactsService_Merge(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/merge", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body MergeContactsRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.From != "lead-1" || body.Into != "user-1" {
			t.Errorf("Merge body = %+v, want from=lead-1 into=user-1", body)
		}
		fmt.Fprint(w, `{"type":"contact","id":"user-1","role":"user"}`)
	})

	ctx := context.Background()
	contact, err := client.Contacts.Merge(ctx, &MergeContactsRequest{
		From: "lead-1",
		Into: "user-1",
	})
	if err != nil {
		t.Fatalf("Contacts.Merge returned error: %v", err)
	}
	if contact.ID != "user-1" {
		t.Errorf("Contact.ID = %v, want user-1", contact.ID)
	}
}

func TestContactsService_Archive(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/archive", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"contact","id":"123","archived":true}`)
	})

	ctx := context.Background()
	result, err := client.Contacts.Archive(ctx, "123")
	if err != nil {
		t.Fatalf("Contacts.Archive returned error: %v", err)
	}
	if !result.Archived {
		t.Error("ContactArchived.Archived = false, want true")
	}
}

func TestContactsService_Unarchive(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/unarchive", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"contact","id":"123","archived":false}`)
	})

	ctx := context.Background()
	result, err := client.Contacts.Unarchive(ctx, "123")
	if err != nil {
		t.Fatalf("Contacts.Unarchive returned error: %v", err)
	}
	if result.Archived {
		t.Error("ContactUnarchived.Archived = true, want false")
	}
}

func TestContactsService_Block(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/block", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"contact","id":"123","blocked":true}`)
	})

	ctx := context.Background()
	result, err := client.Contacts.Block(ctx, "123")
	if err != nil {
		t.Fatalf("Contacts.Block returned error: %v", err)
	}
	if !result.Blocked {
		t.Error("ContactBlocked.Blocked = false, want true")
	}
}

func TestContactsService_FindByExternalID(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/find_by_external_id/ext-42", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"type":"contact","id":"abc","external_id":"ext-42","email":"joe@example.com"}`)
	})

	ctx := context.Background()
	contact, err := client.Contacts.FindByExternalID(ctx, "ext-42")
	if err != nil {
		t.Fatalf("Contacts.FindByExternalID returned error: %v", err)
	}
	if contact.ExternalID != "ext-42" {
		t.Errorf("Contact.ExternalID = %v, want ext-42", contact.ExternalID)
	}
}

func TestContactsService_ListCompanies(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/companies", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"list",
			"data":[{"type":"company","id":"c1","name":"Acme"}],
			"total_count":1,
			"pages":{"type":"pages","page":1,"per_page":50,"total_pages":1}
		}`)
	})

	ctx := context.Background()
	result, err := client.Contacts.ListCompanies(ctx, "123")
	if err != nil {
		t.Fatalf("Contacts.ListCompanies returned error: %v", err)
	}
	if len(result.Data) != 1 {
		t.Errorf("ListCompanies returned %d companies, want 1", len(result.Data))
	}
	if result.Data[0].Name != "Acme" {
		t.Errorf("Company.Name = %v, want Acme", result.Data[0].Name)
	}
}

func TestContactsService_AddCompany(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/companies", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body struct {
			ID string `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.ID != "c1" {
			t.Errorf("AddCompany body id = %v, want c1", body.ID)
		}
		fmt.Fprint(w, `{"type":"company","id":"c1","name":"Acme"}`)
	})

	ctx := context.Background()
	company, err := client.Contacts.AddCompany(ctx, "123", "c1")
	if err != nil {
		t.Fatalf("Contacts.AddCompany returned error: %v", err)
	}
	if company.ID != "c1" {
		t.Errorf("Company.ID = %v, want c1", company.ID)
	}
}

func TestContactsService_RemoveCompany(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/companies/c1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{"type":"company","id":"c1","name":"Acme"}`)
	})

	ctx := context.Background()
	company, err := client.Contacts.RemoveCompany(ctx, "123", "c1")
	if err != nil {
		t.Fatalf("Contacts.RemoveCompany returned error: %v", err)
	}
	if company.ID != "c1" {
		t.Errorf("Company.ID = %v, want c1", company.ID)
	}
}

func TestContactsService_ListNotes(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/notes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"list",
			"data":[{"type":"note","id":"n1","body":"<p>Hello</p>"}],
			"total_count":1,
			"pages":{"type":"pages","page":1,"per_page":50,"total_pages":1}
		}`)
	})

	ctx := context.Background()
	result, err := client.Contacts.ListNotes(ctx, "123")
	if err != nil {
		t.Fatalf("Contacts.ListNotes returned error: %v", err)
	}
	if len(result.Data) != 1 {
		t.Errorf("ListNotes returned %d notes, want 1", len(result.Data))
	}
}

func TestContactsService_CreateNote(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/notes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body CreateNoteRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.Body != "Hello" {
			t.Errorf("CreateNote body = %v, want Hello", body.Body)
		}
		fmt.Fprint(w, `{"type":"note","id":"n1","body":"<p>Hello</p>"}`)
	})

	ctx := context.Background()
	note, err := client.Contacts.CreateNote(ctx, "123", &CreateNoteRequest{
		Body:    "Hello",
		AdminID: "admin-1",
	})
	if err != nil {
		t.Fatalf("Contacts.CreateNote returned error: %v", err)
	}
	if note.ID != "n1" {
		t.Errorf("Note.ID = %v, want n1", note.ID)
	}
}

func TestContactsService_ListSegments(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/segments", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"list",
			"data":[{"type":"segment","id":"s1","name":"Active Users"}]
		}`)
	})

	ctx := context.Background()
	result, err := client.Contacts.ListSegments(ctx, "123")
	if err != nil {
		t.Fatalf("Contacts.ListSegments returned error: %v", err)
	}
	if len(result.Data) != 1 {
		t.Errorf("ListSegments returned %d segments, want 1", len(result.Data))
	}
	if result.Data[0].Name != "Active Users" {
		t.Errorf("Segment.Name = %v, want Active Users", result.Data[0].Name)
	}
}

func TestContactsService_ListSubscriptions(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/subscriptions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"list",
			"data":[{"type":"subscription","id":"sub-1","state":"live","consent_type":"opt_in"}]
		}`)
	})

	ctx := context.Background()
	result, err := client.Contacts.ListSubscriptions(ctx, "123")
	if err != nil {
		t.Fatalf("Contacts.ListSubscriptions returned error: %v", err)
	}
	if len(result.Data) != 1 {
		t.Errorf("ListSubscriptions returned %d subscriptions, want 1", len(result.Data))
	}
}

func TestContactsService_AddSubscription(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/subscriptions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body AddSubscriptionRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.ID != "sub-1" || body.ConsentType != "opt_in" {
			t.Errorf("AddSubscription body = %+v", body)
		}
		fmt.Fprint(w, `{"type":"subscription","id":"sub-1","state":"live","consent_type":"opt_in"}`)
	})

	ctx := context.Background()
	sub, err := client.Contacts.AddSubscription(ctx, "123", &AddSubscriptionRequest{
		ID:          "sub-1",
		ConsentType: "opt_in",
	})
	if err != nil {
		t.Fatalf("Contacts.AddSubscription returned error: %v", err)
	}
	if sub.ID != "sub-1" {
		t.Errorf("Subscription.ID = %v, want sub-1", sub.ID)
	}
}

func TestContactsService_RemoveSubscription(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/subscriptions/sub-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{"type":"subscription","id":"sub-1","state":"live","consent_type":"opt_in"}`)
	})

	ctx := context.Background()
	sub, err := client.Contacts.RemoveSubscription(ctx, "123", "sub-1")
	if err != nil {
		t.Fatalf("Contacts.RemoveSubscription returned error: %v", err)
	}
	if sub.ID != "sub-1" {
		t.Errorf("Subscription.ID = %v, want sub-1", sub.ID)
	}
}

func TestContactsService_AddTag(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/tags", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body struct {
			ID string `json:"id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.ID != "tag-1" {
			t.Errorf("AddTag body id = %v, want tag-1", body.ID)
		}
		fmt.Fprint(w, `{"type":"tag","id":"tag-1","name":"VIP"}`)
	})

	ctx := context.Background()
	tag, err := client.Contacts.AddTag(ctx, "123", "tag-1")
	if err != nil {
		t.Fatalf("Contacts.AddTag returned error: %v", err)
	}
	if tag.ID != "tag-1" {
		t.Errorf("Tag.ID = %v, want tag-1", tag.ID)
	}
	if tag.Name != "VIP" {
		t.Errorf("Tag.Name = %v, want VIP", tag.Name)
	}
}

func TestContactsService_RemoveTag(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/tags/tag-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{"type":"tag","id":"tag-1","name":"VIP"}`)
	})

	ctx := context.Background()
	tag, err := client.Contacts.RemoveTag(ctx, "123", "tag-1")
	if err != nil {
		t.Fatalf("Contacts.RemoveTag returned error: %v", err)
	}
	if tag.ID != "tag-1" {
		t.Errorf("Tag.ID = %v, want tag-1", tag.ID)
	}
}

// --- Raw method tests ---

func TestContactsService_GetRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-abc")
		fmt.Fprint(w, `{"type":"contact","id":"123","email":"joe@example.com"}`)
	})

	ctx := context.Background()
	result, err := client.Contacts.GetRaw(ctx, "123")
	if err != nil {
		t.Fatalf("GetRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-abc" {
		t.Errorf("Header X-Request-Id = %q, want req-abc", result.Header.Get("X-Request-Id"))
	}
	if result.Error != nil {
		t.Errorf("Result.Error = %v, want nil", result.Error)
	}
	contact, err := ParseContactGetResult(result)
	if err != nil {
		t.Fatalf("ParseContactGetResult returned error: %v", err)
	}
	if contact.ID != "123" {
		t.Errorf("Contact.ID = %v, want 123", contact.ID)
	}
	if contact.Email != "joe@example.com" {
		t.Errorf("Contact.Email = %v, want joe@example.com", contact.Email)
	}
}

func TestContactsService_GetRaw_NotFound(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/999", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"User Not Found"}]}`)
	})

	ctx := context.Background()
	result, err := client.Contacts.GetRaw(ctx, "999")
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

func TestContactsService_ListRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"list",
			"data":[{"type":"contact","id":"1"},{"type":"contact","id":"2"}],
			"total_count":2,
			"pages":{"type":"pages","page":1,"per_page":50,"total_pages":1}
		}`)
	})

	ctx := context.Background()
	result, err := client.Contacts.ListRaw(ctx, nil)
	if err != nil {
		t.Fatalf("ListRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	page, err := ParseContactListResult(result)
	if err != nil {
		t.Fatalf("ParseContactListResult returned error: %v", err)
	}
	if len(page.Data) != 2 {
		t.Errorf("Data len = %d, want 2", len(page.Data))
	}
}

func TestContactsService_CreateRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"type":"contact","id":"456","role":"user","email":"joe@example.com"}`)
	})

	ctx := context.Background()
	result, err := client.Contacts.CreateRaw(ctx, &CreateContactRequest{
		Role:  "user",
		Email: "joe@example.com",
	})
	if err != nil {
		t.Fatalf("CreateRaw returned error: %v", err)
	}
	contact, err := ParseContactCreateResult(result)
	if err != nil {
		t.Fatalf("ParseContactCreateResult returned error: %v", err)
	}
	if contact.ID != "456" {
		t.Errorf("Contact.ID = %v, want 456", contact.ID)
	}
}

func TestContactsService_UpdateRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `{"type":"contact","id":"123","name":"Joe Updated"}`)
	})

	ctx := context.Background()
	result, err := client.Contacts.UpdateRaw(ctx, "123", &UpdateContactRequest{Name: "Joe Updated"})
	if err != nil {
		t.Fatalf("UpdateRaw returned error: %v", err)
	}
	contact, err := ParseContactUpdateResult(result)
	if err != nil {
		t.Fatalf("ParseContactUpdateResult returned error: %v", err)
	}
	if contact.Name != "Joe Updated" {
		t.Errorf("Contact.Name = %v, want Joe Updated", contact.Name)
	}
}

func TestContactsService_DeleteRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{"type":"contact","id":"123","deleted":true}`)
	})

	ctx := context.Background()
	result, err := client.Contacts.DeleteRaw(ctx, "123")
	if err != nil {
		t.Fatalf("DeleteRaw returned error: %v", err)
	}
	deleted, err := ParseContactDeleteResult(result)
	if err != nil {
		t.Fatalf("ParseContactDeleteResult returned error: %v", err)
	}
	if !deleted.Deleted {
		t.Error("Deleted = false, want true")
	}
}

func TestContactsService_SearchRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/search", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{
			"type":"list",
			"data":[{"type":"contact","id":"1","email":"match@example.com"}],
			"total_count":1,
			"pages":{"type":"pages","page":1,"per_page":5,"total_pages":1}
		}`)
	})

	ctx := context.Background()
	result, err := client.Contacts.SearchRaw(ctx, &SearchRequest{
		Query: And(SingleFilterOf("created_at", ">", "1306054154")),
	})
	if err != nil {
		t.Fatalf("SearchRaw returned error: %v", err)
	}
	page, err := ParseContactSearchResult(result)
	if err != nil {
		t.Fatalf("ParseContactSearchResult returned error: %v", err)
	}
	if len(page.Data) != 1 {
		t.Errorf("Data len = %d, want 1", len(page.Data))
	}
}

func TestContactsService_MergeRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/merge", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"contact","id":"user-1","role":"user"}`)
	})

	ctx := context.Background()
	result, err := client.Contacts.MergeRaw(ctx, &MergeContactsRequest{From: "lead-1", Into: "user-1"})
	if err != nil {
		t.Fatalf("MergeRaw returned error: %v", err)
	}
	contact, err := ParseContactMergeResult(result)
	if err != nil {
		t.Fatalf("ParseContactMergeResult returned error: %v", err)
	}
	if contact.ID != "user-1" {
		t.Errorf("Contact.ID = %v, want user-1", contact.ID)
	}
}

func TestContactsService_ArchiveRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/archive", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"contact","id":"123","archived":true}`)
	})

	ctx := context.Background()
	result, err := client.Contacts.ArchiveRaw(ctx, "123")
	if err != nil {
		t.Fatalf("ArchiveRaw returned error: %v", err)
	}
	archived, err := ParseContactArchiveResult(result)
	if err != nil {
		t.Fatalf("ParseContactArchiveResult returned error: %v", err)
	}
	if !archived.Archived {
		t.Error("Archived = false, want true")
	}
}

func TestContactsService_UnarchiveRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/unarchive", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"contact","id":"123","archived":false}`)
	})

	ctx := context.Background()
	result, err := client.Contacts.UnarchiveRaw(ctx, "123")
	if err != nil {
		t.Fatalf("UnarchiveRaw returned error: %v", err)
	}
	unarchived, err := ParseContactUnarchiveResult(result)
	if err != nil {
		t.Fatalf("ParseContactUnarchiveResult returned error: %v", err)
	}
	if unarchived.Archived {
		t.Error("Archived = true, want false")
	}
}

func TestContactsService_BlockRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/block", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"contact","id":"123","blocked":true}`)
	})

	ctx := context.Background()
	result, err := client.Contacts.BlockRaw(ctx, "123")
	if err != nil {
		t.Fatalf("BlockRaw returned error: %v", err)
	}
	blocked, err := ParseContactBlockResult(result)
	if err != nil {
		t.Fatalf("ParseContactBlockResult returned error: %v", err)
	}
	if !blocked.Blocked {
		t.Error("Blocked = false, want true")
	}
}

func TestContactsService_FindByExternalIDRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/find_by_external_id/ext-42", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"type":"contact","id":"abc","external_id":"ext-42"}`)
	})

	ctx := context.Background()
	result, err := client.Contacts.FindByExternalIDRaw(ctx, "ext-42")
	if err != nil {
		t.Fatalf("FindByExternalIDRaw returned error: %v", err)
	}
	contact, err := ParseContactFindByExternalIDResult(result)
	if err != nil {
		t.Fatalf("ParseContactFindByExternalIDResult returned error: %v", err)
	}
	if contact.ExternalID != "ext-42" {
		t.Errorf("Contact.ExternalID = %v, want ext-42", contact.ExternalID)
	}
}

func TestContactsService_ListCompaniesRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/companies", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"list",
			"data":[{"type":"company","id":"c1","name":"Acme"}],
			"total_count":1,
			"pages":{"type":"pages","page":1,"per_page":50,"total_pages":1}
		}`)
	})

	ctx := context.Background()
	result, err := client.Contacts.ListCompaniesRaw(ctx, "123")
	if err != nil {
		t.Fatalf("ListCompaniesRaw returned error: %v", err)
	}
	companies, err := ParseContactListCompaniesResult(result)
	if err != nil {
		t.Fatalf("ParseContactListCompaniesResult returned error: %v", err)
	}
	if len(companies.Data) != 1 {
		t.Errorf("Data len = %d, want 1", len(companies.Data))
	}
}

func TestContactsService_AddCompanyRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/companies", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"company","id":"c1","name":"Acme"}`)
	})

	ctx := context.Background()
	result, err := client.Contacts.AddCompanyRaw(ctx, "123", "c1")
	if err != nil {
		t.Fatalf("AddCompanyRaw returned error: %v", err)
	}
	company, err := ParseContactAddCompanyResult(result)
	if err != nil {
		t.Fatalf("ParseContactAddCompanyResult returned error: %v", err)
	}
	if company.ID != "c1" {
		t.Errorf("Company.ID = %v, want c1", company.ID)
	}
}

func TestContactsService_RemoveCompanyRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/companies/c1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{"type":"company","id":"c1","name":"Acme"}`)
	})

	ctx := context.Background()
	result, err := client.Contacts.RemoveCompanyRaw(ctx, "123", "c1")
	if err != nil {
		t.Fatalf("RemoveCompanyRaw returned error: %v", err)
	}
	company, err := ParseContactRemoveCompanyResult(result)
	if err != nil {
		t.Fatalf("ParseContactRemoveCompanyResult returned error: %v", err)
	}
	if company.ID != "c1" {
		t.Errorf("Company.ID = %v, want c1", company.ID)
	}
}

func TestContactsService_ListNotesRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/notes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"list",
			"data":[{"type":"note","id":"n1","body":"<p>Hello</p>"}],
			"total_count":1,
			"pages":{"type":"pages","page":1,"per_page":50,"total_pages":1}
		}`)
	})

	ctx := context.Background()
	result, err := client.Contacts.ListNotesRaw(ctx, "123")
	if err != nil {
		t.Fatalf("ListNotesRaw returned error: %v", err)
	}
	notes, err := ParseContactListNotesResult(result)
	if err != nil {
		t.Fatalf("ParseContactListNotesResult returned error: %v", err)
	}
	if len(notes.Data) != 1 {
		t.Errorf("Data len = %d, want 1", len(notes.Data))
	}
}

func TestContactsService_CreateNoteRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/notes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"note","id":"n1","body":"<p>Hello</p>"}`)
	})

	ctx := context.Background()
	result, err := client.Contacts.CreateNoteRaw(ctx, "123", &CreateNoteRequest{Body: "Hello"})
	if err != nil {
		t.Fatalf("CreateNoteRaw returned error: %v", err)
	}
	note, err := ParseContactCreateNoteResult(result)
	if err != nil {
		t.Fatalf("ParseContactCreateNoteResult returned error: %v", err)
	}
	if note.ID != "n1" {
		t.Errorf("Note.ID = %v, want n1", note.ID)
	}
}

func TestContactsService_ListSegmentsRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/segments", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"list",
			"data":[{"type":"segment","id":"s1","name":"Active Users"}]
		}`)
	})

	ctx := context.Background()
	result, err := client.Contacts.ListSegmentsRaw(ctx, "123")
	if err != nil {
		t.Fatalf("ListSegmentsRaw returned error: %v", err)
	}
	segments, err := ParseContactListSegmentsResult(result)
	if err != nil {
		t.Fatalf("ParseContactListSegmentsResult returned error: %v", err)
	}
	if len(segments.Data) != 1 {
		t.Errorf("Data len = %d, want 1", len(segments.Data))
	}
}

func TestContactsService_ListSubscriptionsRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/subscriptions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"list",
			"data":[{"type":"subscription","id":"sub-1","state":"live","consent_type":"opt_in"}]
		}`)
	})

	ctx := context.Background()
	result, err := client.Contacts.ListSubscriptionsRaw(ctx, "123")
	if err != nil {
		t.Fatalf("ListSubscriptionsRaw returned error: %v", err)
	}
	subs, err := ParseContactListSubscriptionsResult(result)
	if err != nil {
		t.Fatalf("ParseContactListSubscriptionsResult returned error: %v", err)
	}
	if len(subs.Data) != 1 {
		t.Errorf("Data len = %d, want 1", len(subs.Data))
	}
}

func TestContactsService_AddSubscriptionRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/subscriptions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"subscription","id":"sub-1","state":"live","consent_type":"opt_in"}`)
	})

	ctx := context.Background()
	result, err := client.Contacts.AddSubscriptionRaw(ctx, "123", &AddSubscriptionRequest{
		ID:          "sub-1",
		ConsentType: "opt_in",
	})
	if err != nil {
		t.Fatalf("AddSubscriptionRaw returned error: %v", err)
	}
	sub, err := ParseContactAddSubscriptionResult(result)
	if err != nil {
		t.Fatalf("ParseContactAddSubscriptionResult returned error: %v", err)
	}
	if sub.ID != "sub-1" {
		t.Errorf("Subscription.ID = %v, want sub-1", sub.ID)
	}
}

func TestContactsService_RemoveSubscriptionRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/subscriptions/sub-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{"type":"subscription","id":"sub-1","state":"live","consent_type":"opt_in"}`)
	})

	ctx := context.Background()
	result, err := client.Contacts.RemoveSubscriptionRaw(ctx, "123", "sub-1")
	if err != nil {
		t.Fatalf("RemoveSubscriptionRaw returned error: %v", err)
	}
	sub, err := ParseContactRemoveSubscriptionResult(result)
	if err != nil {
		t.Fatalf("ParseContactRemoveSubscriptionResult returned error: %v", err)
	}
	if sub.ID != "sub-1" {
		t.Errorf("Subscription.ID = %v, want sub-1", sub.ID)
	}
}

func TestContactsService_AddTagRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/tags", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"tag","id":"tag-1","name":"VIP"}`)
	})

	ctx := context.Background()
	result, err := client.Contacts.AddTagRaw(ctx, "123", "tag-1")
	if err != nil {
		t.Fatalf("AddTagRaw returned error: %v", err)
	}
	tag, err := ParseContactAddTagResult(result)
	if err != nil {
		t.Fatalf("ParseContactAddTagResult returned error: %v", err)
	}
	if tag.ID != "tag-1" {
		t.Errorf("Tag.ID = %v, want tag-1", tag.ID)
	}
}

func TestContactsService_RemoveTagRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/tags/tag-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{"type":"tag","id":"tag-1","name":"VIP"}`)
	})

	ctx := context.Background()
	result, err := client.Contacts.RemoveTagRaw(ctx, "123", "tag-1")
	if err != nil {
		t.Fatalf("RemoveTagRaw returned error: %v", err)
	}
	tag, err := ParseContactRemoveTagResult(result)
	if err != nil {
		t.Fatalf("ParseContactRemoveTagResult returned error: %v", err)
	}
	if tag.ID != "tag-1" {
		t.Errorf("Tag.ID = %v, want tag-1", tag.ID)
	}
}

func TestContactsService_Get_NotFound(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/999", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"User Not Found"}]}`)
	})

	ctx := context.Background()
	_, err := client.Contacts.Get(ctx, "999")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("IsNotFound = false, want true")
	}
}

func TestContactsService_List_RateLimit(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		fmt.Fprint(w, `{"type":"error.list","errors":[{"code":"rate_limit","message":"Rate limit exceeded"}]}`)
	})

	ctx := context.Background()
	_, err := client.Contacts.List(ctx, nil)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !IsRateLimited(err) {
		t.Errorf("IsRateLimited = false, want true")
	}
}

func TestContactsService_ListTags(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/tags", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"list",
			"data":[
				{"type":"tag","id":"80","name":"Manual tag","applied_at":1663597223,"applied_by":{"type":"admin","id":"456"}},
				{"type":"tag","id":"81","name":"Auto tag"}
			]
		}`)
	})

	ctx := context.Background()
	result, err := client.Contacts.ListTags(ctx, "123")
	if err != nil {
		t.Fatalf("Contacts.ListTags returned error: %v", err)
	}
	if result.Type != "list" {
		t.Errorf("TagList.Type = %v, want list", result.Type)
	}
	if len(result.Data) != 2 {
		t.Fatalf("ListTags returned %d tags, want 2", len(result.Data))
	}
	if result.Data[0].Name != "Manual tag" {
		t.Errorf("Tag[0].Name = %v, want Manual tag", result.Data[0].Name)
	}
	if result.Data[0].AppliedAt == nil || *result.Data[0].AppliedAt != 1663597223 {
		t.Errorf("Tag[0].AppliedAt = %v, want 1663597223", result.Data[0].AppliedAt)
	}
}

func TestContactsService_ListTags_NotFound(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/999/tags", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"User Not Found"}]}`)
	})

	ctx := context.Background()
	_, err := client.Contacts.ListTags(ctx, "999")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("IsNotFound = false, want true")
	}
}

func TestContactsService_ListTagsRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/tags", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"list",
			"data":[{"type":"tag","id":"80","name":"Manual tag"}]
		}`)
	})

	ctx := context.Background()
	result, err := client.Contacts.ListTagsRaw(ctx, "123")
	if err != nil {
		t.Fatalf("ListTagsRaw returned error: %v", err)
	}
	tags, err := ParseContactListTagsResult(result)
	if err != nil {
		t.Fatalf("ParseContactListTagsResult returned error: %v", err)
	}
	if len(tags.Data) != 1 {
		t.Errorf("Data len = %d, want 1", len(tags.Data))
	}
}
