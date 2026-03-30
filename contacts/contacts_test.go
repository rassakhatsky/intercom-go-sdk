package contacts_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rassakhatsky/intercom-go-sdk/contacts"
	"github.com/rassakhatsky/intercom-go-sdk/api"
)

// testCaller implements api.Caller for testing, backed by an httptest.Server.
type testCaller struct {
	baseURL string
	client  *http.Client
}

func (tc *testCaller) NewRequest(method, urlStr string, body any) (*http.Request, error) {
	var buf io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, err
		}
		buf = bytes.NewBuffer(jsonBody)
	}
	req, err := http.NewRequest(method, tc.baseURL+"/"+urlStr, buf)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Authorization", "Bearer test-token")
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	return req, nil
}

func (tc *testCaller) DoRaw(ctx context.Context, req *http.Request) (*api.Result, error) {
	req = req.WithContext(ctx)
	resp, err := tc.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}
	return api.BuildResult(resp, b), nil
}

func (tc *testCaller) Do(ctx context.Context, req *http.Request, v any) (*api.Response, error) {
	result, err := tc.DoRaw(ctx, req)
	if err != nil {
		return nil, err
	}
	response := &api.Response{Result: result}
	if result.Error != nil {
		return response, api.ResultError(result)
	}
	if v != nil && result.StatusCode != http.StatusNoContent && len(result.Body) > 0 {
		if err := json.Unmarshal(result.Body, v); err != nil {
			return response, err
		}
	}
	return response, nil
}

func (tc *testCaller) DoRawNoRedirect(ctx context.Context, req *http.Request) (*api.Result, error) {
	return tc.DoRaw(ctx, req)
}

func (tc *testCaller) DoDownload(ctx context.Context, req *http.Request, w io.Writer) error {
	req = req.WithContext(ctx)
	resp, err := tc.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 400 {
		b, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return fmt.Errorf("HTTP %d: failed to read error body: %w", resp.StatusCode, readErr)
		}
		result := api.BuildResult(resp, b)
		return api.ResultError(result)
	}
	_, err = io.Copy(w, resp.Body)
	return err
}

func setup() (svc *contacts.Service, mux *http.ServeMux, teardown func()) {
	mux = http.NewServeMux()
	server := httptest.NewServer(mux)
	caller := &testCaller{baseURL: server.URL, client: server.Client()}
	svc = contacts.NewService(caller)
	return svc, mux, server.Close
}

func setupVisitors() (svc *contacts.VisitorsService, mux *http.ServeMux, teardown func()) {
	mux = http.NewServeMux()
	server := httptest.NewServer(mux)
	caller := &testCaller{baseURL: server.URL, client: server.Client()}
	svc = contacts.NewVisitorsService(caller)
	return svc, mux, server.Close
}

func testMethod(t *testing.T, r *http.Request, want string) {
	t.Helper()
	if got := r.Method; got != want {
		t.Errorf("Request method = %v, want %v", got, want)
	}
}

func testHeader(t *testing.T, r *http.Request, header, want string) {
	t.Helper()
	if got := r.Header.Get(header); got != want {
		t.Errorf("Header %v = %q, want %q", header, got, want)
	}
}

func TestService_Get(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testHeader(t, r, "Authorization", "Bearer test-token")
		fmt.Fprint(w, `{"type":"contact","id":"123","role":"user","email":"joe@example.com","name":"Joe"}`)
	})

	ctx := context.Background()
	contact, err := svc.Get(ctx, "123")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
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

func TestService_Create(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body contacts.CreateRequest
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
	contact, err := svc.Create(ctx, &contacts.CreateRequest{
		Role:  "user",
		Email: "joe@example.com",
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if contact.ID != "456" {
		t.Errorf("Contact.ID = %v, want 456", contact.ID)
	}
}

func TestService_Update(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		var body contacts.UpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.Name != "Joe Updated" {
			t.Errorf("Update body name = %v, want Joe Updated", body.Name)
		}
		fmt.Fprint(w, `{"type":"contact","id":"123","name":"Joe Updated","email":"joe@example.com"}`)
	})

	ctx := context.Background()
	contact, err := svc.Update(ctx, "123", &contacts.UpdateRequest{
		Name: "Joe Updated",
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if contact.Name != "Joe Updated" {
		t.Errorf("Contact.Name = %v, want Joe Updated", contact.Name)
	}
}

func TestService_Delete(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{"type":"contact","id":"123","external_id":"ext-1","deleted":true}`)
	})

	ctx := context.Background()
	deleted, err := svc.Delete(ctx, "123")
	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if !deleted.Deleted {
		t.Error("Deleted.Deleted = false, want true")
	}
	if deleted.ID != "123" {
		t.Errorf("Deleted.ID = %v, want 123", deleted.ID)
	}
}

func TestService_List(t *testing.T) {
	svc, mux, teardown := setup()
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
	result, err := svc.List(ctx, nil)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(result.Data) != 2 {
		t.Errorf("List returned %d contacts, want 2", len(result.Data))
	}
	if result.TotalCount != 2 {
		t.Errorf("TotalCount = %d, want 2", result.TotalCount)
	}
}

func TestService_ListAll(t *testing.T) {
	svc, mux, teardown := setup()
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
	iter := svc.ListAll(ctx, &api.ListOptions{PerPage: 1})
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

func TestService_Search(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/search", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body api.SearchRequest
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
	result, err := svc.Search(ctx, &api.SearchRequest{
		Query: api.And(
			api.SingleFilterOf("created_at", api.OpGreaterThan, "1306054154"),
		),
		Pagination: &api.SearchPagination{PerPage: 5},
	})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if len(result.Data) != 1 {
		t.Errorf("Search returned %d contacts, want 1", len(result.Data))
	}
}

func TestService_Merge(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/merge", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body contacts.MergeRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.From != "lead-1" || body.Into != "user-1" {
			t.Errorf("Merge body = %+v, want from=lead-1 into=user-1", body)
		}
		fmt.Fprint(w, `{"type":"contact","id":"user-1","role":"user"}`)
	})

	ctx := context.Background()
	contact, err := svc.Merge(ctx, &contacts.MergeRequest{
		From: "lead-1",
		Into: "user-1",
	})
	if err != nil {
		t.Fatalf("Merge returned error: %v", err)
	}
	if contact.ID != "user-1" {
		t.Errorf("Contact.ID = %v, want user-1", contact.ID)
	}
}

func TestService_Archive(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/archive", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"contact","id":"123","archived":true}`)
	})

	ctx := context.Background()
	result, err := svc.Archive(ctx, "123")
	if err != nil {
		t.Fatalf("Archive returned error: %v", err)
	}
	if !result.Archived {
		t.Error("Archived.Archived = false, want true")
	}
}

func TestService_Unarchive(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/unarchive", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"contact","id":"123","archived":false}`)
	})

	ctx := context.Background()
	result, err := svc.Unarchive(ctx, "123")
	if err != nil {
		t.Fatalf("Unarchive returned error: %v", err)
	}
	if result.Archived {
		t.Error("Unarchived.Archived = true, want false")
	}
}

func TestService_Block(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/block", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"contact","id":"123","blocked":true}`)
	})

	ctx := context.Background()
	result, err := svc.Block(ctx, "123")
	if err != nil {
		t.Fatalf("Block returned error: %v", err)
	}
	if !result.Blocked {
		t.Error("Blocked.Blocked = false, want true")
	}
}

func TestService_FindByExternalID(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/find_by_external_id/ext-42", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"type":"contact","id":"abc","external_id":"ext-42","email":"joe@example.com"}`)
	})

	ctx := context.Background()
	contact, err := svc.FindByExternalID(ctx, "ext-42")
	if err != nil {
		t.Fatalf("FindByExternalID returned error: %v", err)
	}
	if contact.ExternalID != "ext-42" {
		t.Errorf("Contact.ExternalID = %v, want ext-42", contact.ExternalID)
	}
}

func TestService_ListCompanies(t *testing.T) {
	svc, mux, teardown := setup()
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
	result, err := svc.ListCompanies(ctx, "123")
	if err != nil {
		t.Fatalf("ListCompanies returned error: %v", err)
	}
	if len(result.Data) != 1 {
		t.Errorf("ListCompanies returned %d companies, want 1", len(result.Data))
	}
	if result.Data[0].Name != "Acme" {
		t.Errorf("Company.Name = %v, want Acme", result.Data[0].Name)
	}
}

func TestService_AddCompany(t *testing.T) {
	svc, mux, teardown := setup()
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
	company, err := svc.AddCompany(ctx, "123", "c1")
	if err != nil {
		t.Fatalf("AddCompany returned error: %v", err)
	}
	if company.ID != "c1" {
		t.Errorf("Company.ID = %v, want c1", company.ID)
	}
}

func TestService_RemoveCompany(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/companies/c1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{"type":"company","id":"c1","name":"Acme"}`)
	})

	ctx := context.Background()
	company, err := svc.RemoveCompany(ctx, "123", "c1")
	if err != nil {
		t.Fatalf("RemoveCompany returned error: %v", err)
	}
	if company.ID != "c1" {
		t.Errorf("Company.ID = %v, want c1", company.ID)
	}
}

func TestService_ListNotes(t *testing.T) {
	svc, mux, teardown := setup()
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
	result, err := svc.ListNotes(ctx, "123")
	if err != nil {
		t.Fatalf("ListNotes returned error: %v", err)
	}
	if len(result.Data) != 1 {
		t.Errorf("ListNotes returned %d notes, want 1", len(result.Data))
	}
}

func TestService_CreateNote(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/notes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body contacts.CreateNoteRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.Body != "Hello" {
			t.Errorf("CreateNote body = %v, want Hello", body.Body)
		}
		fmt.Fprint(w, `{"type":"note","id":"n1","body":"<p>Hello</p>"}`)
	})

	ctx := context.Background()
	note, err := svc.CreateNote(ctx, "123", &contacts.CreateNoteRequest{
		Body:    "Hello",
		AdminID: "admin-1",
	})
	if err != nil {
		t.Fatalf("CreateNote returned error: %v", err)
	}
	if note.ID != "n1" {
		t.Errorf("Note.ID = %v, want n1", note.ID)
	}
}

func TestService_ListSegments(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/segments", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"list",
			"data":[{"type":"segment","id":"s1","name":"Active Users"}]
		}`)
	})

	ctx := context.Background()
	result, err := svc.ListSegments(ctx, "123")
	if err != nil {
		t.Fatalf("ListSegments returned error: %v", err)
	}
	if len(result.Data) != 1 {
		t.Errorf("ListSegments returned %d segments, want 1", len(result.Data))
	}
	if result.Data[0].Name != "Active Users" {
		t.Errorf("Segment.Name = %v, want Active Users", result.Data[0].Name)
	}
}

func TestService_ListSubscriptions(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/subscriptions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"list",
			"data":[{"type":"subscription","id":"sub-1","state":"live","consent_type":"opt_in"}]
		}`)
	})

	ctx := context.Background()
	result, err := svc.ListSubscriptions(ctx, "123")
	if err != nil {
		t.Fatalf("ListSubscriptions returned error: %v", err)
	}
	if len(result.Data) != 1 {
		t.Errorf("ListSubscriptions returned %d subscriptions, want 1", len(result.Data))
	}
}

func TestService_AddSubscription(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/subscriptions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body contacts.AddSubscriptionRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.ID != "sub-1" || body.ConsentType != "opt_in" {
			t.Errorf("AddSubscription body = %+v", body)
		}
		fmt.Fprint(w, `{"type":"subscription","id":"sub-1","state":"live","consent_type":"opt_in"}`)
	})

	ctx := context.Background()
	sub, err := svc.AddSubscription(ctx, "123", &contacts.AddSubscriptionRequest{
		ID:          "sub-1",
		ConsentType: "opt_in",
	})
	if err != nil {
		t.Fatalf("AddSubscription returned error: %v", err)
	}
	if sub.ID != "sub-1" {
		t.Errorf("Subscription.ID = %v, want sub-1", sub.ID)
	}
}

func TestService_RemoveSubscription(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/subscriptions/sub-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{"type":"subscription","id":"sub-1","state":"live","consent_type":"opt_in"}`)
	})

	ctx := context.Background()
	sub, err := svc.RemoveSubscription(ctx, "123", "sub-1")
	if err != nil {
		t.Fatalf("RemoveSubscription returned error: %v", err)
	}
	if sub.ID != "sub-1" {
		t.Errorf("Subscription.ID = %v, want sub-1", sub.ID)
	}
}

func TestService_AddTag(t *testing.T) {
	svc, mux, teardown := setup()
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
	tag, err := svc.AddTag(ctx, "123", "tag-1")
	if err != nil {
		t.Fatalf("AddTag returned error: %v", err)
	}
	if tag.ID != "tag-1" {
		t.Errorf("Tag.ID = %v, want tag-1", tag.ID)
	}
	if tag.Name != "VIP" {
		t.Errorf("Tag.Name = %v, want VIP", tag.Name)
	}
}

func TestService_RemoveTag(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/tags/tag-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{"type":"tag","id":"tag-1","name":"VIP"}`)
	})

	ctx := context.Background()
	tag, err := svc.RemoveTag(ctx, "123", "tag-1")
	if err != nil {
		t.Fatalf("RemoveTag returned error: %v", err)
	}
	if tag.ID != "tag-1" {
		t.Errorf("Tag.ID = %v, want tag-1", tag.ID)
	}
}

func TestService_ListTags(t *testing.T) {
	svc, mux, teardown := setup()
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
	result, err := svc.ListTags(ctx, "123")
	if err != nil {
		t.Fatalf("ListTags returned error: %v", err)
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

func TestService_Get_NotFound(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/999", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"User Not Found"}]}`)
	})

	ctx := context.Background()
	_, err := svc.Get(ctx, "999")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !api.IsNotFound(err) {
		t.Errorf("IsNotFound = false, want true")
	}
}

func TestService_List_RateLimit(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		fmt.Fprint(w, `{"type":"error.list","errors":[{"code":"rate_limit","message":"Rate limit exceeded"}]}`)
	})

	ctx := context.Background()
	_, err := svc.List(ctx, nil)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !api.IsRateLimited(err) {
		t.Errorf("IsRateLimited = false, want true")
	}
}

// --- Raw method tests ---

func TestService_GetRaw_Success(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-abc")
		fmt.Fprint(w, `{"type":"contact","id":"123","email":"joe@example.com"}`)
	})

	ctx := context.Background()
	result, err := svc.GetRaw(ctx, "123")
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
	contact, err := contacts.ParseGetResult(result)
	if err != nil {
		t.Fatalf("ParseGetResult returned error: %v", err)
	}
	if contact.ID != "123" {
		t.Errorf("Contact.ID = %v, want 123", contact.ID)
	}
	if contact.Email != "joe@example.com" {
		t.Errorf("Contact.Email = %v, want joe@example.com", contact.Email)
	}
}

func TestService_GetRaw_NotFound(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/999", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"User Not Found"}]}`)
	})

	ctx := context.Background()
	result, err := svc.GetRaw(ctx, "999")
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

func TestService_ListRaw(t *testing.T) {
	svc, mux, teardown := setup()
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
	result, err := svc.ListRaw(ctx, nil)
	if err != nil {
		t.Fatalf("ListRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	page, err := contacts.ParseListResult(result)
	if err != nil {
		t.Fatalf("ParseListResult returned error: %v", err)
	}
	if len(page.Data) != 2 {
		t.Errorf("Data len = %d, want 2", len(page.Data))
	}
}

func TestService_CreateRaw(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, `{"type":"contact","id":"456","role":"user","email":"joe@example.com"}`)
	})

	ctx := context.Background()
	result, err := svc.CreateRaw(ctx, &contacts.CreateRequest{
		Role:  "user",
		Email: "joe@example.com",
	})
	if err != nil {
		t.Fatalf("CreateRaw returned error: %v", err)
	}
	contact, err := contacts.ParseCreateResult(result)
	if err != nil {
		t.Fatalf("ParseCreateResult returned error: %v", err)
	}
	if contact.ID != "456" {
		t.Errorf("Contact.ID = %v, want 456", contact.ID)
	}
}

func TestService_UpdateRaw(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `{"type":"contact","id":"123","name":"Joe Updated"}`)
	})

	ctx := context.Background()
	result, err := svc.UpdateRaw(ctx, "123", &contacts.UpdateRequest{Name: "Joe Updated"})
	if err != nil {
		t.Fatalf("UpdateRaw returned error: %v", err)
	}
	contact, err := contacts.ParseUpdateResult(result)
	if err != nil {
		t.Fatalf("ParseUpdateResult returned error: %v", err)
	}
	if contact.Name != "Joe Updated" {
		t.Errorf("Contact.Name = %v, want Joe Updated", contact.Name)
	}
}

func TestService_DeleteRaw(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{"type":"contact","id":"123","deleted":true}`)
	})

	ctx := context.Background()
	result, err := svc.DeleteRaw(ctx, "123")
	if err != nil {
		t.Fatalf("DeleteRaw returned error: %v", err)
	}
	deleted, err := contacts.ParseDeleteResult(result)
	if err != nil {
		t.Fatalf("ParseDeleteResult returned error: %v", err)
	}
	if !deleted.Deleted {
		t.Error("Deleted = false, want true")
	}
}

func TestService_SearchRaw(t *testing.T) {
	svc, mux, teardown := setup()
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
	result, err := svc.SearchRaw(ctx, &api.SearchRequest{
		Query: api.And(api.SingleFilterOf("created_at", api.OpGreaterThan, "1306054154")),
	})
	if err != nil {
		t.Fatalf("SearchRaw returned error: %v", err)
	}
	page, err := contacts.ParseSearchResult(result)
	if err != nil {
		t.Fatalf("ParseSearchResult returned error: %v", err)
	}
	if len(page.Data) != 1 {
		t.Errorf("Data len = %d, want 1", len(page.Data))
	}
}

func TestService_MergeRaw(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/merge", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"contact","id":"user-1","role":"user"}`)
	})

	ctx := context.Background()
	result, err := svc.MergeRaw(ctx, &contacts.MergeRequest{From: "lead-1", Into: "user-1"})
	if err != nil {
		t.Fatalf("MergeRaw returned error: %v", err)
	}
	contact, err := contacts.ParseMergeResult(result)
	if err != nil {
		t.Fatalf("ParseMergeResult returned error: %v", err)
	}
	if contact.ID != "user-1" {
		t.Errorf("Contact.ID = %v, want user-1", contact.ID)
	}
}

func TestService_ArchiveRaw(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/archive", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"contact","id":"123","archived":true}`)
	})

	ctx := context.Background()
	result, err := svc.ArchiveRaw(ctx, "123")
	if err != nil {
		t.Fatalf("ArchiveRaw returned error: %v", err)
	}
	archived, err := contacts.ParseArchiveResult(result)
	if err != nil {
		t.Fatalf("ParseArchiveResult returned error: %v", err)
	}
	if !archived.Archived {
		t.Error("Archived = false, want true")
	}
}

func TestService_UnarchiveRaw(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/unarchive", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"contact","id":"123","archived":false}`)
	})

	ctx := context.Background()
	result, err := svc.UnarchiveRaw(ctx, "123")
	if err != nil {
		t.Fatalf("UnarchiveRaw returned error: %v", err)
	}
	unarchived, err := contacts.ParseUnarchiveResult(result)
	if err != nil {
		t.Fatalf("ParseUnarchiveResult returned error: %v", err)
	}
	if unarchived.Archived {
		t.Error("Archived = true, want false")
	}
}

func TestService_BlockRaw(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/block", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"contact","id":"123","blocked":true}`)
	})

	ctx := context.Background()
	result, err := svc.BlockRaw(ctx, "123")
	if err != nil {
		t.Fatalf("BlockRaw returned error: %v", err)
	}
	blocked, err := contacts.ParseBlockResult(result)
	if err != nil {
		t.Fatalf("ParseBlockResult returned error: %v", err)
	}
	if !blocked.Blocked {
		t.Error("Blocked = false, want true")
	}
}

func TestService_FindByExternalIDRaw(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/find_by_external_id/ext-42", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{"type":"contact","id":"abc","external_id":"ext-42"}`)
	})

	ctx := context.Background()
	result, err := svc.FindByExternalIDRaw(ctx, "ext-42")
	if err != nil {
		t.Fatalf("FindByExternalIDRaw returned error: %v", err)
	}
	contact, err := contacts.ParseFindByExternalIDResult(result)
	if err != nil {
		t.Fatalf("ParseFindByExternalIDResult returned error: %v", err)
	}
	if contact.ExternalID != "ext-42" {
		t.Errorf("Contact.ExternalID = %v, want ext-42", contact.ExternalID)
	}
}

func TestService_ListCompaniesRaw(t *testing.T) {
	svc, mux, teardown := setup()
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
	result, err := svc.ListCompaniesRaw(ctx, "123")
	if err != nil {
		t.Fatalf("ListCompaniesRaw returned error: %v", err)
	}
	companies, err := contacts.ParseListCompaniesResult(result)
	if err != nil {
		t.Fatalf("ParseListCompaniesResult returned error: %v", err)
	}
	if len(companies.Data) != 1 {
		t.Errorf("Data len = %d, want 1", len(companies.Data))
	}
}

func TestService_AddCompanyRaw(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/companies", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"company","id":"c1","name":"Acme"}`)
	})

	ctx := context.Background()
	result, err := svc.AddCompanyRaw(ctx, "123", "c1")
	if err != nil {
		t.Fatalf("AddCompanyRaw returned error: %v", err)
	}
	company, err := contacts.ParseAddCompanyResult(result)
	if err != nil {
		t.Fatalf("ParseAddCompanyResult returned error: %v", err)
	}
	if company.ID != "c1" {
		t.Errorf("Company.ID = %v, want c1", company.ID)
	}
}

func TestService_RemoveCompanyRaw(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/companies/c1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{"type":"company","id":"c1","name":"Acme"}`)
	})

	ctx := context.Background()
	result, err := svc.RemoveCompanyRaw(ctx, "123", "c1")
	if err != nil {
		t.Fatalf("RemoveCompanyRaw returned error: %v", err)
	}
	company, err := contacts.ParseRemoveCompanyResult(result)
	if err != nil {
		t.Fatalf("ParseRemoveCompanyResult returned error: %v", err)
	}
	if company.ID != "c1" {
		t.Errorf("Company.ID = %v, want c1", company.ID)
	}
}

func TestService_ListNotesRaw(t *testing.T) {
	svc, mux, teardown := setup()
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
	result, err := svc.ListNotesRaw(ctx, "123")
	if err != nil {
		t.Fatalf("ListNotesRaw returned error: %v", err)
	}
	notes, err := contacts.ParseListNotesResult(result)
	if err != nil {
		t.Fatalf("ParseListNotesResult returned error: %v", err)
	}
	if len(notes.Data) != 1 {
		t.Errorf("Data len = %d, want 1", len(notes.Data))
	}
}

func TestService_CreateNoteRaw(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/notes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"note","id":"n1","body":"<p>Hello</p>"}`)
	})

	ctx := context.Background()
	result, err := svc.CreateNoteRaw(ctx, "123", &contacts.CreateNoteRequest{Body: "Hello"})
	if err != nil {
		t.Fatalf("CreateNoteRaw returned error: %v", err)
	}
	note, err := contacts.ParseCreateNoteResult(result)
	if err != nil {
		t.Fatalf("ParseCreateNoteResult returned error: %v", err)
	}
	if note.ID != "n1" {
		t.Errorf("Note.ID = %v, want n1", note.ID)
	}
}

func TestService_ListSegmentsRaw(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/segments", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"list",
			"data":[{"type":"segment","id":"s1","name":"Active Users"}]
		}`)
	})

	ctx := context.Background()
	result, err := svc.ListSegmentsRaw(ctx, "123")
	if err != nil {
		t.Fatalf("ListSegmentsRaw returned error: %v", err)
	}
	segments, err := contacts.ParseListSegmentsResult(result)
	if err != nil {
		t.Fatalf("ParseListSegmentsResult returned error: %v", err)
	}
	if len(segments.Data) != 1 {
		t.Errorf("Data len = %d, want 1", len(segments.Data))
	}
}

func TestService_ListSubscriptionsRaw(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/subscriptions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"list",
			"data":[{"type":"subscription","id":"sub-1","state":"live","consent_type":"opt_in"}]
		}`)
	})

	ctx := context.Background()
	result, err := svc.ListSubscriptionsRaw(ctx, "123")
	if err != nil {
		t.Fatalf("ListSubscriptionsRaw returned error: %v", err)
	}
	subs, err := contacts.ParseListSubscriptionsResult(result)
	if err != nil {
		t.Fatalf("ParseListSubscriptionsResult returned error: %v", err)
	}
	if len(subs.Data) != 1 {
		t.Errorf("Data len = %d, want 1", len(subs.Data))
	}
}

func TestService_AddSubscriptionRaw(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/subscriptions", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"subscription","id":"sub-1","state":"live","consent_type":"opt_in"}`)
	})

	ctx := context.Background()
	result, err := svc.AddSubscriptionRaw(ctx, "123", &contacts.AddSubscriptionRequest{
		ID:          "sub-1",
		ConsentType: "opt_in",
	})
	if err != nil {
		t.Fatalf("AddSubscriptionRaw returned error: %v", err)
	}
	sub, err := contacts.ParseAddSubscriptionResult(result)
	if err != nil {
		t.Fatalf("ParseAddSubscriptionResult returned error: %v", err)
	}
	if sub.ID != "sub-1" {
		t.Errorf("Subscription.ID = %v, want sub-1", sub.ID)
	}
}

func TestService_RemoveSubscriptionRaw(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/subscriptions/sub-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{"type":"subscription","id":"sub-1","state":"live","consent_type":"opt_in"}`)
	})

	ctx := context.Background()
	result, err := svc.RemoveSubscriptionRaw(ctx, "123", "sub-1")
	if err != nil {
		t.Fatalf("RemoveSubscriptionRaw returned error: %v", err)
	}
	sub, err := contacts.ParseRemoveSubscriptionResult(result)
	if err != nil {
		t.Fatalf("ParseRemoveSubscriptionResult returned error: %v", err)
	}
	if sub.ID != "sub-1" {
		t.Errorf("Subscription.ID = %v, want sub-1", sub.ID)
	}
}

func TestService_AddTagRaw(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/tags", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"tag","id":"tag-1","name":"VIP"}`)
	})

	ctx := context.Background()
	result, err := svc.AddTagRaw(ctx, "123", "tag-1")
	if err != nil {
		t.Fatalf("AddTagRaw returned error: %v", err)
	}
	tag, err := contacts.ParseAddTagResult(result)
	if err != nil {
		t.Fatalf("ParseAddTagResult returned error: %v", err)
	}
	if tag.ID != "tag-1" {
		t.Errorf("Tag.ID = %v, want tag-1", tag.ID)
	}
}

func TestService_RemoveTagRaw(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/tags/tag-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{"type":"tag","id":"tag-1","name":"VIP"}`)
	})

	ctx := context.Background()
	result, err := svc.RemoveTagRaw(ctx, "123", "tag-1")
	if err != nil {
		t.Fatalf("RemoveTagRaw returned error: %v", err)
	}
	tag, err := contacts.ParseRemoveTagResult(result)
	if err != nil {
		t.Fatalf("ParseRemoveTagResult returned error: %v", err)
	}
	if tag.ID != "tag-1" {
		t.Errorf("Tag.ID = %v, want tag-1", tag.ID)
	}
}

func TestService_ListTagsRaw(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/123/tags", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"list",
			"data":[{"type":"tag","id":"80","name":"Manual tag"}]
		}`)
	})

	ctx := context.Background()
	result, err := svc.ListTagsRaw(ctx, "123")
	if err != nil {
		t.Fatalf("ListTagsRaw returned error: %v", err)
	}
	tags, err := contacts.ParseListTagsResult(result)
	if err != nil {
		t.Fatalf("ParseListTagsResult returned error: %v", err)
	}
	if len(tags.Data) != 1 {
		t.Errorf("Data len = %d, want 1", len(tags.Data))
	}
}

func TestService_ListTags_NotFound(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/contacts/999/tags", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"User Not Found"}]}`)
	})

	ctx := context.Background()
	_, err := svc.ListTags(ctx, "999")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !api.IsNotFound(err) {
		t.Errorf("IsNotFound = false, want true")
	}
}

// --- Visitor tests ---

func TestVisitorsService_Get(t *testing.T) {
	svc, mux, teardown := setupVisitors()
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
	visitor, err := svc.Get(ctx, "visitor-uid-1")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
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
	svc, mux, teardown := setupVisitors()
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
	_, err := svc.Get(ctx, "nonexistent")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !api.IsNotFound(err) {
		t.Errorf("Expected IsNotFound, got: %v", err)
	}
}

func TestVisitorsService_Update(t *testing.T) {
	svc, mux, teardown := setupVisitors()
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
	visitor, err := svc.Update(ctx, &contacts.UpdateVisitorRequest{
		ID:   "530370b477ad7120001d",
		Name: "Updated Name",
		CustomAttributes: map[string]any{
			"plan": "premium",
		},
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if visitor.Name != "Updated Name" {
		t.Errorf("Name = %v, want Updated Name", visitor.Name)
	}
	if visitor.UpdatedAt != 1663597999 {
		t.Errorf("UpdatedAt = %v, want 1663597999", visitor.UpdatedAt)
	}
}

func TestVisitorsService_Update_ByUserID(t *testing.T) {
	svc, mux, teardown := setupVisitors()
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
	visitor, err := svc.Update(ctx, &contacts.UpdateVisitorRequest{
		UserID: "visitor-uid-1",
		Name:   "Updated Via UserID",
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if visitor.Name != "Updated Via UserID" {
		t.Errorf("Name = %v, want Updated Via UserID", visitor.Name)
	}
}

func TestVisitorsService_Convert(t *testing.T) {
	svc, mux, teardown := setupVisitors()
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
	contact, err := svc.Convert(ctx, &contacts.ConvertVisitorRequest{
		Type: "user",
		Visitor: contacts.ConvertVisitorIdentifier{
			ID: "530370b477ad7120001d",
		},
		User: contacts.ConvertVisitorUser{
			Email: "jane@example.com",
		},
	})
	if err != nil {
		t.Fatalf("Convert returned error: %v", err)
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
	svc, mux, teardown := setupVisitors()
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
	result, err := svc.GetRaw(ctx, "visitor-uid-1")
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
	visitor, err := contacts.ParseVisitorGetResult(result)
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
	svc, mux, teardown := setupVisitors()
	defer teardown()

	mux.HandleFunc("/visitors", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"Visitor Not Found"}]}`)
	})

	ctx := context.Background()
	result, err := svc.GetRaw(ctx, "nonexistent")
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
	svc, mux, teardown := setupVisitors()
	defer teardown()

	mux.HandleFunc("/visitors", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		w.Header().Set("X-Request-Id", "req-visitor-update")
		fmt.Fprint(w, `{"type":"visitor","id":"530370b477ad7120001d","name":"Updated Name"}`)
	})

	ctx := context.Background()
	result, err := svc.UpdateRaw(ctx, &contacts.UpdateVisitorRequest{
		ID:   "530370b477ad7120001d",
		Name: "Updated Name",
	})
	if err != nil {
		t.Fatalf("UpdateRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	visitor, err := contacts.ParseVisitorUpdateResult(result)
	if err != nil {
		t.Fatalf("ParseVisitorUpdateResult returned error: %v", err)
	}
	if visitor.Name != "Updated Name" {
		t.Errorf("Name = %v, want Updated Name", visitor.Name)
	}
}

func TestVisitorsService_ConvertRaw_Success(t *testing.T) {
	svc, mux, teardown := setupVisitors()
	defer teardown()

	mux.HandleFunc("/visitors/convert", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.Header().Set("X-Request-Id", "req-visitor-convert")
		fmt.Fprint(w, `{"type":"contact","id":"converted-contact-id","role":"user","email":"jane@example.com"}`)
	})

	ctx := context.Background()
	result, err := svc.ConvertRaw(ctx, &contacts.ConvertVisitorRequest{
		Type:    "user",
		Visitor: contacts.ConvertVisitorIdentifier{ID: "530370b477ad7120001d"},
		User:    contacts.ConvertVisitorUser{Email: "jane@example.com"},
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
	contact, err := contacts.ParseVisitorConvertResult(result)
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
	svc, mux, teardown := setupVisitors()
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
	contact, err := svc.Convert(ctx, &contacts.ConvertVisitorRequest{
		Type: "lead",
		Visitor: contacts.ConvertVisitorIdentifier{
			UserID: "visitor-uid-1",
		},
		User: contacts.ConvertVisitorUser{
			UserID: "external-user-id",
		},
	})
	if err != nil {
		t.Fatalf("Convert returned error: %v", err)
	}
	if contact.Role != "lead" {
		t.Errorf("Role = %v, want lead", contact.Role)
	}
}

func TestVisitorsService_Convert_Unauthorized(t *testing.T) {
	svc, mux, teardown := setupVisitors()
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
	_, err := svc.Convert(ctx, &contacts.ConvertVisitorRequest{
		Type:    "user",
		Visitor: contacts.ConvertVisitorIdentifier{ID: "some-id"},
		User:    contacts.ConvertVisitorUser{Email: "test@example.com"},
	})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !api.IsUnauthorized(err) {
		t.Errorf("Expected IsUnauthorized, got: %v", err)
	}
}
