package intercom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestTicketsService_Get(t *testing.T) {
	client, mux, teardown := setup()
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
	ticket, err := client.Tickets.Get(ctx, "123")
	if err != nil {
		t.Fatalf("Tickets.Get returned error: %v", err)
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

func TestTicketsService_Create(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tickets", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body CreateTicketRequest
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
	ticket, err := client.Tickets.Create(ctx, &CreateTicketRequest{
		TicketTypeID: "tt-1",
		Contacts:     []TicketContactRef{{ID: "c-1"}},
		TicketAttributes: map[string]any{
			"_default_title_":       "New bug",
			"_default_description_": "Something is broken",
		},
	})
	if err != nil {
		t.Fatalf("Tickets.Create returned error: %v", err)
	}
	if ticket.ID != "456" {
		t.Errorf("Ticket.ID = %v, want 456", ticket.ID)
	}
	if ticket.TicketID != "#1002" {
		t.Errorf("Ticket.TicketID = %v, want #1002", ticket.TicketID)
	}
}

func TestTicketsService_Update(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tickets/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		var body UpdateTicketRequest
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
	ticket, err := client.Tickets.Update(ctx, "123", &UpdateTicketRequest{
		TicketStateID: "ts-2",
		AssigneeID:    "admin-2",
	})
	if err != nil {
		t.Fatalf("Tickets.Update returned error: %v", err)
	}
	if ticket.ID != "123" {
		t.Errorf("Ticket.ID = %v, want 123", ticket.ID)
	}
	if ticket.AdminAssigneeID != "admin-2" {
		t.Errorf("Ticket.AdminAssigneeID = %v, want admin-2", ticket.AdminAssigneeID)
	}
}

func TestTicketsService_Delete(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tickets/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{"id":"123","object":"ticket","deleted":true}`)
	})

	ctx := context.Background()
	deleted, err := client.Tickets.Delete(ctx, "123")
	if err != nil {
		t.Fatalf("Tickets.Delete returned error: %v", err)
	}
	if !deleted.Deleted {
		t.Error("TicketDeleted.Deleted = false, want true")
	}
	if deleted.ID != "123" {
		t.Errorf("TicketDeleted.ID = %v, want 123", deleted.ID)
	}
}

func TestTicketsService_Search(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tickets/search", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body SearchRequest
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
	result, err := client.Tickets.Search(ctx, &SearchRequest{
		Query:      And(SingleFilterOf("open", "=", "true")),
		Pagination: &SearchPagination{PerPage: 20},
	})
	if err != nil {
		t.Fatalf("Tickets.Search returned error: %v", err)
	}
	if len(result.Data) != 2 {
		t.Errorf("Search returned %d tickets, want 2", len(result.Data))
	}
	if result.TotalCount != 2 {
		t.Errorf("TotalCount = %d, want 2", result.TotalCount)
	}
}

func TestTicketsService_Reply_Contact(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tickets/123/reply", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body ReplyTicketRequest
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
	part, err := client.Tickets.Reply(ctx, "123", &ReplyTicketRequest{
		MessageType:    "comment",
		Type:           "user",
		Body:           "I still need help",
		IntercomUserID: "c-1",
	})
	if err != nil {
		t.Fatalf("Tickets.Reply returned error: %v", err)
	}
	if part.ID != "tp-2" {
		t.Errorf("TicketPart.ID = %v, want tp-2", part.ID)
	}
	if part.PartType != "comment" {
		t.Errorf("TicketPart.PartType = %v, want comment", part.PartType)
	}
}

func TestTicketsService_Reply_Admin(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tickets/123/reply", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body ReplyTicketRequest
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
	part, err := client.Tickets.Reply(ctx, "123", &ReplyTicketRequest{
		MessageType: "note",
		Type:        "admin",
		AdminID:     "admin-1",
		Body:        "<p>Internal note</p>",
	})
	if err != nil {
		t.Fatalf("Tickets.Reply returned error: %v", err)
	}
	if part.ID != "tp-3" {
		t.Errorf("TicketPart.ID = %v, want tp-3", part.ID)
	}
	if part.PartType != "note" {
		t.Errorf("TicketPart.PartType = %v, want note", part.PartType)
	}
}

func TestTicketsService_AddTag(t *testing.T) {
	client, mux, teardown := setup()
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
	tag, err := client.Tickets.AddTag(ctx, "123", "tag-1", "admin-1")
	if err != nil {
		t.Fatalf("Tickets.AddTag returned error: %v", err)
	}
	if tag.ID != "tag-1" {
		t.Errorf("Tag.ID = %v, want tag-1", tag.ID)
	}
	if tag.Name != "Urgent" {
		t.Errorf("Tag.Name = %v, want Urgent", tag.Name)
	}
}

func TestTicketsService_RemoveTag(t *testing.T) {
	client, mux, teardown := setup()
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
	tag, err := client.Tickets.RemoveTag(ctx, "123", "tag-1", "admin-1")
	if err != nil {
		t.Fatalf("Tickets.RemoveTag returned error: %v", err)
	}
	if tag.ID != "tag-1" {
		t.Errorf("Tag.ID = %v, want tag-1", tag.ID)
	}
}

func TestTicketsService_Enqueue(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tickets/enqueue", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body CreateTicketRequest
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
	job, err := client.Tickets.Enqueue(ctx, &CreateTicketRequest{
		TicketTypeID: "tt-1",
		Contacts:     []TicketContactRef{{ID: "c-1"}},
	})
	if err != nil {
		t.Fatalf("Tickets.Enqueue returned error: %v", err)
	}
	if job.ID != "job-1" {
		t.Errorf("Job.ID = %v, want job-1", job.ID)
	}
	if job.Status != "queued" {
		t.Errorf("Job.Status = %v, want queued", job.Status)
	}
}

func TestTicketsService_Get_NotFound(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tickets/999", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"Resource Not Found"}]}`)
	})

	ctx := context.Background()
	_, err := client.Tickets.Get(ctx, "999")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("IsNotFound = false, want true")
	}
}

// --- Raw method tests ---

func TestTicketsService_GetRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tickets/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-tick-1")
		fmt.Fprint(w, `{"type":"ticket","id":"123","ticket_id":"#1001","open":true}`)
	})

	ctx := context.Background()
	result, err := client.Tickets.GetRaw(ctx, "123")
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
	data, err := ParseTicketGetResult(result)
	if err != nil {
		t.Fatalf("ParseTicketGetResult returned error: %v", err)
	}
	if data.ID != "123" {
		t.Errorf("Data.ID = %v, want 123", data.ID)
	}
	if data.TicketID != "#1001" {
		t.Errorf("Data.TicketID = %v, want #1001", data.TicketID)
	}
}

func TestTicketsService_GetRaw_NotFound(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tickets/999", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"Resource Not Found"}]}`)
	})

	ctx := context.Background()
	result, err := client.Tickets.GetRaw(ctx, "999")
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

func TestTicketsService_CreateRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tickets", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"ticket","id":"456","ticket_id":"#1002","open":true}`)
	})

	ctx := context.Background()
	result, err := client.Tickets.CreateRaw(ctx, &CreateTicketRequest{
		TicketTypeID: "tt-1",
		Contacts:     []TicketContactRef{{ID: "c-1"}},
	})
	if err != nil {
		t.Fatalf("CreateRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	data, err := ParseTicketCreateResult(result)
	if err != nil {
		t.Fatalf("ParseTicketCreateResult returned error: %v", err)
	}
	if data.ID != "456" {
		t.Errorf("Data.ID = %v, want 456", data.ID)
	}
}

func TestTicketsService_UpdateRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tickets/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `{"type":"ticket","id":"123","admin_assignee_id":"admin-2"}`)
	})

	ctx := context.Background()
	result, err := client.Tickets.UpdateRaw(ctx, "123", &UpdateTicketRequest{
		AssigneeID: "admin-2",
	})
	if err != nil {
		t.Fatalf("UpdateRaw returned error: %v", err)
	}
	data, err := ParseTicketUpdateResult(result)
	if err != nil {
		t.Fatalf("ParseTicketUpdateResult returned error: %v", err)
	}
	if data.ID != "123" {
		t.Errorf("Data.ID = %v, want 123", data.ID)
	}
	if data.AdminAssigneeID != "admin-2" {
		t.Errorf("Data.AdminAssigneeID = %v, want admin-2", data.AdminAssigneeID)
	}
}

func TestTicketsService_DeleteRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tickets/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{"id":"123","object":"ticket","deleted":true}`)
	})

	ctx := context.Background()
	result, err := client.Tickets.DeleteRaw(ctx, "123")
	if err != nil {
		t.Fatalf("DeleteRaw returned error: %v", err)
	}
	data, err := ParseTicketDeleteResult(result)
	if err != nil {
		t.Fatalf("ParseTicketDeleteResult returned error: %v", err)
	}
	if !data.Deleted {
		t.Error("Data.Deleted = false, want true")
	}
}

func TestTicketsService_SearchRaw(t *testing.T) {
	client, mux, teardown := setup()
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
	result, err := client.Tickets.SearchRaw(ctx, &SearchRequest{
		Query: And(SingleFilterOf("open", "=", "true")),
	})
	if err != nil {
		t.Fatalf("SearchRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	data, err := ParseTicketSearchResult(result)
	if err != nil {
		t.Fatalf("ParseTicketSearchResult returned error: %v", err)
	}
	if len(data.Data) != 2 {
		t.Errorf("Data.Data length = %d, want 2", len(data.Data))
	}
	if data.TotalCount != 2 {
		t.Errorf("Data.TotalCount = %d, want 2", data.TotalCount)
	}
}

func TestTicketsService_ReplyRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tickets/123/reply", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"ticket_part","id":"tp-2","part_type":"comment","body":"<p>Reply</p>"}`)
	})

	ctx := context.Background()
	result, err := client.Tickets.ReplyRaw(ctx, "123", &ReplyTicketRequest{
		MessageType: "comment",
		Type:        "user",
		Body:        "Reply",
	})
	if err != nil {
		t.Fatalf("ReplyRaw returned error: %v", err)
	}
	data, err := ParseTicketReplyResult(result)
	if err != nil {
		t.Fatalf("ParseTicketReplyResult returned error: %v", err)
	}
	if data.ID != "tp-2" {
		t.Errorf("Data.ID = %v, want tp-2", data.ID)
	}
	if data.PartType != "comment" {
		t.Errorf("Data.PartType = %v, want comment", data.PartType)
	}
}

func TestTicketsService_EnqueueRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tickets/enqueue", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"id":"job-1","status":"queued"}`)
	})

	ctx := context.Background()
	result, err := client.Tickets.EnqueueRaw(ctx, &CreateTicketRequest{
		TicketTypeID: "tt-1",
		Contacts:     []TicketContactRef{{ID: "c-1"}},
	})
	if err != nil {
		t.Fatalf("EnqueueRaw returned error: %v", err)
	}
	data, err := ParseTicketEnqueueResult(result)
	if err != nil {
		t.Fatalf("ParseTicketEnqueueResult returned error: %v", err)
	}
	if data.ID != "job-1" {
		t.Errorf("Data.ID = %v, want job-1", data.ID)
	}
	if data.Status != "queued" {
		t.Errorf("Data.Status = %v, want queued", data.Status)
	}
}

func TestTicketsService_AddTagRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tickets/123/tags", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"tag","id":"tag-1","name":"Urgent"}`)
	})

	ctx := context.Background()
	result, err := client.Tickets.AddTagRaw(ctx, "123", "tag-1", "admin-1")
	if err != nil {
		t.Fatalf("AddTagRaw returned error: %v", err)
	}
	data, err := ParseTicketAddTagResult(result)
	if err != nil {
		t.Fatalf("ParseTicketAddTagResult returned error: %v", err)
	}
	if data.ID != "tag-1" {
		t.Errorf("Data.ID = %v, want tag-1", data.ID)
	}
	if data.Name != "Urgent" {
		t.Errorf("Data.Name = %v, want Urgent", data.Name)
	}
}

func TestTicketsService_RemoveTagRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tickets/123/tags/tag-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{"type":"tag","id":"tag-1","name":"Urgent"}`)
	})

	ctx := context.Background()
	result, err := client.Tickets.RemoveTagRaw(ctx, "123", "tag-1", "admin-1")
	if err != nil {
		t.Fatalf("RemoveTagRaw returned error: %v", err)
	}
	data, err := ParseTicketRemoveTagResult(result)
	if err != nil {
		t.Fatalf("ParseTicketRemoveTagResult returned error: %v", err)
	}
	if data.ID != "tag-1" {
		t.Errorf("Data.ID = %v, want tag-1", data.ID)
	}
}

func TestTicketsService_Search_RateLimit(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/tickets/search", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		fmt.Fprint(w, `{"type":"error.list","errors":[{"code":"rate_limit","message":"Rate limit exceeded"}]}`)
	})

	ctx := context.Background()
	_, err := client.Tickets.Search(ctx, &SearchRequest{
		Query: And(SingleFilterOf("open", "=", "true")),
	})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !IsRateLimited(err) {
		t.Errorf("IsRateLimited = false, want true")
	}
}

func TestTicketsService_Create_WithSkipNotifications(t *testing.T) {
	client, mux, teardown := setup()
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
	ticket, err := client.Tickets.Create(ctx, &CreateTicketRequest{
		TicketTypeID:      "tt-1",
		Contacts:          []TicketContactRef{{ID: "c-1"}},
		SkipNotifications: &skipNotif,
	})
	if err != nil {
		t.Fatalf("Tickets.Create returned error: %v", err)
	}
	if ticket.ID != "t-1" {
		t.Errorf("Ticket.ID = %v, want t-1", ticket.ID)
	}
}

func TestTicketsService_Update_WithSkipNotifications(t *testing.T) {
	client, mux, teardown := setup()
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
	ticket, err := client.Tickets.Update(ctx, "t-1", &UpdateTicketRequest{
		TicketStateID:     "ts-2",
		SkipNotifications: &skipNotif,
	})
	if err != nil {
		t.Fatalf("Tickets.Update returned error: %v", err)
	}
	if ticket.ID != "t-1" {
		t.Errorf("Ticket.ID = %v, want t-1", ticket.ID)
	}
}
