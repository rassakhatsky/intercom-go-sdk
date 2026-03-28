package intercom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestMessagesService_Create_InApp(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		testHeader(t, r, "Authorization", "Bearer test-token")
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["message_type"] != "in_app" {
			t.Errorf("message_type = %v, want in_app", body["message_type"])
		}
		if body["body"] != "Hey there!" {
			t.Errorf("body = %v, want Hey there!", body["body"])
		}
		from, ok := body["from"].(map[string]any)
		if !ok {
			t.Fatal("from is not an object")
		}
		if from["type"] != "admin" {
			t.Errorf("from.type = %v, want admin", from["type"])
		}
		to, ok := body["to"].(map[string]any)
		if !ok {
			t.Fatal("to is not an object")
		}
		if to["type"] != "user" {
			t.Errorf("to.type = %v, want user", to["type"])
		}
		if to["id"] != "536e564f316c83104c000020" {
			t.Errorf("to.id = %v, want 536e564f316c83104c000020", to["id"])
		}
		fmt.Fprint(w, `{
			"type": "admin_message",
			"id": "19",
			"created_at": 1734537786,
			"body": "Hey there!",
			"message_type": "inapp",
			"conversation_id": "614"
		}`)
	})

	ctx := context.Background()
	msg, err := client.Messages.Create(ctx, &CreateMessageRequest{
		MessageType: "in_app",
		Body:        "Hey there!",
		From: MessageSender{
			Type: "admin",
			ID:   "394051",
		},
		To: MessageRecipient{
			Type: "user",
			ID:   "536e564f316c83104c000020",
		},
	})
	if err != nil {
		t.Fatalf("Messages.Create returned error: %v", err)
	}
	if msg.Type != "admin_message" {
		t.Errorf("Type = %v, want admin_message", msg.Type)
	}
	if msg.ID != "19" {
		t.Errorf("ID = %v, want 19", msg.ID)
	}
	if msg.MessageType != "inapp" {
		t.Errorf("MessageType = %v, want inapp", msg.MessageType)
	}
	if msg.ConversationID != "614" {
		t.Errorf("ConversationID = %v, want 614", msg.ConversationID)
	}
	if msg.Body != "Hey there!" {
		t.Errorf("Body = %v, want Hey there!", msg.Body)
	}
}

func TestMessagesService_Create_Email(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["message_type"] != "email" {
			t.Errorf("message_type = %v, want email", body["message_type"])
		}
		if body["subject"] != "Thanks for everything" {
			t.Errorf("subject = %v, want Thanks for everything", body["subject"])
		}
		if body["body"] != "<p>Hello there</p>" {
			t.Errorf("body = %v, want <p>Hello there</p>", body["body"])
		}
		if body["template"] != "personal" {
			t.Errorf("template = %v, want personal", body["template"])
		}
		fmt.Fprint(w, `{
			"type": "admin_message",
			"id": "20",
			"created_at": 1734537790,
			"subject": "Thanks for everything",
			"body": "<p>Hello there</p>",
			"message_type": "email",
			"conversation_id": "615"
		}`)
	})

	ctx := context.Background()
	msg, err := client.Messages.Create(ctx, &CreateMessageRequest{
		MessageType: "email",
		Subject:     "Thanks for everything",
		Body:        "<p>Hello there</p>",
		Template:    "personal",
		From: MessageSender{
			Type: "admin",
			ID:   "394051",
		},
		To: MessageRecipient{
			Type: "user",
			ID:   "536e564f316c83104c000020",
		},
	})
	if err != nil {
		t.Fatalf("Messages.Create returned error: %v", err)
	}
	if msg.Subject != "Thanks for everything" {
		t.Errorf("Subject = %v, want Thanks for everything", msg.Subject)
	}
	if msg.MessageType != "email" {
		t.Errorf("MessageType = %v, want email", msg.MessageType)
	}
}

func TestMessagesService_Create_WithCreatedAt(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["created_at"] != float64(1590000000) {
			t.Errorf("created_at = %v, want 1590000000", body["created_at"])
		}
		fmt.Fprint(w, `{
			"type": "admin_message",
			"id": "21",
			"created_at": 1590000000,
			"body": "Hello",
			"message_type": "inapp"
		}`)
	})

	ctx := context.Background()
	msg, err := client.Messages.Create(ctx, &CreateMessageRequest{
		MessageType: "in_app",
		Body:        "Hello",
		CreatedAt:   1590000000,
		From: MessageSender{
			Type: "admin",
			ID:   "394051",
		},
		To: MessageRecipient{
			Type: "user",
			ID:   "536e564f316c83104c000020",
		},
	})
	if err != nil {
		t.Fatalf("Messages.Create returned error: %v", err)
	}
	if msg.CreatedAt != 1590000000 {
		t.Errorf("CreatedAt = %v, want 1590000000", msg.CreatedAt)
	}
}

func TestMessagesService_Create_WithCreateConversation(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["create_conversation_without_contact_reply"] != true {
			t.Errorf("create_conversation_without_contact_reply = %v, want true", body["create_conversation_without_contact_reply"])
		}
		fmt.Fprint(w, `{
			"type": "admin_message",
			"id": "22",
			"created_at": 1734537800,
			"body": "Hi",
			"message_type": "inapp",
			"conversation_id": "616"
		}`)
	})

	ctx := context.Background()
	createConvo := true
	msg, err := client.Messages.Create(ctx, &CreateMessageRequest{
		MessageType:                           "in_app",
		Body:                                  "Hi",
		CreateConversationWithoutContactReply: &createConvo,
		From: MessageSender{
			Type: "admin",
			ID:   "394051",
		},
		To: MessageRecipient{
			Type: "lead",
			ID:   "536e564f316c83104c000021",
		},
	})
	if err != nil {
		t.Fatalf("Messages.Create returned error: %v", err)
	}
	if msg.ConversationID != "616" {
		t.Errorf("ConversationID = %v, want 616", msg.ConversationID)
	}
}

func TestMessagesService_Create_Unauthorized(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{
			"type": "error.list",
			"request_id": "req-123",
			"errors": [{"code": "unauthorized", "message": "Access Token Invalid"}]
		}`)
	})

	ctx := context.Background()
	_, err := client.Messages.Create(ctx, &CreateMessageRequest{
		MessageType: "in_app",
		Body:        "Hello",
		From:        MessageSender{Type: "admin", ID: "394051"},
		To:          MessageRecipient{Type: "user", ID: "123"},
	})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !IsUnauthorized(err) {
		t.Errorf("Expected IsUnauthorized, got: %v", err)
	}
}

func TestMessagesService_Create_BadRequest(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{
			"type": "error.list",
			"request_id": "req-456",
			"errors": [{"code": "parameter_invalid", "message": "No body supplied for email message"}]
		}`)
	})

	ctx := context.Background()
	_, err := client.Messages.Create(ctx, &CreateMessageRequest{
		MessageType: "email",
		From:        MessageSender{Type: "admin", ID: "394051"},
		To:          MessageRecipient{Type: "user", ID: "123"},
	})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestMessagesService_CreateRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{
			"type": "admin_message",
			"id": "19",
			"created_at": 1734537786,
			"body": "Hey there!",
			"message_type": "inapp",
			"conversation_id": "614"
		}`)
	})

	ctx := context.Background()
	result, err := client.Messages.CreateRaw(ctx, &CreateMessageRequest{
		MessageType: "in_app",
		Body:        "Hey there!",
		From:        MessageSender{Type: "admin", ID: "394051"},
		To:          MessageRecipient{Type: "user", ID: "536e564f316c83104c000020"},
	})
	if err != nil {
		t.Fatalf("Messages.CreateRaw returned error: %v", err)
	}
	if result.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
	msg, err := ParseMessageCreateResult(result)
	if err != nil {
		t.Fatalf("ParseMessageCreateResult returned error: %v", err)
	}
	if msg.ID != "19" {
		t.Errorf("ID = %v, want 19", msg.ID)
	}
	if msg.Body != "Hey there!" {
		t.Errorf("Body = %v, want Hey there!", msg.Body)
	}
}

func TestMessagesService_CreateRaw_Unauthorized(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{
			"type": "error.list",
			"request_id": "req-123",
			"errors": [{"code": "unauthorized", "message": "Access Token Invalid"}]
		}`)
	})

	ctx := context.Background()
	result, err := client.Messages.CreateRaw(ctx, &CreateMessageRequest{
		MessageType: "in_app",
		Body:        "Hello",
		From:        MessageSender{Type: "admin", ID: "394051"},
		To:          MessageRecipient{Type: "user", ID: "123"},
	})
	if err != nil {
		t.Fatalf("Messages.CreateRaw returned error: %v", err)
	}
	if result.Error == nil {
		t.Fatal("Expected Error to be populated")
	}
	if result.Error.Code != "unauthorized" {
		t.Errorf("Error.Code = %v, want unauthorized", result.Error.Code)
	}
}

func TestMessagesService_Create_WithCCBCC(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		cc, ok := body["cc"].([]any)
		if !ok || len(cc) != 1 {
			t.Fatalf("cc length = %v, want 1", len(cc))
		}
		ccItem := cc[0].(map[string]any)
		if ccItem["type"] != "user" || ccItem["id"] != "cc-user-1" {
			t.Errorf("cc[0] = %v, want {type:user, id:cc-user-1}", ccItem)
		}
		bcc, ok := body["bcc"].([]any)
		if !ok || len(bcc) != 1 {
			t.Fatalf("bcc length = %v, want 1", len(bcc))
		}
		bccItem := bcc[0].(map[string]any)
		if bccItem["type"] != "user" || bccItem["id"] != "bcc-user-1" {
			t.Errorf("bcc[0] = %v, want {type:user, id:bcc-user-1}", bccItem)
		}
		fmt.Fprint(w, `{"type":"message","id":"msg-1","message_type":"conversation","body":"Hello"}`)
	})

	ctx := context.Background()
	msg, err := client.Messages.Create(ctx, &CreateMessageRequest{
		MessageType: "conversation",
		Body:        "Hello",
		From:        MessageSender{Type: "admin", ID: "admin-1"},
		To:          MessageRecipient{Type: "user", ID: "user-1"},
		CC:          []MessageRecipient{{Type: "user", ID: "cc-user-1"}},
		BCC:         []MessageRecipient{{Type: "user", ID: "bcc-user-1"}},
	})
	if err != nil {
		t.Fatalf("Messages.Create returned error: %v", err)
	}
	if msg.ID != "msg-1" {
		t.Errorf("Message.ID = %v, want msg-1", msg.ID)
	}
}
