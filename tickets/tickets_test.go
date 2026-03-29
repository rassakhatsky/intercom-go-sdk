package tickets_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rassakhatsky/intercom-go-sdk/internal/api"
	"github.com/rassakhatsky/intercom-go-sdk/tickets"
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

func setupTickets() (svc *tickets.Service, mux *http.ServeMux, teardown func()) {
	mux = http.NewServeMux()
	server := httptest.NewServer(mux)
	caller := &testCaller{baseURL: server.URL, client: server.Client()}
	svc = tickets.NewService(caller)
	return svc, mux, server.Close
}

func setupTypes() (svc *tickets.TypesService, mux *http.ServeMux, teardown func()) {
	mux = http.NewServeMux()
	server := httptest.NewServer(mux)
	caller := &testCaller{baseURL: server.URL, client: server.Client()}
	svc = tickets.NewTypesService(caller)
	return svc, mux, server.Close
}

func setupStates() (svc *tickets.StatesService, mux *http.ServeMux, teardown func()) {
	mux = http.NewServeMux()
	server := httptest.NewServer(mux)
	caller := &testCaller{baseURL: server.URL, client: server.Client()}
	svc = tickets.NewStatesService(caller)
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

// --- Tickets Service Tests ---

func TestService_Get(t *testing.T) {
	svc, mux, teardown := setupTickets()
	defer teardown()

	mux.HandleFunc("/tickets/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testHeader(t, r, "Authorization", "Bearer test-token")
		fmt.Fprint(w, `{
			"type":"ticket",
			"id":"123",
			"ticket_id":"#1001",
			"category":"Customer",
			"created_at":1734537460,
			"updated_at":1734537460,
			"open":true,
			"is_shared":true,
			"ticket_attributes":{"_default_title_":"Help needed","_default_description_":"I need help"},
			"ticket_state":{"type":"ticket_state","id":"ts-1","category":"submitted","internal_label":"New","external_label":"Submitted"},
			"ticket_type":{"type":"ticket_type","id":"tt-1","name":"Bug Report"},
			"contacts":{"type":"contact.list","contacts":[{"type":"contact","id":"c-1"}]},
			"admin_assignee_id":"admin-1",
			"team_assignee_id":"team-1",
			"ticket_parts":{"type":"ticket_part.list","ticket_parts":[{"type":"ticket_part","id":"tp-1","part_type":"comment","body":"<p>Hello</p>"}],"total_count":1},
			"linked_objects":{"type":"list","data":[],"total_count":0,"has_more":false}
		}`)
	})

	ctx := context.Background()
	ticket, err := svc.Get(ctx, "123")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if ticket.ID != "123" {
		t.Errorf("Ticket.ID = %v, want 123", ticket.ID)
	}
	if ticket.TicketID != "#1001" {
		t.Errorf("Ticket.TicketID = %v, want #1001", ticket.TicketID)
	}
	if ticket.Category != "Customer" {
		t.Errorf("Ticket.Category = %v, want Customer", ticket.Category)
	}
	if !ticket.Open {
		t.Error("Ticket.Open = false, want true")
	}
	if !ticket.IsShared {
		t.Error("Ticket.IsShared = false, want true")
	}
	if ticket.TicketAttributes["_default_title_"] != "Help needed" {
		t.Errorf("Ticket.TicketAttributes[_default_title_] = %v, want Help needed", ticket.TicketAttributes["_default_title_"])
	}
	if ticket.TicketState == nil || ticket.TicketState.ID != "ts-1" {
		t.Error("Ticket.TicketState unexpected")
	}
	if ticket.TicketState.Category != "submitted" {
		t.Errorf("Ticket.TicketState.Category = %v, want submitted", ticket.TicketState.Category)
	}
	if ticket.TicketType == nil || ticket.TicketType.Name != "Bug Report" {
		t.Error("Ticket.TicketType unexpected")
	}
	if ticket.Contacts == nil || len(ticket.Contacts.Contacts) != 1 {
		t.Error("Ticket.Contacts count unexpected")
	}
	if ticket.AdminAssigneeID != "admin-1" {
		t.Errorf("Ticket.AdminAssigneeID = %v, want admin-1", ticket.AdminAssigneeID)
	}
	if ticket.TeamAssigneeID != "team-1" {
		t.Errorf("Ticket.TeamAssigneeID = %v, want team-1", ticket.TeamAssigneeID)
	}
	if ticket.TicketParts == nil || len(ticket.TicketParts.Parts) != 1 {
		t.Error("Ticket.TicketParts count unexpected")
	}
	if ticket.TicketParts.Parts[0].PartType != "comment" {
		t.Errorf("TicketPart.PartType = %v, want comment", ticket.TicketParts.Parts[0].PartType)
	}
}

func TestService_Create(t *testing.T) {
	svc, mux, teardown := setupTickets()
	defer teardown()

	mux.HandleFunc("/tickets", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body tickets.CreateRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.TicketTypeID != "tt-1" {
			t.Errorf("Create body ticket_type_id = %v, want tt-1", body.TicketTypeID)
		}
		if len(body.Contacts) != 1 {
			t.Fatalf("Create body contacts length = %d, want 1", len(body.Contacts))
		}
		if body.Contacts[0].ID != "c-1" {
			t.Errorf("Create body contacts[0].id = %v, want c-1", body.Contacts[0].ID)
		}
		fmt.Fprint(w, `{"type":"ticket","id":"456","ticket_id":"#1002","category":"Customer","open":true}`)
	})

	ctx := context.Background()
	ticket, err := svc.Create(ctx, &tickets.CreateRequest{
		TicketTypeID: "tt-1",
		Contacts:     []tickets.ContactRef{{ID: "c-1"}},
		TicketAttributes: map[string]any{
			"_default_title_":       "New bug",
			"_default_description_": "Something is broken",
		},
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if ticket.ID != "456" {
		t.Errorf("Ticket.ID = %v, want 456", ticket.ID)
	}
	if ticket.TicketID != "#1002" {
		t.Errorf("Ticket.TicketID = %v, want #1002", ticket.TicketID)
	}
}

func TestService_Update(t *testing.T) {
	svc, mux, teardown := setupTickets()
	defer teardown()

	mux.HandleFunc("/tickets/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		var body tickets.UpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.TicketStateID != "ts-2" {
			t.Errorf("Update body ticket_state_id = %v, want ts-2", body.TicketStateID)
		}
		if body.AssigneeID != "admin-2" {
			t.Errorf("Update body assignee_id = %v, want admin-2", body.AssigneeID)
		}
		fmt.Fprint(w, `{"type":"ticket","id":"123","open":true,"admin_assignee_id":"admin-2"}`)
	})

	ctx := context.Background()
	ticket, err := svc.Update(ctx, "123", &tickets.UpdateRequest{
		TicketStateID: "ts-2",
		AssigneeID:    "admin-2",
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if ticket.ID != "123" {
		t.Errorf("Ticket.ID = %v, want 123", ticket.ID)
	}
	if ticket.AdminAssigneeID != "admin-2" {
		t.Errorf("Ticket.AdminAssigneeID = %v, want admin-2", ticket.AdminAssigneeID)
	}
}

func TestService_Delete(t *testing.T) {
	svc, mux, teardown := setupTickets()
	defer teardown()

	mux.HandleFunc("/tickets/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{"id":"123","object":"ticket","deleted":true}`)
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

func TestService_Search(t *testing.T) {
	svc, mux, teardown := setupTickets()
	defer teardown()

	mux.HandleFunc("/tickets/search", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body api.SearchRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.Query.Operator != "AND" {
			t.Errorf("Search query operator = %v, want AND", body.Query.Operator)
		}
		fmt.Fprint(w, `{
			"type":"ticket.list",
			"tickets":[
				{"type":"ticket","id":"1","open":true},
				{"type":"ticket","id":"2","open":false}
			],
			"total_count":2,
			"pages":{"type":"pages","page":1,"per_page":20,"total_pages":1}
		}`)
	})

	ctx := context.Background()
	result, err := svc.Search(ctx, &api.SearchRequest{
		Query:      api.And(api.SingleFilterOf("open", api.OpEquals, "true")),
		Pagination: &api.SearchPagination{PerPage: 20},
	})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if len(result.Data) != 2 {
		t.Errorf("Search returned %d tickets, want 2", len(result.Data))
	}
	if result.TotalCount != 2 {
		t.Errorf("TotalCount = %d, want 2", result.TotalCount)
	}
}

func TestService_Reply_Contact(t *testing.T) {
	svc, mux, teardown := setupTickets()
	defer teardown()

	mux.HandleFunc("/tickets/123/reply", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body tickets.ReplyRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.MessageType != "comment" {
			t.Errorf("Reply body message_type = %v, want comment", body.MessageType)
		}
		if body.Type != "user" {
			t.Errorf("Reply body type = %v, want user", body.Type)
		}
		if body.Body != "I still need help" {
			t.Errorf("Reply body body = %v, want I still need help", body.Body)
		}
		if body.IntercomUserID != "c-1" {
			t.Errorf("Reply body intercom_user_id = %v, want c-1", body.IntercomUserID)
		}
		fmt.Fprint(w, `{"type":"ticket_part","id":"tp-2","part_type":"comment","body":"<p>I still need help</p>"}`)
	})

	ctx := context.Background()
	part, err := svc.Reply(ctx, "123", &tickets.ReplyRequest{
		MessageType:    "comment",
		Type:           "user",
		Body:           "I still need help",
		IntercomUserID: "c-1",
	})
	if err != nil {
		t.Fatalf("Reply returned error: %v", err)
	}
	if part.ID != "tp-2" {
		t.Errorf("Part.ID = %v, want tp-2", part.ID)
	}
	if part.PartType != "comment" {
		t.Errorf("Part.PartType = %v, want comment", part.PartType)
	}
}

func TestService_Reply_Admin(t *testing.T) {
	svc, mux, teardown := setupTickets()
	defer teardown()

	mux.HandleFunc("/tickets/123/reply", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body tickets.ReplyRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.MessageType != "note" {
			t.Errorf("Reply body message_type = %v, want note", body.MessageType)
		}
		if body.Type != "admin" {
			t.Errorf("Reply body type = %v, want admin", body.Type)
		}
		if body.AdminID != "admin-1" {
			t.Errorf("Reply body admin_id = %v, want admin-1", body.AdminID)
		}
		fmt.Fprint(w, `{"type":"ticket_part","id":"tp-3","part_type":"note","body":"<p>Internal note</p>"}`)
	})

	ctx := context.Background()
	part, err := svc.Reply(ctx, "123", &tickets.ReplyRequest{
		MessageType: "note",
		Type:        "admin",
		AdminID:     "admin-1",
		Body:        "<p>Internal note</p>",
	})
	if err != nil {
		t.Fatalf("Reply returned error: %v", err)
	}
	if part.ID != "tp-3" {
		t.Errorf("Part.ID = %v, want tp-3", part.ID)
	}
	if part.PartType != "note" {
		t.Errorf("Part.PartType = %v, want note", part.PartType)
	}
}

func TestService_AddTag(t *testing.T) {
	svc, mux, teardown := setupTickets()
	defer teardown()

	mux.HandleFunc("/tickets/123/tags", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body struct {
			ID      string `json:"id"`
			AdminID string `json:"admin_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.ID != "tag-1" {
			t.Errorf("AddTag body id = %v, want tag-1", body.ID)
		}
		if body.AdminID != "admin-1" {
			t.Errorf("AddTag body admin_id = %v, want admin-1", body.AdminID)
		}
		fmt.Fprint(w, `{"type":"tag","id":"tag-1","name":"Urgent"}`)
	})

	ctx := context.Background()
	tag, err := svc.AddTag(ctx, "123", "tag-1", "admin-1")
	if err != nil {
		t.Fatalf("AddTag returned error: %v", err)
	}
	if tag.ID != "tag-1" {
		t.Errorf("Tag.ID = %v, want tag-1", tag.ID)
	}
	if tag.Name != "Urgent" {
		t.Errorf("Tag.Name = %v, want Urgent", tag.Name)
	}
}

func TestService_RemoveTag(t *testing.T) {
	svc, mux, teardown := setupTickets()
	defer teardown()

	mux.HandleFunc("/tickets/123/tags/tag-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		var body struct {
			AdminID string `json:"admin_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.AdminID != "admin-1" {
			t.Errorf("RemoveTag body admin_id = %v, want admin-1", body.AdminID)
		}
		fmt.Fprint(w, `{"type":"tag","id":"tag-1","name":"Urgent"}`)
	})

	ctx := context.Background()
	tag, err := svc.RemoveTag(ctx, "123", "tag-1", "admin-1")
	if err != nil {
		t.Fatalf("RemoveTag returned error: %v", err)
	}
	if tag.ID != "tag-1" {
		t.Errorf("Tag.ID = %v, want tag-1", tag.ID)
	}
}

func TestService_Enqueue(t *testing.T) {
	svc, mux, teardown := setupTickets()
	defer teardown()

	mux.HandleFunc("/tickets/enqueue", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body tickets.CreateRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.TicketTypeID != "tt-1" {
			t.Errorf("Enqueue body ticket_type_id = %v, want tt-1", body.TicketTypeID)
		}
		if len(body.Contacts) != 1 {
			t.Fatalf("Enqueue body contacts length = %d, want 1", len(body.Contacts))
		}
		fmt.Fprint(w, `{"id":"job-1","status":"queued"}`)
	})

	ctx := context.Background()
	job, err := svc.Enqueue(ctx, &tickets.CreateRequest{
		TicketTypeID: "tt-1",
		Contacts:     []tickets.ContactRef{{ID: "c-1"}},
	})
	if err != nil {
		t.Fatalf("Enqueue returned error: %v", err)
	}
	if job.ID != "job-1" {
		t.Errorf("Job.ID = %v, want job-1", job.ID)
	}
	if job.Status != "queued" {
		t.Errorf("Job.Status = %v, want queued", job.Status)
	}
}

func TestService_Get_NotFound(t *testing.T) {
	svc, mux, teardown := setupTickets()
	defer teardown()

	mux.HandleFunc("/tickets/999", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"Resource Not Found"}]}`)
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

func TestService_GetRaw_Success(t *testing.T) {
	svc, mux, teardown := setupTickets()
	defer teardown()

	mux.HandleFunc("/tickets/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-tick-1")
		fmt.Fprint(w, `{"type":"ticket","id":"123","ticket_id":"#1001","open":true}`)
	})

	ctx := context.Background()
	result, err := svc.GetRaw(ctx, "123")
	if err != nil {
		t.Fatalf("GetRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-tick-1" {
		t.Errorf("Header X-Request-Id = %q, want req-tick-1", result.Header.Get("X-Request-Id"))
	}
	if result.Error != nil {
		t.Errorf("Result.Error = %v, want nil", result.Error)
	}
	data, err := tickets.ParseGetResult(result)
	if err != nil {
		t.Fatalf("ParseGetResult returned error: %v", err)
	}
	if data.ID != "123" {
		t.Errorf("Data.ID = %v, want 123", data.ID)
	}
	if data.TicketID != "#1001" {
		t.Errorf("Data.TicketID = %v, want #1001", data.TicketID)
	}
}

func TestService_GetRaw_NotFound(t *testing.T) {
	svc, mux, teardown := setupTickets()
	defer teardown()

	mux.HandleFunc("/tickets/999", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"Resource Not Found"}]}`)
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

func TestService_CreateRaw(t *testing.T) {
	svc, mux, teardown := setupTickets()
	defer teardown()

	mux.HandleFunc("/tickets", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"ticket","id":"456","ticket_id":"#1002","open":true}`)
	})

	ctx := context.Background()
	result, err := svc.CreateRaw(ctx, &tickets.CreateRequest{
		TicketTypeID: "tt-1",
		Contacts:     []tickets.ContactRef{{ID: "c-1"}},
	})
	if err != nil {
		t.Fatalf("CreateRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	data, err := tickets.ParseCreateResult(result)
	if err != nil {
		t.Fatalf("ParseCreateResult returned error: %v", err)
	}
	if data.ID != "456" {
		t.Errorf("Data.ID = %v, want 456", data.ID)
	}
}

func TestService_UpdateRaw(t *testing.T) {
	svc, mux, teardown := setupTickets()
	defer teardown()

	mux.HandleFunc("/tickets/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `{"type":"ticket","id":"123","admin_assignee_id":"admin-2"}`)
	})

	ctx := context.Background()
	result, err := svc.UpdateRaw(ctx, "123", &tickets.UpdateRequest{
		AssigneeID: "admin-2",
	})
	if err != nil {
		t.Fatalf("UpdateRaw returned error: %v", err)
	}
	data, err := tickets.ParseUpdateResult(result)
	if err != nil {
		t.Fatalf("ParseUpdateResult returned error: %v", err)
	}
	if data.ID != "123" {
		t.Errorf("Data.ID = %v, want 123", data.ID)
	}
	if data.AdminAssigneeID != "admin-2" {
		t.Errorf("Data.AdminAssigneeID = %v, want admin-2", data.AdminAssigneeID)
	}
}

func TestService_DeleteRaw(t *testing.T) {
	svc, mux, teardown := setupTickets()
	defer teardown()

	mux.HandleFunc("/tickets/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{"id":"123","object":"ticket","deleted":true}`)
	})

	ctx := context.Background()
	result, err := svc.DeleteRaw(ctx, "123")
	if err != nil {
		t.Fatalf("DeleteRaw returned error: %v", err)
	}
	data, err := tickets.ParseDeleteResult(result)
	if err != nil {
		t.Fatalf("ParseDeleteResult returned error: %v", err)
	}
	if !data.Deleted {
		t.Error("Data.Deleted = false, want true")
	}
}

func TestService_SearchRaw(t *testing.T) {
	svc, mux, teardown := setupTickets()
	defer teardown()

	mux.HandleFunc("/tickets/search", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{
			"type":"ticket.list",
			"tickets":[
				{"type":"ticket","id":"1","open":true},
				{"type":"ticket","id":"2","open":false}
			],
			"total_count":2,
			"pages":{"type":"pages","page":1,"per_page":20,"total_pages":1}
		}`)
	})

	ctx := context.Background()
	result, err := svc.SearchRaw(ctx, &api.SearchRequest{
		Query: api.And(api.SingleFilterOf("open", api.OpEquals, "true")),
	})
	if err != nil {
		t.Fatalf("SearchRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	data, err := tickets.ParseSearchResult(result)
	if err != nil {
		t.Fatalf("ParseSearchResult returned error: %v", err)
	}
	if len(data.Data) != 2 {
		t.Errorf("Data.Data length = %d, want 2", len(data.Data))
	}
	if data.TotalCount != 2 {
		t.Errorf("Data.TotalCount = %d, want 2", data.TotalCount)
	}
}

func TestService_ReplyRaw(t *testing.T) {
	svc, mux, teardown := setupTickets()
	defer teardown()

	mux.HandleFunc("/tickets/123/reply", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"ticket_part","id":"tp-2","part_type":"comment","body":"<p>Reply</p>"}`)
	})

	ctx := context.Background()
	result, err := svc.ReplyRaw(ctx, "123", &tickets.ReplyRequest{
		MessageType: "comment",
		Type:        "user",
		Body:        "Reply",
	})
	if err != nil {
		t.Fatalf("ReplyRaw returned error: %v", err)
	}
	data, err := tickets.ParseReplyResult(result)
	if err != nil {
		t.Fatalf("ParseReplyResult returned error: %v", err)
	}
	if data.ID != "tp-2" {
		t.Errorf("Data.ID = %v, want tp-2", data.ID)
	}
	if data.PartType != "comment" {
		t.Errorf("Data.PartType = %v, want comment", data.PartType)
	}
}

func TestService_EnqueueRaw(t *testing.T) {
	svc, mux, teardown := setupTickets()
	defer teardown()

	mux.HandleFunc("/tickets/enqueue", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"id":"job-1","status":"queued"}`)
	})

	ctx := context.Background()
	result, err := svc.EnqueueRaw(ctx, &tickets.CreateRequest{
		TicketTypeID: "tt-1",
		Contacts:     []tickets.ContactRef{{ID: "c-1"}},
	})
	if err != nil {
		t.Fatalf("EnqueueRaw returned error: %v", err)
	}
	data, err := tickets.ParseEnqueueResult(result)
	if err != nil {
		t.Fatalf("ParseEnqueueResult returned error: %v", err)
	}
	if data.ID != "job-1" {
		t.Errorf("Data.ID = %v, want job-1", data.ID)
	}
	if data.Status != "queued" {
		t.Errorf("Data.Status = %v, want queued", data.Status)
	}
}

func TestService_AddTagRaw(t *testing.T) {
	svc, mux, teardown := setupTickets()
	defer teardown()

	mux.HandleFunc("/tickets/123/tags", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"tag","id":"tag-1","name":"Urgent"}`)
	})

	ctx := context.Background()
	result, err := svc.AddTagRaw(ctx, "123", "tag-1", "admin-1")
	if err != nil {
		t.Fatalf("AddTagRaw returned error: %v", err)
	}
	data, err := tickets.ParseAddTagResult(result)
	if err != nil {
		t.Fatalf("ParseAddTagResult returned error: %v", err)
	}
	if data.ID != "tag-1" {
		t.Errorf("Data.ID = %v, want tag-1", data.ID)
	}
	if data.Name != "Urgent" {
		t.Errorf("Data.Name = %v, want Urgent", data.Name)
	}
}

func TestService_RemoveTagRaw(t *testing.T) {
	svc, mux, teardown := setupTickets()
	defer teardown()

	mux.HandleFunc("/tickets/123/tags/tag-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{"type":"tag","id":"tag-1","name":"Urgent"}`)
	})

	ctx := context.Background()
	result, err := svc.RemoveTagRaw(ctx, "123", "tag-1", "admin-1")
	if err != nil {
		t.Fatalf("RemoveTagRaw returned error: %v", err)
	}
	data, err := tickets.ParseRemoveTagResult(result)
	if err != nil {
		t.Fatalf("ParseRemoveTagResult returned error: %v", err)
	}
	if data.ID != "tag-1" {
		t.Errorf("Data.ID = %v, want tag-1", data.ID)
	}
}

func TestService_Search_RateLimit(t *testing.T) {
	svc, mux, teardown := setupTickets()
	defer teardown()

	mux.HandleFunc("/tickets/search", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		fmt.Fprint(w, `{"type":"error.list","errors":[{"code":"rate_limit","message":"Rate limit exceeded"}]}`)
	})

	ctx := context.Background()
	_, err := svc.Search(ctx, &api.SearchRequest{
		Query: api.And(api.SingleFilterOf("open", api.OpEquals, "true")),
	})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !api.IsRateLimited(err) {
		t.Errorf("IsRateLimited = false, want true")
	}
}

func TestService_Create_WithSkipNotifications(t *testing.T) {
	svc, mux, teardown := setupTickets()
	defer teardown()

	mux.HandleFunc("/tickets", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		skipNotif, ok := body["skip_notifications"]
		if !ok || skipNotif != true {
			t.Errorf("skip_notifications = %v, want true", skipNotif)
		}
		fmt.Fprint(w, `{"type":"ticket","id":"t-1","ticket_id":"123"}`)
	})

	ctx := context.Background()
	skipNotif := true
	ticket, err := svc.Create(ctx, &tickets.CreateRequest{
		TicketTypeID:      "tt-1",
		Contacts:          []tickets.ContactRef{{ID: "c-1"}},
		SkipNotifications: &skipNotif,
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if ticket.ID != "t-1" {
		t.Errorf("Ticket.ID = %v, want t-1", ticket.ID)
	}
}

func TestService_Update_WithSkipNotifications(t *testing.T) {
	svc, mux, teardown := setupTickets()
	defer teardown()

	mux.HandleFunc("/tickets/t-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		skipNotif, ok := body["skip_notifications"]
		if !ok || skipNotif != true {
			t.Errorf("skip_notifications = %v, want true", skipNotif)
		}
		fmt.Fprint(w, `{"type":"ticket","id":"t-1","ticket_id":"123"}`)
	})

	ctx := context.Background()
	skipNotif := true
	ticket, err := svc.Update(ctx, "t-1", &tickets.UpdateRequest{
		TicketStateID:     "ts-2",
		SkipNotifications: &skipNotif,
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if ticket.ID != "t-1" {
		t.Errorf("Ticket.ID = %v, want t-1", ticket.ID)
	}
}

// --- Ticket Types Tests ---

func TestTypesService_Get(t *testing.T) {
	svc, mux, teardown := setupTypes()
	defer teardown()

	mux.HandleFunc("/ticket_types/tt-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testHeader(t, r, "Authorization", "Bearer test-token")
		fmt.Fprint(w, `{
			"type":"ticket_type",
			"id":"tt-1",
			"category":"Customer",
			"name":"Bug Report",
			"description":"Report a bug",
			"icon":"🐛",
			"workspace_id":"ws-1",
			"archived":false,
			"created_at":1734537460,
			"updated_at":1734537460,
			"ticket_type_attributes":{
				"type":"list",
				"data":[
					{"type":"ticket_type_attribute","id":"attr-1","name":"Priority","description":"Bug priority","data_type":"list","order":0,"required_to_create":true,"visible_on_create":true,"visible_to_contacts":true,"default":false,"ticket_type_id":1,"archived":false}
				]
			}
		}`)
	})

	ctx := context.Background()
	tt, err := svc.Get(ctx, "tt-1")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if tt.ID != "tt-1" {
		t.Errorf("Type.ID = %v, want tt-1", tt.ID)
	}
	if tt.Name != "Bug Report" {
		t.Errorf("Type.Name = %v, want Bug Report", tt.Name)
	}
	if tt.Description != "Report a bug" {
		t.Errorf("Type.Description = %v, want Report a bug", tt.Description)
	}
	if tt.Category != "Customer" {
		t.Errorf("Type.Category = %v, want Customer", tt.Category)
	}
	if tt.Icon != "🐛" {
		t.Errorf("Type.Icon = %v, want 🐛", tt.Icon)
	}
	if tt.WorkspaceID != "ws-1" {
		t.Errorf("Type.WorkspaceID = %v, want ws-1", tt.WorkspaceID)
	}
	if tt.Archived {
		t.Error("Type.Archived = true, want false")
	}
	if tt.Attributes == nil || len(tt.Attributes.Data) != 1 {
		t.Fatal("Type.Attributes unexpected")
	}
	attr := tt.Attributes.Data[0]
	if attr.ID != "attr-1" {
		t.Errorf("Attribute.ID = %v, want attr-1", attr.ID)
	}
	if attr.Name != "Priority" {
		t.Errorf("Attribute.Name = %v, want Priority", attr.Name)
	}
	if attr.DataType != "list" {
		t.Errorf("Attribute.DataType = %v, want list", attr.DataType)
	}
	if !attr.RequiredToCreate {
		t.Error("Attribute.RequiredToCreate = false, want true")
	}
}

func TestTypesService_List(t *testing.T) {
	svc, mux, teardown := setupTypes()
	defer teardown()

	mux.HandleFunc("/ticket_types", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"list",
			"data":[
				{"type":"ticket_type","id":"tt-1","name":"Bug Report","category":"Customer"},
				{"type":"ticket_type","id":"tt-2","name":"Feature Request","category":"Customer"}
			]
		}`)
	})

	ctx := context.Background()
	list, err := svc.List(ctx)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(list.Data) != 2 {
		t.Fatalf("List returned %d ticket types, want 2", len(list.Data))
	}
	if list.Data[0].ID != "tt-1" {
		t.Errorf("Type[0].ID = %v, want tt-1", list.Data[0].ID)
	}
	if list.Data[1].Name != "Feature Request" {
		t.Errorf("Type[1].Name = %v, want Feature Request", list.Data[1].Name)
	}
}

func TestTypesService_Create(t *testing.T) {
	svc, mux, teardown := setupTypes()
	defer teardown()

	mux.HandleFunc("/ticket_types", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body tickets.CreateTypeRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.Name != "Bug Report" {
			t.Errorf("Create body name = %v, want Bug Report", body.Name)
		}
		if body.Description != "Report a bug" {
			t.Errorf("Create body description = %v, want Report a bug", body.Description)
		}
		if body.Category != "Customer" {
			t.Errorf("Create body category = %v, want Customer", body.Category)
		}
		if body.Icon != "🐛" {
			t.Errorf("Create body icon = %v, want 🐛", body.Icon)
		}
		fmt.Fprint(w, `{"type":"ticket_type","id":"tt-new","name":"Bug Report","description":"Report a bug","category":"Customer","icon":"🐛","archived":false}`)
	})

	ctx := context.Background()
	tt, err := svc.Create(ctx, &tickets.CreateTypeRequest{
		Name:        "Bug Report",
		Description: "Report a bug",
		Category:    "Customer",
		Icon:        "🐛",
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if tt.ID != "tt-new" {
		t.Errorf("Type.ID = %v, want tt-new", tt.ID)
	}
	if tt.Name != "Bug Report" {
		t.Errorf("Type.Name = %v, want Bug Report", tt.Name)
	}
}

func TestTypesService_Update(t *testing.T) {
	svc, mux, teardown := setupTypes()
	defer teardown()

	mux.HandleFunc("/ticket_types/tt-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		var body tickets.UpdateTypeRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.Name != "Critical Bug" {
			t.Errorf("Update body name = %v, want Critical Bug", body.Name)
		}
		if body.Description != "Report a critical bug" {
			t.Errorf("Update body description = %v, want Report a critical bug", body.Description)
		}
		fmt.Fprint(w, `{"type":"ticket_type","id":"tt-1","name":"Critical Bug","description":"Report a critical bug","category":"Customer"}`)
	})

	ctx := context.Background()
	tt, err := svc.Update(ctx, "tt-1", &tickets.UpdateTypeRequest{
		Name:        "Critical Bug",
		Description: "Report a critical bug",
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if tt.Name != "Critical Bug" {
		t.Errorf("Type.Name = %v, want Critical Bug", tt.Name)
	}
}

func TestTypesService_Update_Archive(t *testing.T) {
	svc, mux, teardown := setupTypes()
	defer teardown()

	mux.HandleFunc("/ticket_types/tt-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["archived"] != true {
			t.Errorf("Update body archived = %v, want true", body["archived"])
		}
		fmt.Fprint(w, `{"type":"ticket_type","id":"tt-1","name":"Bug Report","archived":true}`)
	})

	ctx := context.Background()
	archived := true
	tt, err := svc.Update(ctx, "tt-1", &tickets.UpdateTypeRequest{
		Archived: &archived,
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if !tt.Archived {
		t.Error("Type.Archived = false, want true")
	}
}

func TestTypesService_CreateAttribute(t *testing.T) {
	svc, mux, teardown := setupTypes()
	defer teardown()

	mux.HandleFunc("/ticket_types/tt-1/attributes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body tickets.CreateTypeAttributeRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.Name != "Priority" {
			t.Errorf("CreateAttribute body name = %v, want Priority", body.Name)
		}
		if body.Description != "Bug priority level" {
			t.Errorf("CreateAttribute body description = %v, want Bug priority level", body.Description)
		}
		if body.DataType != "list" {
			t.Errorf("CreateAttribute body data_type = %v, want list", body.DataType)
		}
		if body.RequiredToCreate == nil || !*body.RequiredToCreate {
			t.Error("CreateAttribute body required_to_create = false/nil, want true")
		}
		if body.VisibleOnCreate == nil || !*body.VisibleOnCreate {
			t.Error("CreateAttribute body visible_on_create = false/nil, want true")
		}
		fmt.Fprint(w, `{
			"type":"ticket_type_attribute",
			"id":"attr-new",
			"name":"Priority",
			"description":"Bug priority level",
			"data_type":"list",
			"required_to_create":true,
			"visible_on_create":true,
			"visible_to_contacts":false,
			"default":false,
			"order":1,
			"ticket_type_id":1,
			"archived":false,
			"created_at":1734537460,
			"updated_at":1734537460
		}`)
	})

	ctx := context.Background()
	boolTrue := true
	attr, err := svc.CreateAttribute(ctx, "tt-1", &tickets.CreateTypeAttributeRequest{
		Name:             "Priority",
		Description:      "Bug priority level",
		DataType:         "list",
		RequiredToCreate: &boolTrue,
		VisibleOnCreate:  &boolTrue,
	})
	if err != nil {
		t.Fatalf("CreateAttribute returned error: %v", err)
	}
	if attr.ID != "attr-new" {
		t.Errorf("Attribute.ID = %v, want attr-new", attr.ID)
	}
	if attr.Name != "Priority" {
		t.Errorf("Attribute.Name = %v, want Priority", attr.Name)
	}
	if attr.DataType != "list" {
		t.Errorf("Attribute.DataType = %v, want list", attr.DataType)
	}
	if !attr.RequiredToCreate {
		t.Error("Attribute.RequiredToCreate = false, want true")
	}
}

func TestTypesService_UpdateAttribute(t *testing.T) {
	svc, mux, teardown := setupTypes()
	defer teardown()

	mux.HandleFunc("/ticket_types/tt-1/attributes/attr-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		var body tickets.UpdateTypeAttributeRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.Name != "Severity" {
			t.Errorf("UpdateAttribute body name = %v, want Severity", body.Name)
		}
		if body.Description != "Bug severity level" {
			t.Errorf("UpdateAttribute body description = %v, want Bug severity level", body.Description)
		}
		fmt.Fprint(w, `{
			"type":"ticket_type_attribute",
			"id":"attr-1",
			"name":"Severity",
			"description":"Bug severity level",
			"data_type":"list",
			"required_to_create":true,
			"visible_on_create":true,
			"archived":false
		}`)
	})

	ctx := context.Background()
	attr, err := svc.UpdateAttribute(ctx, "tt-1", "attr-1", &tickets.UpdateTypeAttributeRequest{
		Name:        "Severity",
		Description: "Bug severity level",
	})
	if err != nil {
		t.Fatalf("UpdateAttribute returned error: %v", err)
	}
	if attr.Name != "Severity" {
		t.Errorf("Attribute.Name = %v, want Severity", attr.Name)
	}
	if attr.Description != "Bug severity level" {
		t.Errorf("Attribute.Description = %v, want Bug severity level", attr.Description)
	}
}

func TestTypesService_UpdateAttribute_Archive(t *testing.T) {
	svc, mux, teardown := setupTypes()
	defer teardown()

	mux.HandleFunc("/ticket_types/tt-1/attributes/attr-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["archived"] != true {
			t.Errorf("UpdateAttribute body archived = %v, want true", body["archived"])
		}
		fmt.Fprint(w, `{"type":"ticket_type_attribute","id":"attr-1","name":"Priority","archived":true}`)
	})

	ctx := context.Background()
	archived := true
	attr, err := svc.UpdateAttribute(ctx, "tt-1", "attr-1", &tickets.UpdateTypeAttributeRequest{
		Archived: &archived,
	})
	if err != nil {
		t.Fatalf("UpdateAttribute returned error: %v", err)
	}
	if !attr.Archived {
		t.Error("Attribute.Archived = false, want true")
	}
}

func TestTypesService_GetRaw_Success(t *testing.T) {
	svc, mux, teardown := setupTypes()
	defer teardown()

	mux.HandleFunc("/ticket_types/tt-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-tt-get")
		fmt.Fprint(w, `{"type":"ticket_type","id":"tt-1","name":"Bug Report","category":"Customer"}`)
	})

	ctx := context.Background()
	result, err := svc.GetRaw(ctx, "tt-1")
	if err != nil {
		t.Fatalf("GetRaw returned error: %v", err)
	}
	data, err := tickets.ParseTypeGetResult(result)
	if err != nil {
		t.Fatalf("ParseTypeGetResult returned error: %v", err)
	}
	if data.ID != "tt-1" {
		t.Errorf("Data.ID = %v, want tt-1", data.ID)
	}
	if data.Name != "Bug Report" {
		t.Errorf("Data.Name = %v, want Bug Report", data.Name)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-tt-get" {
		t.Errorf("Header X-Request-Id = %q, want req-tt-get", result.Header.Get("X-Request-Id"))
	}
	if result.Error != nil {
		t.Errorf("Result.Error = %v, want nil", result.Error)
	}
}

func TestTypesService_GetRaw_NotFound(t *testing.T) {
	svc, mux, teardown := setupTypes()
	defer teardown()

	mux.HandleFunc("/ticket_types/tt-999", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"Not found"}]}`)
	})

	ctx := context.Background()
	result, err := svc.GetRaw(ctx, "tt-999")
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

func TestTypesService_ListRaw_Success(t *testing.T) {
	svc, mux, teardown := setupTypes()
	defer teardown()

	mux.HandleFunc("/ticket_types", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-tt-list")
		fmt.Fprint(w, `{"type":"list","data":[{"type":"ticket_type","id":"tt-1","name":"Bug Report"},{"type":"ticket_type","id":"tt-2","name":"Feature Request"}]}`)
	})

	ctx := context.Background()
	result, err := svc.ListRaw(ctx)
	if err != nil {
		t.Fatalf("ListRaw returned error: %v", err)
	}
	data, err := tickets.ParseTypeListResult(result)
	if err != nil {
		t.Fatalf("ParseTypeListResult returned error: %v", err)
	}
	if len(data.Data) != 2 {
		t.Fatalf("Data.Data length = %d, want 2", len(data.Data))
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestTypesService_CreateRaw_Success(t *testing.T) {
	svc, mux, teardown := setupTypes()
	defer teardown()

	mux.HandleFunc("/ticket_types", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.Header().Set("X-Request-Id", "req-tt-create")
		fmt.Fprint(w, `{"type":"ticket_type","id":"tt-new","name":"Bug Report"}`)
	})

	ctx := context.Background()
	result, err := svc.CreateRaw(ctx, &tickets.CreateTypeRequest{Name: "Bug Report"})
	if err != nil {
		t.Fatalf("CreateRaw returned error: %v", err)
	}
	data, err := tickets.ParseTypeCreateResult(result)
	if err != nil {
		t.Fatalf("ParseTypeCreateResult returned error: %v", err)
	}
	if data.ID != "tt-new" {
		t.Errorf("Data.ID = %v, want tt-new", data.ID)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestTypesService_UpdateRaw_Success(t *testing.T) {
	svc, mux, teardown := setupTypes()
	defer teardown()

	mux.HandleFunc("/ticket_types/tt-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		w.Header().Set("X-Request-Id", "req-tt-update")
		fmt.Fprint(w, `{"type":"ticket_type","id":"tt-1","name":"Critical Bug"}`)
	})

	ctx := context.Background()
	result, err := svc.UpdateRaw(ctx, "tt-1", &tickets.UpdateTypeRequest{Name: "Critical Bug"})
	if err != nil {
		t.Fatalf("UpdateRaw returned error: %v", err)
	}
	data, err := tickets.ParseTypeUpdateResult(result)
	if err != nil {
		t.Fatalf("ParseTypeUpdateResult returned error: %v", err)
	}
	if data.Name != "Critical Bug" {
		t.Errorf("Data.Name = %v, want Critical Bug", data.Name)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestTypesService_CreateAttributeRaw_Success(t *testing.T) {
	svc, mux, teardown := setupTypes()
	defer teardown()

	mux.HandleFunc("/ticket_types/tt-1/attributes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.Header().Set("X-Request-Id", "req-tt-attr-create")
		fmt.Fprint(w, `{"type":"ticket_type_attribute","id":"attr-new","name":"Priority","data_type":"list"}`)
	})

	ctx := context.Background()
	result, err := svc.CreateAttributeRaw(ctx, "tt-1", &tickets.CreateTypeAttributeRequest{
		Name:     "Priority",
		DataType: "list",
	})
	if err != nil {
		t.Fatalf("CreateAttributeRaw returned error: %v", err)
	}
	data, err := tickets.ParseTypeCreateAttributeResult(result)
	if err != nil {
		t.Fatalf("ParseTypeCreateAttributeResult returned error: %v", err)
	}
	if data.ID != "attr-new" {
		t.Errorf("Data.ID = %v, want attr-new", data.ID)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestTypesService_UpdateAttributeRaw_Success(t *testing.T) {
	svc, mux, teardown := setupTypes()
	defer teardown()

	mux.HandleFunc("/ticket_types/tt-1/attributes/attr-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		w.Header().Set("X-Request-Id", "req-tt-attr-update")
		fmt.Fprint(w, `{"type":"ticket_type_attribute","id":"attr-1","name":"Severity","description":"Bug severity level"}`)
	})

	ctx := context.Background()
	result, err := svc.UpdateAttributeRaw(ctx, "tt-1", "attr-1", &tickets.UpdateTypeAttributeRequest{
		Name:        "Severity",
		Description: "Bug severity level",
	})
	if err != nil {
		t.Fatalf("UpdateAttributeRaw returned error: %v", err)
	}
	data, err := tickets.ParseTypeUpdateAttributeResult(result)
	if err != nil {
		t.Fatalf("ParseTypeUpdateAttributeResult returned error: %v", err)
	}
	if data.Name != "Severity" {
		t.Errorf("Data.Name = %v, want Severity", data.Name)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestTypesService_Get_NotFound(t *testing.T) {
	svc, mux, teardown := setupTypes()
	defer teardown()

	mux.HandleFunc("/ticket_types/tt-999", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"Resource Not Found"}]}`)
	})

	ctx := context.Background()
	_, err := svc.Get(ctx, "tt-999")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !api.IsNotFound(err) {
		t.Errorf("IsNotFound = false, want true")
	}
}

func TestTypesService_Create_Unauthorized(t *testing.T) {
	svc, mux, teardown := setupTypes()
	defer teardown()

	mux.HandleFunc("/ticket_types", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{"type":"error.list","errors":[{"code":"unauthorized","message":"Unauthorized"}]}`)
	})

	ctx := context.Background()
	_, err := svc.Create(ctx, &tickets.CreateTypeRequest{
		Name: "Test",
	})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !api.IsUnauthorized(err) {
		t.Errorf("IsUnauthorized = false, want true")
	}
}

// --- Ticket States Tests ---

func TestStatesService_List(t *testing.T) {
	svc, mux, teardown := setupStates()
	defer teardown()

	mux.HandleFunc("/ticket_states", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testHeader(t, r, "Authorization", "Bearer test-token")
		fmt.Fprint(w, `{
			"type":"list",
			"data":[
				{
					"type":"ticket_state",
					"id":"8269",
					"category":"submitted",
					"internal_label":"Submitted",
					"external_label":"Submitted",
					"archived":false
				},
				{
					"type":"ticket_state",
					"id":"8270",
					"category":"in_progress",
					"internal_label":"In progress",
					"external_label":"In progress",
					"archived":false
				},
				{
					"type":"ticket_state",
					"id":"8271",
					"category":"waiting_on_customer",
					"internal_label":"Waiting on customer",
					"external_label":"Waiting on customer",
					"archived":true
				}
			]
		}`)
	})

	ctx := context.Background()
	result, err := svc.List(ctx)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if result.Type != "list" {
		t.Errorf("Type = %v, want list", result.Type)
	}
	if len(result.Data) != 3 {
		t.Fatalf("Data length = %d, want 3", len(result.Data))
	}

	ts := result.Data[0]
	if ts.Type != "ticket_state" {
		t.Errorf("Data[0].Type = %v, want ticket_state", ts.Type)
	}
	if ts.ID != "8269" {
		t.Errorf("Data[0].ID = %v, want 8269", ts.ID)
	}
	if ts.Category != "submitted" {
		t.Errorf("Data[0].Category = %v, want submitted", ts.Category)
	}
	if ts.InternalLabel != "Submitted" {
		t.Errorf("Data[0].InternalLabel = %v, want Submitted", ts.InternalLabel)
	}
	if ts.ExternalLabel != "Submitted" {
		t.Errorf("Data[0].ExternalLabel = %v, want Submitted", ts.ExternalLabel)
	}
	if ts.Archived != false {
		t.Errorf("Data[0].Archived = %v, want false", ts.Archived)
	}

	ts2 := result.Data[1]
	if ts2.Category != "in_progress" {
		t.Errorf("Data[1].Category = %v, want in_progress", ts2.Category)
	}

	ts3 := result.Data[2]
	if ts3.Archived != true {
		t.Errorf("Data[2].Archived = %v, want true", ts3.Archived)
	}
}

func TestStatesService_ListRaw_Success(t *testing.T) {
	svc, mux, teardown := setupStates()
	defer teardown()

	mux.HandleFunc("/ticket_states", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-ts-list")
		fmt.Fprint(w, `{"type":"list","data":[{"type":"ticket_state","id":"8269","category":"submitted","internal_label":"Submitted","external_label":"Submitted","archived":false}]}`)
	})

	ctx := context.Background()
	result, err := svc.ListRaw(ctx)
	if err != nil {
		t.Fatalf("ListRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-ts-list" {
		t.Errorf("Header X-Request-Id = %q, want req-ts-list", result.Header.Get("X-Request-Id"))
	}
	tsl, err := tickets.ParseStateListResult(result)
	if err != nil {
		t.Fatalf("ParseStateListResult returned error: %v", err)
	}
	if len(tsl.Data) != 1 {
		t.Fatalf("Data length = %d, want 1", len(tsl.Data))
	}
	if tsl.Data[0].ID != "8269" {
		t.Errorf("Data[0].ID = %v, want 8269", tsl.Data[0].ID)
	}
	if tsl.Data[0].Category != "submitted" {
		t.Errorf("Data[0].Category = %v, want submitted", tsl.Data[0].Category)
	}
}

func TestStatesService_List_Unauthorized(t *testing.T) {
	svc, mux, teardown := setupStates()
	defer teardown()

	mux.HandleFunc("/ticket_states", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{
			"type":"error.list",
			"request_id":"req-456",
			"errors":[{"code":"unauthorized","message":"Unauthorized"}]
		}`)
	})

	ctx := context.Background()
	_, err := svc.List(ctx)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !api.IsUnauthorized(err) {
		t.Errorf("Expected IsUnauthorized, got: %v", err)
	}
}
