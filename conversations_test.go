package intercom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestConversationsService_Get(t *testing.T) {
	client, mux, teardown := setup()
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
	conv, err := client.Conversations.Get(ctx, "123")
	if err != nil {
		t.Fatalf("Conversations.Get returned error: %v", err)
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

func TestConversationsService_Create(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body CreateConversationRequest
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
	msg, err := client.Conversations.Create(ctx, &CreateConversationRequest{
		From: ConversationFrom{Type: "user", ID: "abc123"},
		Body: "Hello there",
	})
	if err != nil {
		t.Fatalf("Conversations.Create returned error: %v", err)
	}
	if msg.ConversationID != "499" {
		t.Errorf("Message.ConversationID = %v, want 499", msg.ConversationID)
	}
}

func TestConversationsService_Update(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		var body UpdateConversationRequest
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
	conv, err := client.Conversations.Update(ctx, "123", &UpdateConversationRequest{
		Read:  &read,
		Title: "new title",
		CustomAttributes: map[string]any{
			"issue_type": "Billing",
		},
	})
	if err != nil {
		t.Fatalf("Conversations.Update returned error: %v", err)
	}
	if conv.ID != "123" {
		t.Errorf("Conversation.ID = %v, want 123", conv.ID)
	}
}

func TestConversationsService_Delete(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{"id":"123","object":"conversation","deleted":true}`)
	})

	ctx := context.Background()
	deleted, err := client.Conversations.Delete(ctx, "123")
	if err != nil {
		t.Fatalf("Conversations.Delete returned error: %v", err)
	}
	if !deleted.Deleted {
		t.Error("ConversationDeleted.Deleted = false, want true")
	}
	if deleted.ID != "123" {
		t.Errorf("ConversationDeleted.ID = %v, want 123", deleted.ID)
	}
}

func TestConversationsService_List(t *testing.T) {
	client, mux, teardown := setup()
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
	result, err := client.Conversations.List(ctx, nil)
	if err != nil {
		t.Fatalf("Conversations.List returned error: %v", err)
	}
	if len(result.Data) != 2 {
		t.Errorf("List returned %d conversations, want 2", len(result.Data))
	}
	if result.TotalCount != 2 {
		t.Errorf("TotalCount = %d, want 2", result.TotalCount)
	}
}

func TestConversationsService_ListAll(t *testing.T) {
	client, mux, teardown := setup()
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
	iter := client.Conversations.ListAll(ctx, &ListOptions{PerPage: 1})
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

func TestConversationsService_Search(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/search", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body SearchRequest
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
	result, err := client.Conversations.Search(ctx, &SearchRequest{
		Query:      And(SingleFilterOf("created_at", ">", "1306054154")),
		Pagination: &SearchPagination{PerPage: 5},
	})
	if err != nil {
		t.Fatalf("Conversations.Search returned error: %v", err)
	}
	if len(result.Data) != 1 {
		t.Errorf("Search returned %d conversations, want 1", len(result.Data))
	}
}

func TestConversationsService_Reply(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123/reply", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body ReplyConversationRequest
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
	conv, err := client.Conversations.Reply(ctx, "123", &ReplyConversationRequest{
		MessageType:    "comment",
		Type:           "user",
		IntercomUserID: "abc123",
		Body:           "Thanks again :)",
	})
	if err != nil {
		t.Fatalf("Conversations.Reply returned error: %v", err)
	}
	if conv.ID != "123" {
		t.Errorf("Conversation.ID = %v, want 123", conv.ID)
	}
}

func TestConversationsService_Reply_AdminNote(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123/reply", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body ReplyConversationRequest
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
	conv, err := client.Conversations.Reply(ctx, "123", &ReplyConversationRequest{
		MessageType: "note",
		Type:        "admin",
		AdminID:     "admin-1",
		Body:        "<p>Internal note</p>",
	})
	if err != nil {
		t.Fatalf("Conversations.Reply returned error: %v", err)
	}
	if conv.ID != "123" {
		t.Errorf("Conversation.ID = %v, want 123", conv.ID)
	}
}

func TestConversationsService_Close(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123/parts", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body ManageConversationRequest
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
	conv, err := client.Conversations.Close(ctx, "123", &ManageConversationRequest{
		MessageType: "close",
		Type:        "admin",
		AdminID:     "admin-1",
		Body:        "Goodbye :)",
	})
	if err != nil {
		t.Fatalf("Conversations.Close returned error: %v", err)
	}
	if conv.State != "closed" {
		t.Errorf("Conversation.State = %v, want closed", conv.State)
	}
}

func TestConversationsService_Open(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123/parts", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body ManageConversationRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.MessageType != "open" {
			t.Errorf("ManageParts body message_type = %v, want open", body.MessageType)
		}
		fmt.Fprint(w, `{"type":"conversation","id":"123","state":"open","open":true}`)
	})

	ctx := context.Background()
	conv, err := client.Conversations.Open(ctx, "123", &ManageConversationRequest{
		MessageType: "open",
		AdminID:     "admin-1",
	})
	if err != nil {
		t.Fatalf("Conversations.Open returned error: %v", err)
	}
	if conv.State != "open" {
		t.Errorf("Conversation.State = %v, want open", conv.State)
	}
}

func TestConversationsService_Snooze(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123/parts", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body ManageConversationRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.MessageType != "snoozed" {
			t.Errorf("ManageParts body message_type = %v, want snoozed", body.MessageType)
		}
		if body.SnoozedUntil != 1734541187 {
			t.Errorf("ManageParts body snoozed_until = %v, want 1734541187", body.SnoozedUntil)
		}
		fmt.Fprint(w, `{"type":"conversation","id":"123","state":"snoozed","snoozed_until":1734541187}`)
	})

	ctx := context.Background()
	conv, err := client.Conversations.Snooze(ctx, "123", &ManageConversationRequest{
		MessageType:  "snoozed",
		AdminID:      "admin-1",
		SnoozedUntil: 1734541187,
	})
	if err != nil {
		t.Fatalf("Conversations.Snooze returned error: %v", err)
	}
	if conv.State != "snoozed" {
		t.Errorf("Conversation.State = %v, want snoozed", conv.State)
	}
}

func TestConversationsService_Assign(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123/parts", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body ManageConversationRequest
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
	conv, err := client.Conversations.Assign(ctx, "123", &ManageConversationRequest{
		MessageType: "assignment",
		Type:        "admin",
		AdminID:     "admin-1",
		AssigneeID:  "admin-2",
	})
	if err != nil {
		t.Fatalf("Conversations.Assign returned error: %v", err)
	}
	if conv.AdminAssigneeID != "admin-2" {
		t.Errorf("Conversation.AdminAssigneeID = %v, want admin-2", conv.AdminAssigneeID)
	}
}

func TestConversationsService_Convert(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123/convert", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body ConvertConversationRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.TicketTypeID != "53" {
			t.Errorf("Convert body ticket_type_id = %v, want 53", body.TicketTypeID)
		}
		fmt.Fprint(w, `{"type":"ticket","id":"611","ticket_id":"22"}`)
	})

	ctx := context.Background()
	ticket, err := client.Conversations.Convert(ctx, "123", &ConvertConversationRequest{
		TicketTypeID: "53",
	})
	if err != nil {
		t.Fatalf("Conversations.Convert returned error: %v", err)
	}
	if ticket.ID != "611" {
		t.Errorf("Ticket.ID = %v, want 611", ticket.ID)
	}
	if ticket.TicketID != "22" {
		t.Errorf("Ticket.TicketID = %v, want 22", ticket.TicketID)
	}
}

func TestConversationsService_Redact(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/redact", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body RedactConversationRequest
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
	conv, err := client.Conversations.Redact(ctx, &RedactConversationRequest{
		Type:               "conversation_part",
		ConversationID:     "608",
		ConversationPartID: "149",
	})
	if err != nil {
		t.Fatalf("Conversations.Redact returned error: %v", err)
	}
	if conv.ID != "608" {
		t.Errorf("Conversation.ID = %v, want 608", conv.ID)
	}
}

func TestConversationsService_AddCustomer(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123/customers", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body AttachContactToConversationRequest
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
	conv, err := client.Conversations.AddCustomer(ctx, "123", &AttachContactToConversationRequest{
		AdminID: "admin-1",
		Customer: ConversationCustomer{
			IntercomUserID: "contact-1",
		},
	})
	if err != nil {
		t.Fatalf("Conversations.AddCustomer returned error: %v", err)
	}
	if conv.ID != "123" {
		t.Errorf("Conversation.ID = %v, want 123", conv.ID)
	}
}

func TestConversationsService_RemoveCustomer(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123/customers/contact-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		var body DetachContactFromConversationRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.AdminID != "admin-1" {
			t.Errorf("RemoveCustomer body admin_id = %v, want admin-1", body.AdminID)
		}
		fmt.Fprint(w, `{"type":"conversation","id":"123"}`)
	})

	ctx := context.Background()
	conv, err := client.Conversations.RemoveCustomer(ctx, "123", "contact-1", &DetachContactFromConversationRequest{
		AdminID: "admin-1",
	})
	if err != nil {
		t.Fatalf("Conversations.RemoveCustomer returned error: %v", err)
	}
	if conv.ID != "123" {
		t.Errorf("Conversation.ID = %v, want 123", conv.ID)
	}
}

func TestConversationsService_AddTag(t *testing.T) {
	client, mux, teardown := setup()
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
	tag, err := client.Conversations.AddTag(ctx, "123", "tag-1", "admin-1")
	if err != nil {
		t.Fatalf("Conversations.AddTag returned error: %v", err)
	}
	if tag.ID != "tag-1" {
		t.Errorf("Tag.ID = %v, want tag-1", tag.ID)
	}
	if tag.Name != "Manual tag" {
		t.Errorf("Tag.Name = %v, want Manual tag", tag.Name)
	}
}

func TestConversationsService_RemoveTag(t *testing.T) {
	client, mux, teardown := setup()
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
	tag, err := client.Conversations.RemoveTag(ctx, "123", "tag-1", "admin-1")
	if err != nil {
		t.Fatalf("Conversations.RemoveTag returned error: %v", err)
	}
	if tag.ID != "tag-1" {
		t.Errorf("Tag.ID = %v, want tag-1", tag.ID)
	}
}

func TestConversationsService_Get_NotFound(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/999", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"Resource Not Found"}]}`)
	})

	ctx := context.Background()
	_, err := client.Conversations.Get(ctx, "999")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("IsNotFound = false, want true")
	}
}

func TestConversationsService_List_RateLimit(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		fmt.Fprint(w, `{"type":"error.list","errors":[{"code":"rate_limit","message":"Rate limit exceeded"}]}`)
	})

	ctx := context.Background()
	_, err := client.Conversations.List(ctx, nil)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !IsRateLimited(err) {
		t.Errorf("IsRateLimited = false, want true")
	}
}

// --- Raw method tests ---

func TestConversationsService_GetRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-conv-1")
		fmt.Fprint(w, `{"type":"conversation","id":"123","state":"open"}`)
	})

	ctx := context.Background()
	result, err := client.Conversations.GetRaw(ctx, "123")
	if err != nil {
		t.Fatalf("GetRaw returned error: %v", err)
	}
	conv, err := ParseConversationGetResult(result)
	if err != nil {
		t.Fatalf("ParseConversationGetResult returned error: %v", err)
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

func TestConversationsService_GetRaw_NotFound(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/999", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"Resource Not Found"}]}`)
	})

	ctx := context.Background()
	result, err := client.Conversations.GetRaw(ctx, "999")
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

func TestConversationsService_ListRaw(t *testing.T) {
	client, mux, teardown := setup()
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
	result, err := client.Conversations.ListRaw(ctx, nil)
	if err != nil {
		t.Fatalf("ListRaw returned error: %v", err)
	}
	page, err := ParseConversationListResult(result)
	if err != nil {
		t.Fatalf("ParseConversationListResult returned error: %v", err)
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

func TestConversationsService_CreateRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"user_message","id":"403918330","created_at":1734537501,"body":"Hello","message_type":"inapp","conversation_id":"499"}`)
	})

	ctx := context.Background()
	result, err := client.Conversations.CreateRaw(ctx, &CreateConversationRequest{
		From: ConversationFrom{Type: "user", ID: "abc123"},
		Body: "Hello",
	})
	if err != nil {
		t.Fatalf("CreateRaw returned error: %v", err)
	}
	msg, err := ParseConversationCreateResult(result)
	if err != nil {
		t.Fatalf("ParseConversationCreateResult returned error: %v", err)
	}
	if msg.ConversationID != "499" {
		t.Errorf("msg.ConversationID = %v, want 499", msg.ConversationID)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestConversationsService_UpdateRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `{"type":"conversation","id":"123","title":"new title"}`)
	})

	ctx := context.Background()
	result, err := client.Conversations.UpdateRaw(ctx, "123", &UpdateConversationRequest{
		Title: "new title",
	})
	if err != nil {
		t.Fatalf("UpdateRaw returned error: %v", err)
	}
	conv, err := ParseConversationUpdateResult(result)
	if err != nil {
		t.Fatalf("ParseConversationUpdateResult returned error: %v", err)
	}
	if conv.ID != "123" {
		t.Errorf("conv.ID = %v, want 123", conv.ID)
	}
}

func TestConversationsService_DeleteRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{"id":"123","object":"conversation","deleted":true}`)
	})

	ctx := context.Background()
	result, err := client.Conversations.DeleteRaw(ctx, "123")
	if err != nil {
		t.Fatalf("DeleteRaw returned error: %v", err)
	}
	deleted, err := ParseConversationDeleteResult(result)
	if err != nil {
		t.Fatalf("ParseConversationDeleteResult returned error: %v", err)
	}
	if !deleted.Deleted {
		t.Error("deleted.Deleted = false, want true")
	}
	if deleted.ID != "123" {
		t.Errorf("deleted.ID = %v, want 123", deleted.ID)
	}
}

func TestConversationsService_SearchRaw(t *testing.T) {
	client, mux, teardown := setup()
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
	result, err := client.Conversations.SearchRaw(ctx, &SearchRequest{
		Query:      And(SingleFilterOf("created_at", ">", "1306054154")),
		Pagination: &SearchPagination{PerPage: 5},
	})
	if err != nil {
		t.Fatalf("SearchRaw returned error: %v", err)
	}
	page, err := ParseConversationSearchResult(result)
	if err != nil {
		t.Fatalf("ParseConversationSearchResult returned error: %v", err)
	}
	if len(page.Data) != 1 {
		t.Errorf("Data count = %d, want 1", len(page.Data))
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestConversationsService_ReplyRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123/reply", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"conversation","id":"123","state":"open"}`)
	})

	ctx := context.Background()
	result, err := client.Conversations.ReplyRaw(ctx, "123", &ReplyConversationRequest{
		MessageType:    "comment",
		Type:           "user",
		IntercomUserID: "abc123",
		Body:           "Thanks again :)",
	})
	if err != nil {
		t.Fatalf("ReplyRaw returned error: %v", err)
	}
	conv, err := ParseConversationReplyResult(result)
	if err != nil {
		t.Fatalf("ParseConversationReplyResult returned error: %v", err)
	}
	if conv.ID != "123" {
		t.Errorf("conv.ID = %v, want 123", conv.ID)
	}
}

func TestConversationsService_CloseRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123/parts", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"conversation","id":"123","state":"closed","open":false}`)
	})

	ctx := context.Background()
	result, err := client.Conversations.CloseRaw(ctx, "123", &ManageConversationRequest{
		MessageType: "close",
		Type:        "admin",
		AdminID:     "admin-1",
	})
	if err != nil {
		t.Fatalf("CloseRaw returned error: %v", err)
	}
	conv, err := ParseConversationCloseResult(result)
	if err != nil {
		t.Fatalf("ParseConversationCloseResult returned error: %v", err)
	}
	if conv.State != "closed" {
		t.Errorf("conv.State = %v, want closed", conv.State)
	}
}

func TestConversationsService_OpenRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123/parts", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"conversation","id":"123","state":"open","open":true}`)
	})

	ctx := context.Background()
	result, err := client.Conversations.OpenRaw(ctx, "123", &ManageConversationRequest{
		MessageType: "open",
		AdminID:     "admin-1",
	})
	if err != nil {
		t.Fatalf("OpenRaw returned error: %v", err)
	}
	conv, err := ParseConversationOpenResult(result)
	if err != nil {
		t.Fatalf("ParseConversationOpenResult returned error: %v", err)
	}
	if conv.State != "open" {
		t.Errorf("conv.State = %v, want open", conv.State)
	}
}

func TestConversationsService_SnoozeRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123/parts", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"conversation","id":"123","state":"snoozed","snoozed_until":1734541187}`)
	})

	ctx := context.Background()
	result, err := client.Conversations.SnoozeRaw(ctx, "123", &ManageConversationRequest{
		MessageType:  "snoozed",
		AdminID:      "admin-1",
		SnoozedUntil: 1734541187,
	})
	if err != nil {
		t.Fatalf("SnoozeRaw returned error: %v", err)
	}
	conv, err := ParseConversationSnoozeResult(result)
	if err != nil {
		t.Fatalf("ParseConversationSnoozeResult returned error: %v", err)
	}
	if conv.State != "snoozed" {
		t.Errorf("conv.State = %v, want snoozed", conv.State)
	}
}

func TestConversationsService_AssignRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123/parts", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"conversation","id":"123","admin_assignee_id":"admin-2"}`)
	})

	ctx := context.Background()
	result, err := client.Conversations.AssignRaw(ctx, "123", &ManageConversationRequest{
		MessageType: "assignment",
		Type:        "admin",
		AdminID:     "admin-1",
		AssigneeID:  "admin-2",
	})
	if err != nil {
		t.Fatalf("AssignRaw returned error: %v", err)
	}
	conv, err := ParseConversationAssignResult(result)
	if err != nil {
		t.Fatalf("ParseConversationAssignResult returned error: %v", err)
	}
	if conv.AdminAssigneeID != "admin-2" {
		t.Errorf("conv.AdminAssigneeID = %v, want admin-2", conv.AdminAssigneeID)
	}
}

func TestConversationsService_ConvertRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123/convert", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"ticket","id":"611","ticket_id":"22"}`)
	})

	ctx := context.Background()
	result, err := client.Conversations.ConvertRaw(ctx, "123", &ConvertConversationRequest{
		TicketTypeID: "53",
	})
	if err != nil {
		t.Fatalf("ConvertRaw returned error: %v", err)
	}
	ticket, err := ParseConversationConvertResult(result)
	if err != nil {
		t.Fatalf("ParseConversationConvertResult returned error: %v", err)
	}
	if ticket.ID != "611" {
		t.Errorf("ticket.ID = %v, want 611", ticket.ID)
	}
	if ticket.TicketID != "22" {
		t.Errorf("ticket.TicketID = %v, want 22", ticket.TicketID)
	}
}

func TestConversationsService_RedactRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/redact", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"conversation","id":"608","state":"open"}`)
	})

	ctx := context.Background()
	result, err := client.Conversations.RedactRaw(ctx, &RedactConversationRequest{
		Type:               "conversation_part",
		ConversationID:     "608",
		ConversationPartID: "149",
	})
	if err != nil {
		t.Fatalf("RedactRaw returned error: %v", err)
	}
	conv, err := ParseConversationRedactResult(result)
	if err != nil {
		t.Fatalf("ParseConversationRedactResult returned error: %v", err)
	}
	if conv.ID != "608" {
		t.Errorf("conv.ID = %v, want 608", conv.ID)
	}
}

func TestConversationsService_AddCustomerRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123/customers", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"conversation","id":"123"}`)
	})

	ctx := context.Background()
	result, err := client.Conversations.AddCustomerRaw(ctx, "123", &AttachContactToConversationRequest{
		AdminID: "admin-1",
		Customer: ConversationCustomer{
			IntercomUserID: "contact-1",
		},
	})
	if err != nil {
		t.Fatalf("AddCustomerRaw returned error: %v", err)
	}
	conv, err := ParseConversationAddCustomerResult(result)
	if err != nil {
		t.Fatalf("ParseConversationAddCustomerResult returned error: %v", err)
	}
	if conv.ID != "123" {
		t.Errorf("conv.ID = %v, want 123", conv.ID)
	}
}

func TestConversationsService_RemoveCustomerRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123/customers/contact-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{"type":"conversation","id":"123"}`)
	})

	ctx := context.Background()
	result, err := client.Conversations.RemoveCustomerRaw(ctx, "123", "contact-1", &DetachContactFromConversationRequest{
		AdminID: "admin-1",
	})
	if err != nil {
		t.Fatalf("RemoveCustomerRaw returned error: %v", err)
	}
	conv, err := ParseConversationRemoveCustomerResult(result)
	if err != nil {
		t.Fatalf("ParseConversationRemoveCustomerResult returned error: %v", err)
	}
	if conv.ID != "123" {
		t.Errorf("conv.ID = %v, want 123", conv.ID)
	}
}

func TestConversationsService_AddTagRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123/tags", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"tag","id":"tag-1","name":"Manual tag"}`)
	})

	ctx := context.Background()
	result, err := client.Conversations.AddTagRaw(ctx, "123", "tag-1", "admin-1")
	if err != nil {
		t.Fatalf("AddTagRaw returned error: %v", err)
	}
	tag, err := ParseConversationAddTagResult(result)
	if err != nil {
		t.Fatalf("ParseConversationAddTagResult returned error: %v", err)
	}
	if tag.ID != "tag-1" {
		t.Errorf("tag.ID = %v, want tag-1", tag.ID)
	}
	if tag.Name != "Manual tag" {
		t.Errorf("tag.Name = %v, want Manual tag", tag.Name)
	}
}

func TestConversationsService_RemoveTagRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/conversations/123/tags/tag-1", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{"type":"tag","id":"tag-1","name":"Manual tag"}`)
	})

	ctx := context.Background()
	result, err := client.Conversations.RemoveTagRaw(ctx, "123", "tag-1", "admin-1")
	if err != nil {
		t.Fatalf("RemoveTagRaw returned error: %v", err)
	}
	tag, err := ParseConversationRemoveTagResult(result)
	if err != nil {
		t.Fatalf("ParseConversationRemoveTagResult returned error: %v", err)
	}
	if tag.ID != "tag-1" {
		t.Errorf("tag.ID = %v, want tag-1", tag.ID)
	}
}

