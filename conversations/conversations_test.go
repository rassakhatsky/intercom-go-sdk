package conversations_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rassakhatsky/intercom-go-sdk/conversations"
	"github.com/rassakhatsky/intercom-go-sdk/internal/api"
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

func setup() (svc *conversations.Service, mux *http.ServeMux, teardown func()) {
	mux = http.NewServeMux()
	server := httptest.NewServer(mux)
	caller := &testCaller{baseURL: server.URL, client: server.Client()}
	svc = conversations.NewService(caller)
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
		t.Errorf("Header %v = %v, want %v", header, got, want)
	}
}

func TestService_Get(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testHeader(t, r, "Authorization", "Bearer test-token")
		fmt.Fprint(w, `{
			"type":"conversation",
			"id":"123",
			"created_at":1734537460,
			"updated_at":1734537460,
			"state":"closed",
			"open":false,
			"read":false,
			"priority":"not_priority",
			"source":{
				"type":"conversation",
				"id":"403918320",
				"delivered_as":"admin_initiated",
				"subject":"",
				"body":"<p>this is the message body</p>",
				"author":{"type":"admin","id":"991267628","name":"Ciaran Lee","email":"admin@email.com"},
				"attachments":[],
				"redacted":false
			},
			"contacts":{
				"type":"contact.list",
				"contacts":[{"type":"contact","id":"abc123","external_id":"70"}]
			},
			"tags":{
				"type":"tag.list",
				"tags":[{"type":"tag","id":"t1","name":"VIP"}]
			},
			"conversation_parts":{
				"type":"conversation_part.list",
				"conversation_parts":[{
					"type":"conversation_part",
					"id":"1",
					"part_type":"comment",
					"body":"<p>Okay!</p>",
					"created_at":1663597223,
					"author":{"type":"admin","id":"274","name":"Operator","email":"operator@intercom.io"},
					"redacted":false
				}],
				"total_count":1
			},
			"custom_attributes":{"issue_type":"Billing"}
		}`)
	})

	ctx := context.Background()
	conv, err := svc.Get(ctx, "123")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if conv.ID != "123" {
		t.Errorf("Conversation.ID = %v, want 123", conv.ID)
	}
	if conv.State != "closed" {
		t.Errorf("Conversation.State = %v, want closed", conv.State)
	}
	if conv.Source == nil || conv.Source.Body != "<p>this is the message body</p>" {
		t.Errorf("Conversation.Source.Body unexpected")
	}
	if conv.Source.Author == nil || conv.Source.Author.Name != "Ciaran Lee" {
		t.Errorf("Conversation.Source.Author.Name unexpected")
	}
	if conv.Contacts == nil || len(conv.Contacts.Contacts) != 1 {
		t.Errorf("Conversation.Contacts count unexpected")
	}
	if conv.Tags == nil || len(conv.Tags.Tags) != 1 {
		t.Errorf("Conversation.Tags count unexpected")
	}
	if conv.ConversationParts == nil || len(conv.ConversationParts.Parts) != 1 {
		t.Errorf("Conversation.ConversationParts count unexpected")
	}
	if conv.ConversationParts.Parts[0].PartType != "comment" {
		t.Errorf("Part.PartType = %v, want comment", conv.ConversationParts.Parts[0].PartType)
	}
	if conv.CustomAttributes["issue_type"] != "Billing" {
		t.Errorf("CustomAttributes[issue_type] = %v, want Billing", conv.CustomAttributes["issue_type"])
	}
}

func TestService_Create(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body conversations.CreateRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.From.Type != "user" {
			t.Errorf("Create body from.type = %v, want user", body.From.Type)
		}
		if body.From.ID != "abc123" {
			t.Errorf("Create body from.id = %v, want abc123", body.From.ID)
		}
		if body.Body != "Hello there" {
			t.Errorf("Create body body = %v, want Hello there", body.Body)
		}
		fmt.Fprint(w, `{"type":"user_message","id":"403918330","created_at":1734537501,"body":"Hello there","message_type":"inapp","conversation_id":"499"}`)
	})

	ctx := context.Background()
	msg, err := svc.Create(ctx, &conversations.CreateRequest{
		From: conversations.From{Type: "user", ID: "abc123"},
		Body: "Hello there",
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if msg.ConversationID != "499" {
		t.Errorf("Message.ConversationID = %v, want 499", msg.ConversationID)
	}
}

func TestService_Update(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		var body conversations.UpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.Title != "new title" {
			t.Errorf("Update body title = %v, want new title", body.Title)
		}
		fmt.Fprint(w, `{"type":"conversation","id":"123","state":"closed","title":"new title","custom_attributes":{"issue_type":"Billing"}}`)
	})

	ctx := context.Background()
	read := true
	conv, err := svc.Update(ctx, "123", &conversations.UpdateRequest{
		Read:  &read,
		Title: "new title",
		CustomAttributes: map[string]any{
			"issue_type": "Billing",
		},
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if conv.ID != "123" {
		t.Errorf("Conversation.ID = %v, want 123", conv.ID)
	}
}

func TestService_Delete(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{"id":"123","object":"conversation","deleted":true}`)
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

	mux.HandleFunc("/conversations", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"conversation.list",
			"conversations":[
				{"type":"conversation","id":"1","state":"open"},
				{"type":"conversation","id":"2","state":"closed"}
			],
			"total_count":2,
			"pages":{"type":"pages","page":1,"per_page":20,"total_pages":1}
		}`)
	})

	ctx := context.Background()
	result, err := svc.List(ctx, nil)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(result.Data) != 2 {
		t.Errorf("List returned %d conversations, want 2", len(result.Data))
	}
	if result.TotalCount != 2 {
		t.Errorf("TotalCount = %d, want 2", result.TotalCount)
	}
}

func TestService_ListAll(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	callCount := 0
	mux.HandleFunc("/conversations", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		callCount++
		if callCount == 1 {
			fmt.Fprint(w, `{
				"type":"conversation.list",
				"conversations":[{"type":"conversation","id":"1"}],
				"total_count":2,
				"pages":{"type":"pages","page":1,"per_page":1,"total_pages":2,"next":{"per_page":1,"starting_after":"cursor1"}}
			}`)
		} else {
			fmt.Fprint(w, `{
				"type":"conversation.list",
				"conversations":[{"type":"conversation","id":"2"}],
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

	mux.HandleFunc("/conversations/search", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body api.SearchRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.Query.Operator != "AND" {
			t.Errorf("Search query operator = %v, want AND", body.Query.Operator)
		}
		fmt.Fprint(w, `{
			"type":"conversation.list",
			"conversations":[{"type":"conversation","id":"1","state":"open"}],
			"total_count":1,
			"pages":{"type":"pages","page":1,"per_page":5,"total_pages":1}
		}`)
	})

	ctx := context.Background()
	result, err := svc.Search(ctx, &api.SearchRequest{
		Query:      api.And(api.SingleFilterOf("created_at", api.OpGreaterThan, "1306054154")),
		Pagination: &api.SearchPagination{PerPage: 5},
	})
	if err != nil {
		t.Fatalf("Search returned error: %v", err)
	}
	if len(result.Data) != 1 {
		t.Errorf("Search returned %d conversations, want 1", len(result.Data))
	}
}

func TestService_Reply(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123/reply", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body conversations.ReplyRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.MessageType != "comment" {
			t.Errorf("Reply body message_type = %v, want comment", body.MessageType)
		}
		if body.Body != "Thanks again :)" {
			t.Errorf("Reply body body = %v, want Thanks again :)", body.Body)
		}
		fmt.Fprint(w, `{"type":"conversation","id":"123","state":"open"}`)
	})

	ctx := context.Background()
	conv, err := svc.Reply(ctx, "123", &conversations.ReplyRequest{
		MessageType:    "comment",
		Type:           "user",
		IntercomUserID: "abc123",
		Body:           "Thanks again :)",
	})
	if err != nil {
		t.Fatalf("Reply returned error: %v", err)
	}
	if conv.ID != "123" {
		t.Errorf("Conversation.ID = %v, want 123", conv.ID)
	}
}

func TestService_Reply_AdminNote(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123/reply", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body conversations.ReplyRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.MessageType != "note" {
			t.Errorf("Reply body message_type = %v, want note", body.MessageType)
		}
		if body.AdminID != "admin-1" {
			t.Errorf("Reply body admin_id = %v, want admin-1", body.AdminID)
		}
		fmt.Fprint(w, `{"type":"conversation","id":"123","state":"closed"}`)
	})

	ctx := context.Background()
	conv, err := svc.Reply(ctx, "123", &conversations.ReplyRequest{
		MessageType: "note",
		Type:        "admin",
		AdminID:     "admin-1",
		Body:        "<p>Internal note</p>",
	})
	if err != nil {
		t.Fatalf("Reply returned error: %v", err)
	}
	if conv.ID != "123" {
		t.Errorf("Conversation.ID = %v, want 123", conv.ID)
	}
}

func TestService_Close(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123/parts", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body conversations.ManageRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.MessageType != "close" {
			t.Errorf("ManageParts body message_type = %v, want close", body.MessageType)
		}
		if body.AdminID != "admin-1" {
			t.Errorf("ManageParts body admin_id = %v, want admin-1", body.AdminID)
		}
		fmt.Fprint(w, `{"type":"conversation","id":"123","state":"closed","open":false}`)
	})

	ctx := context.Background()
	conv, err := svc.Close(ctx, "123", &conversations.ManageRequest{
		MessageType: "close",
		Type:        "admin",
		AdminID:     "admin-1",
		Body:        "Goodbye :)",
	})
	if err != nil {
		t.Fatalf("Close returned error: %v", err)
	}
	if conv.State != "closed" {
		t.Errorf("Conversation.State = %v, want closed", conv.State)
	}
}

func TestService_Open(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123/parts", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body conversations.ManageRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.MessageType != "open" {
			t.Errorf("ManageParts body message_type = %v, want open", body.MessageType)
		}
		fmt.Fprint(w, `{"type":"conversation","id":"123","state":"open","open":true}`)
	})

	ctx := context.Background()
	conv, err := svc.Open(ctx, "123", &conversations.ManageRequest{
		MessageType: "open",
		AdminID:     "admin-1",
	})
	if err != nil {
		t.Fatalf("Open returned error: %v", err)
	}
	if conv.State != "open" {
		t.Errorf("Conversation.State = %v, want open", conv.State)
	}
}

func TestService_Snooze(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123/parts", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body conversations.ManageRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.MessageType != "snoozed" {
			t.Errorf("ManageParts body message_type = %v, want snoozed", body.MessageType)
		}
		if body.SnoozedUntil == nil || *body.SnoozedUntil != 1734541187 {
			t.Errorf("ManageParts body snoozed_until = %v, want 1734541187", body.SnoozedUntil)
		}
		fmt.Fprint(w, `{"type":"conversation","id":"123","state":"snoozed","snoozed_until":1734541187}`)
	})

	ctx := context.Background()
	snoozeUntil := int64(1734541187)
	conv, err := svc.Snooze(ctx, "123", &conversations.ManageRequest{
		MessageType:  "snoozed",
		AdminID:      "admin-1",
		SnoozedUntil: &snoozeUntil,
	})
	if err != nil {
		t.Fatalf("Snooze returned error: %v", err)
	}
	if conv.State != "snoozed" {
		t.Errorf("Conversation.State = %v, want snoozed", conv.State)
	}
}

func TestService_Assign(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123/parts", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body conversations.ManageRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.MessageType != "assignment" {
			t.Errorf("ManageParts body message_type = %v, want assignment", body.MessageType)
		}
		if body.AssigneeID != "admin-2" {
			t.Errorf("ManageParts body assignee_id = %v, want admin-2", body.AssigneeID)
		}
		fmt.Fprint(w, `{"type":"conversation","id":"123","admin_assignee_id":"admin-2"}`)
	})

	ctx := context.Background()
	conv, err := svc.Assign(ctx, "123", &conversations.ManageRequest{
		MessageType: "assignment",
		Type:        "admin",
		AdminID:     "admin-1",
		AssigneeID:  "admin-2",
	})
	if err != nil {
		t.Fatalf("Assign returned error: %v", err)
	}
	if conv.AdminAssigneeID != "admin-2" {
		t.Errorf("Conversation.AdminAssigneeID = %v, want admin-2", conv.AdminAssigneeID)
	}
}

func TestService_Convert(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123/convert", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body conversations.ConvertRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.TicketTypeID != "53" {
			t.Errorf("Convert body ticket_type_id = %v, want 53", body.TicketTypeID)
		}
		fmt.Fprint(w, `{"type":"ticket","id":"611","ticket_id":"22"}`)
	})

	ctx := context.Background()
	ticket, err := svc.Convert(ctx, "123", &conversations.ConvertRequest{
		TicketTypeID: "53",
	})
	if err != nil {
		t.Fatalf("Convert returned error: %v", err)
	}
	if ticket.ID != "611" {
		t.Errorf("Ticket.ID = %v, want 611", ticket.ID)
	}
	if ticket.TicketID != "22" {
		t.Errorf("Ticket.TicketID = %v, want 22", ticket.TicketID)
	}
}

func TestService_Redact(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/redact", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body conversations.RedactRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.Type != "conversation_part" {
			t.Errorf("Redact body type = %v, want conversation_part", body.Type)
		}
		if body.ConversationID != "608" {
			t.Errorf("Redact body conversation_id = %v, want 608", body.ConversationID)
		}
		if body.ConversationPartID != "149" {
			t.Errorf("Redact body conversation_part_id = %v, want 149", body.ConversationPartID)
		}
		fmt.Fprint(w, `{"type":"conversation","id":"608","state":"open"}`)
	})

	ctx := context.Background()
	conv, err := svc.Redact(ctx, &conversations.RedactRequest{
		Type:               "conversation_part",
		ConversationID:     "608",
		ConversationPartID: "149",
	})
	if err != nil {
		t.Fatalf("Redact returned error: %v", err)
	}
	if conv.ID != "608" {
		t.Errorf("Conversation.ID = %v, want 608", conv.ID)
	}
}

func TestService_AddCustomer(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123/customers", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body conversations.AttachContactRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.AdminID != "admin-1" {
			t.Errorf("AddCustomer body admin_id = %v, want admin-1", body.AdminID)
		}
		if body.Customer.IntercomUserID != "contact-1" {
			t.Errorf("AddCustomer body customer.intercom_user_id = %v, want contact-1", body.Customer.IntercomUserID)
		}
		fmt.Fprint(w, `{"type":"conversation","id":"123"}`)
	})

	ctx := context.Background()
	conv, err := svc.AddCustomer(ctx, "123", &conversations.AttachContactRequest{
		AdminID: "admin-1",
		Customer: conversations.Customer{
			IntercomUserID: "contact-1",
		},
	})
	if err != nil {
		t.Fatalf("AddCustomer returned error: %v", err)
	}
	if conv.ID != "123" {
		t.Errorf("Conversation.ID = %v, want 123", conv.ID)
	}
}

func TestService_RemoveCustomer(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123/customers/contact-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		var body conversations.DetachContactRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.AdminID != "admin-1" {
			t.Errorf("RemoveCustomer body admin_id = %v, want admin-1", body.AdminID)
		}
		fmt.Fprint(w, `{"type":"conversation","id":"123"}`)
	})

	ctx := context.Background()
	conv, err := svc.RemoveCustomer(ctx, "123", "contact-1", &conversations.DetachContactRequest{
		AdminID: "admin-1",
	})
	if err != nil {
		t.Fatalf("RemoveCustomer returned error: %v", err)
	}
	if conv.ID != "123" {
		t.Errorf("Conversation.ID = %v, want 123", conv.ID)
	}
}

func TestService_AddTag(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123/tags", func(w http.ResponseWriter, r *http.Request) {
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
		fmt.Fprint(w, `{"type":"tag","id":"tag-1","name":"Manual tag"}`)
	})

	ctx := context.Background()
	tag, err := svc.AddTag(ctx, "123", "tag-1", "admin-1")
	if err != nil {
		t.Fatalf("AddTag returned error: %v", err)
	}
	if tag.ID != "tag-1" {
		t.Errorf("Tag.ID = %v, want tag-1", tag.ID)
	}
	if tag.Name != "Manual tag" {
		t.Errorf("Tag.Name = %v, want Manual tag", tag.Name)
	}
}

func TestService_RemoveTag(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123/tags/tag-1", func(w http.ResponseWriter, r *http.Request) {
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
		fmt.Fprint(w, `{"type":"tag","id":"tag-1","name":"Manual tag"}`)
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

func TestService_Get_NotFound(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/999", func(w http.ResponseWriter, r *http.Request) {
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

func TestService_List_RateLimit(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations", func(w http.ResponseWriter, r *http.Request) {
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

	mux.HandleFunc("/conversations/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-conv-1")
		fmt.Fprint(w, `{"type":"conversation","id":"123","state":"open"}`)
	})

	ctx := context.Background()
	result, err := svc.GetRaw(ctx, "123")
	if err != nil {
		t.Fatalf("GetRaw returned error: %v", err)
	}
	conv, err := conversations.ParseGetResult(result)
	if err != nil {
		t.Fatalf("ParseGetResult returned error: %v", err)
	}
	if conv.ID != "123" {
		t.Errorf("conv.ID = %v, want 123", conv.ID)
	}
	if conv.State != "open" {
		t.Errorf("conv.State = %v, want open", conv.State)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-conv-1" {
		t.Errorf("Header X-Request-Id = %q, want req-conv-1", result.Header.Get("X-Request-Id"))
	}
	if result.Error != nil {
		t.Errorf("Result.Error = %v, want nil", result.Error)
	}
}

func TestService_GetRaw_NotFound(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/999", func(w http.ResponseWriter, r *http.Request) {
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

func TestService_ListRaw(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"conversation.list",
			"conversations":[
				{"type":"conversation","id":"1","state":"open"},
				{"type":"conversation","id":"2","state":"closed"}
			],
			"total_count":2,
			"pages":{"type":"pages","page":1,"per_page":20,"total_pages":1}
		}`)
	})

	ctx := context.Background()
	result, err := svc.ListRaw(ctx, nil)
	if err != nil {
		t.Fatalf("ListRaw returned error: %v", err)
	}
	page, err := conversations.ParseListResult(result)
	if err != nil {
		t.Fatalf("ParseListResult returned error: %v", err)
	}
	if len(page.Data) != 2 {
		t.Errorf("Data count = %d, want 2", len(page.Data))
	}
	if page.TotalCount != 2 {
		t.Errorf("TotalCount = %d, want 2", page.TotalCount)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestService_CreateRaw(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"user_message","id":"403918330","created_at":1734537501,"body":"Hello","message_type":"inapp","conversation_id":"499"}`)
	})

	ctx := context.Background()
	result, err := svc.CreateRaw(ctx, &conversations.CreateRequest{
		From: conversations.From{Type: "user", ID: "abc123"},
		Body: "Hello",
	})
	if err != nil {
		t.Fatalf("CreateRaw returned error: %v", err)
	}
	msg, err := conversations.ParseCreateResult(result)
	if err != nil {
		t.Fatalf("ParseCreateResult returned error: %v", err)
	}
	if msg.ConversationID != "499" {
		t.Errorf("msg.ConversationID = %v, want 499", msg.ConversationID)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestService_UpdateRaw(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `{"type":"conversation","id":"123","title":"new title"}`)
	})

	ctx := context.Background()
	result, err := svc.UpdateRaw(ctx, "123", &conversations.UpdateRequest{
		Title: "new title",
	})
	if err != nil {
		t.Fatalf("UpdateRaw returned error: %v", err)
	}
	conv, err := conversations.ParseUpdateResult(result)
	if err != nil {
		t.Fatalf("ParseUpdateResult returned error: %v", err)
	}
	if conv.ID != "123" {
		t.Errorf("conv.ID = %v, want 123", conv.ID)
	}
}

func TestService_DeleteRaw(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{"id":"123","object":"conversation","deleted":true}`)
	})

	ctx := context.Background()
	result, err := svc.DeleteRaw(ctx, "123")
	if err != nil {
		t.Fatalf("DeleteRaw returned error: %v", err)
	}
	deleted, err := conversations.ParseDeleteResult(result)
	if err != nil {
		t.Fatalf("ParseDeleteResult returned error: %v", err)
	}
	if !deleted.Deleted {
		t.Error("deleted.Deleted = false, want true")
	}
	if deleted.ID != "123" {
		t.Errorf("deleted.ID = %v, want 123", deleted.ID)
	}
}

func TestService_SearchRaw(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/search", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{
			"type":"conversation.list",
			"conversations":[{"type":"conversation","id":"1","state":"open"}],
			"total_count":1,
			"pages":{"type":"pages","page":1,"per_page":5,"total_pages":1}
		}`)
	})

	ctx := context.Background()
	result, err := svc.SearchRaw(ctx, &api.SearchRequest{
		Query:      api.And(api.SingleFilterOf("created_at", api.OpGreaterThan, "1306054154")),
		Pagination: &api.SearchPagination{PerPage: 5},
	})
	if err != nil {
		t.Fatalf("SearchRaw returned error: %v", err)
	}
	page, err := conversations.ParseSearchResult(result)
	if err != nil {
		t.Fatalf("ParseSearchResult returned error: %v", err)
	}
	if len(page.Data) != 1 {
		t.Errorf("Data count = %d, want 1", len(page.Data))
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestService_ReplyRaw(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123/reply", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"conversation","id":"123","state":"open"}`)
	})

	ctx := context.Background()
	result, err := svc.ReplyRaw(ctx, "123", &conversations.ReplyRequest{
		MessageType:    "comment",
		Type:           "user",
		IntercomUserID: "abc123",
		Body:           "Thanks again :)",
	})
	if err != nil {
		t.Fatalf("ReplyRaw returned error: %v", err)
	}
	conv, err := conversations.ParseReplyResult(result)
	if err != nil {
		t.Fatalf("ParseReplyResult returned error: %v", err)
	}
	if conv.ID != "123" {
		t.Errorf("conv.ID = %v, want 123", conv.ID)
	}
}

func TestService_CloseRaw(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123/parts", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"conversation","id":"123","state":"closed","open":false}`)
	})

	ctx := context.Background()
	result, err := svc.CloseRaw(ctx, "123", &conversations.ManageRequest{
		MessageType: "close",
		Type:        "admin",
		AdminID:     "admin-1",
	})
	if err != nil {
		t.Fatalf("CloseRaw returned error: %v", err)
	}
	conv, err := conversations.ParseCloseResult(result)
	if err != nil {
		t.Fatalf("ParseCloseResult returned error: %v", err)
	}
	if conv.State != "closed" {
		t.Errorf("conv.State = %v, want closed", conv.State)
	}
}

func TestService_OpenRaw(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123/parts", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"conversation","id":"123","state":"open","open":true}`)
	})

	ctx := context.Background()
	result, err := svc.OpenRaw(ctx, "123", &conversations.ManageRequest{
		MessageType: "open",
		AdminID:     "admin-1",
	})
	if err != nil {
		t.Fatalf("OpenRaw returned error: %v", err)
	}
	conv, err := conversations.ParseOpenResult(result)
	if err != nil {
		t.Fatalf("ParseOpenResult returned error: %v", err)
	}
	if conv.State != "open" {
		t.Errorf("conv.State = %v, want open", conv.State)
	}
}

func TestService_SnoozeRaw(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123/parts", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"conversation","id":"123","state":"snoozed","snoozed_until":1734541187}`)
	})

	ctx := context.Background()
	snoozeUntil := int64(1734541187)
	result, err := svc.SnoozeRaw(ctx, "123", &conversations.ManageRequest{
		MessageType:  "snoozed",
		AdminID:      "admin-1",
		SnoozedUntil: &snoozeUntil,
	})
	if err != nil {
		t.Fatalf("SnoozeRaw returned error: %v", err)
	}
	conv, err := conversations.ParseSnoozeResult(result)
	if err != nil {
		t.Fatalf("ParseSnoozeResult returned error: %v", err)
	}
	if conv.State != "snoozed" {
		t.Errorf("conv.State = %v, want snoozed", conv.State)
	}
}

func TestService_AssignRaw(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123/parts", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"conversation","id":"123","admin_assignee_id":"admin-2"}`)
	})

	ctx := context.Background()
	result, err := svc.AssignRaw(ctx, "123", &conversations.ManageRequest{
		MessageType: "assignment",
		Type:        "admin",
		AdminID:     "admin-1",
		AssigneeID:  "admin-2",
	})
	if err != nil {
		t.Fatalf("AssignRaw returned error: %v", err)
	}
	conv, err := conversations.ParseAssignResult(result)
	if err != nil {
		t.Fatalf("ParseAssignResult returned error: %v", err)
	}
	if conv.AdminAssigneeID != "admin-2" {
		t.Errorf("conv.AdminAssigneeID = %v, want admin-2", conv.AdminAssigneeID)
	}
}

func TestService_ConvertRaw(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123/convert", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"ticket","id":"611","ticket_id":"22"}`)
	})

	ctx := context.Background()
	result, err := svc.ConvertRaw(ctx, "123", &conversations.ConvertRequest{
		TicketTypeID: "53",
	})
	if err != nil {
		t.Fatalf("ConvertRaw returned error: %v", err)
	}
	ticket, err := conversations.ParseConvertResult(result)
	if err != nil {
		t.Fatalf("ParseConvertResult returned error: %v", err)
	}
	if ticket.ID != "611" {
		t.Errorf("ticket.ID = %v, want 611", ticket.ID)
	}
	if ticket.TicketID != "22" {
		t.Errorf("ticket.TicketID = %v, want 22", ticket.TicketID)
	}
}

func TestService_RedactRaw(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/redact", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"conversation","id":"608","state":"open"}`)
	})

	ctx := context.Background()
	result, err := svc.RedactRaw(ctx, &conversations.RedactRequest{
		Type:               "conversation_part",
		ConversationID:     "608",
		ConversationPartID: "149",
	})
	if err != nil {
		t.Fatalf("RedactRaw returned error: %v", err)
	}
	conv, err := conversations.ParseRedactResult(result)
	if err != nil {
		t.Fatalf("ParseRedactResult returned error: %v", err)
	}
	if conv.ID != "608" {
		t.Errorf("conv.ID = %v, want 608", conv.ID)
	}
}

func TestService_AddCustomerRaw(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123/customers", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"conversation","id":"123"}`)
	})

	ctx := context.Background()
	result, err := svc.AddCustomerRaw(ctx, "123", &conversations.AttachContactRequest{
		AdminID: "admin-1",
		Customer: conversations.Customer{
			IntercomUserID: "contact-1",
		},
	})
	if err != nil {
		t.Fatalf("AddCustomerRaw returned error: %v", err)
	}
	conv, err := conversations.ParseAddCustomerResult(result)
	if err != nil {
		t.Fatalf("ParseAddCustomerResult returned error: %v", err)
	}
	if conv.ID != "123" {
		t.Errorf("conv.ID = %v, want 123", conv.ID)
	}
}

func TestService_RemoveCustomerRaw(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123/customers/contact-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{"type":"conversation","id":"123"}`)
	})

	ctx := context.Background()
	result, err := svc.RemoveCustomerRaw(ctx, "123", "contact-1", &conversations.DetachContactRequest{
		AdminID: "admin-1",
	})
	if err != nil {
		t.Fatalf("RemoveCustomerRaw returned error: %v", err)
	}
	conv, err := conversations.ParseRemoveCustomerResult(result)
	if err != nil {
		t.Fatalf("ParseRemoveCustomerResult returned error: %v", err)
	}
	if conv.ID != "123" {
		t.Errorf("conv.ID = %v, want 123", conv.ID)
	}
}

func TestService_AddTagRaw(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123/tags", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"tag","id":"tag-1","name":"Manual tag"}`)
	})

	ctx := context.Background()
	result, err := svc.AddTagRaw(ctx, "123", "tag-1", "admin-1")
	if err != nil {
		t.Fatalf("AddTagRaw returned error: %v", err)
	}
	tag, err := conversations.ParseAddTagResult(result)
	if err != nil {
		t.Fatalf("ParseAddTagResult returned error: %v", err)
	}
	if tag.ID != "tag-1" {
		t.Errorf("tag.ID = %v, want tag-1", tag.ID)
	}
	if tag.Name != "Manual tag" {
		t.Errorf("tag.Name = %v, want Manual tag", tag.Name)
	}
}

func TestService_RemoveTagRaw(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123/tags/tag-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{"type":"tag","id":"tag-1","name":"Manual tag"}`)
	})

	ctx := context.Background()
	result, err := svc.RemoveTagRaw(ctx, "123", "tag-1", "admin-1")
	if err != nil {
		t.Fatalf("RemoveTagRaw returned error: %v", err)
	}
	tag, err := conversations.ParseRemoveTagResult(result)
	if err != nil {
		t.Fatalf("ParseRemoveTagResult returned error: %v", err)
	}
	if tag.ID != "tag-1" {
		t.Errorf("tag.ID = %v, want tag-1", tag.ID)
	}
}

func TestService_Reply_WithQuickReplyOptions(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123/reply", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["message_type"] != "quick_reply" {
			t.Errorf("message_type = %v, want quick_reply", body["message_type"])
		}
		opts, ok := body["reply_options"].([]any)
		if !ok || len(opts) != 2 {
			t.Fatalf("reply_options length = %v, want 2", len(opts))
		}
		opt0 := opts[0].(map[string]any)
		if opt0["text"] != "Yes" {
			t.Errorf("reply_options[0].text = %v, want Yes", opt0["text"])
		}
		fmt.Fprint(w, `{"type":"conversation","id":"123","state":"open"}`)
	})

	ctx := context.Background()
	conv, err := svc.Reply(ctx, "123", &conversations.ReplyRequest{
		MessageType: "quick_reply",
		Type:        "admin",
		AdminID:     "admin-1",
		ReplyOptions: []conversations.QuickReplyOption{
			{Text: "Yes", UUID: "uuid-1"},
			{Text: "No", UUID: "uuid-2"},
		},
	})
	if err != nil {
		t.Fatalf("Reply returned error: %v", err)
	}
	if conv.ID != "123" {
		t.Errorf("Conversation.ID = %v, want 123", conv.ID)
	}
}

func TestService_Reply_WithAttachmentFiles(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123/reply", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		files, ok := body["attachment_files"].([]any)
		if !ok || len(files) != 1 {
			t.Fatalf("attachment_files length = %v, want 1", len(files))
		}
		file0 := files[0].(map[string]any)
		if file0["content_type"] != "application/pdf" {
			t.Errorf("attachment_files[0].content_type = %v, want application/pdf", file0["content_type"])
		}
		skipNotif, ok := body["skip_notifications"]
		if !ok || skipNotif != true {
			t.Errorf("skip_notifications = %v, want true", skipNotif)
		}
		fmt.Fprint(w, `{"type":"conversation","id":"123","state":"open"}`)
	})

	ctx := context.Background()
	skipNotif := true
	conv, err := svc.Reply(ctx, "123", &conversations.ReplyRequest{
		MessageType: "comment",
		Type:        "admin",
		AdminID:     "admin-1",
		Body:        "See attachment",
		AttachmentFiles: []conversations.AttachmentFile{
			{ContentType: "application/pdf", Data: "base64data", Name: "doc.pdf"},
		},
		SkipNotifications: &skipNotif,
	})
	if err != nil {
		t.Fatalf("Reply returned error: %v", err)
	}
	if conv.ID != "123" {
		t.Errorf("Conversation.ID = %v, want 123", conv.ID)
	}
}
