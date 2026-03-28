package messaging_test

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
	"github.com/rassakhatsky/intercom-go-sdk/messaging"
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

func newTestCaller() (*testCaller, *http.ServeMux, func()) {
	mux := http.NewServeMux()
	server := httptest.NewServer(mux)
	caller := &testCaller{baseURL: server.URL, client: server.Client()}
	return caller, mux, server.Close
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

// --- Messages Tests ---

func TestMessagesService_Create_InApp(t *testing.T) {
	caller, mux, teardown := newTestCaller()
	defer teardown()
	svc := messaging.NewMessagesService(caller)

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
	msg, err := svc.Create(ctx, &messaging.CreateMessageRequest{
		MessageType: "in_app",
		Body:        "Hey there!",
		From: messaging.MessageSender{
			Type: "admin",
			ID:   "394051",
		},
		To: messaging.MessageRecipient{
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
	caller, mux, teardown := newTestCaller()
	defer teardown()
	svc := messaging.NewMessagesService(caller)

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
	msg, err := svc.Create(ctx, &messaging.CreateMessageRequest{
		MessageType: "email",
		Subject:     "Thanks for everything",
		Body:        "<p>Hello there</p>",
		Template:    "personal",
		From: messaging.MessageSender{
			Type: "admin",
			ID:   "394051",
		},
		To: messaging.MessageRecipient{
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
	caller, mux, teardown := newTestCaller()
	defer teardown()
	svc := messaging.NewMessagesService(caller)

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
	msg, err := svc.Create(ctx, &messaging.CreateMessageRequest{
		MessageType: "in_app",
		Body:        "Hello",
		CreatedAt:   1590000000,
		From: messaging.MessageSender{
			Type: "admin",
			ID:   "394051",
		},
		To: messaging.MessageRecipient{
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
	caller, mux, teardown := newTestCaller()
	defer teardown()
	svc := messaging.NewMessagesService(caller)

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
	msg, err := svc.Create(ctx, &messaging.CreateMessageRequest{
		MessageType:                           "in_app",
		Body:                                  "Hi",
		CreateConversationWithoutContactReply: &createConvo,
		From: messaging.MessageSender{
			Type: "admin",
			ID:   "394051",
		},
		To: messaging.MessageRecipient{
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
	caller, mux, teardown := newTestCaller()
	defer teardown()
	svc := messaging.NewMessagesService(caller)

	mux.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{
			"type": "error.list",
			"request_id": "req-123",
			"errors": [{"code": "unauthorized", "message": "Access Token Invalid"}]
		}`)
	})

	ctx := context.Background()
	_, err := svc.Create(ctx, &messaging.CreateMessageRequest{
		MessageType: "in_app",
		Body:        "Hello",
		From:        messaging.MessageSender{Type: "admin", ID: "394051"},
		To:          messaging.MessageRecipient{Type: "user", ID: "123"},
	})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !api.IsUnauthorized(err) {
		t.Errorf("Expected IsUnauthorized, got: %v", err)
	}
}

func TestMessagesService_Create_BadRequest(t *testing.T) {
	caller, mux, teardown := newTestCaller()
	defer teardown()
	svc := messaging.NewMessagesService(caller)

	mux.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		fmt.Fprint(w, `{
			"type": "error.list",
			"request_id": "req-456",
			"errors": [{"code": "parameter_invalid", "message": "No body supplied for email message"}]
		}`)
	})

	ctx := context.Background()
	_, err := svc.Create(ctx, &messaging.CreateMessageRequest{
		MessageType: "email",
		From:        messaging.MessageSender{Type: "admin", ID: "394051"},
		To:          messaging.MessageRecipient{Type: "user", ID: "123"},
	})
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
}

func TestMessagesService_CreateRaw(t *testing.T) {
	caller, mux, teardown := newTestCaller()
	defer teardown()
	svc := messaging.NewMessagesService(caller)

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
	result, err := svc.CreateRaw(ctx, &messaging.CreateMessageRequest{
		MessageType: "in_app",
		Body:        "Hey there!",
		From:        messaging.MessageSender{Type: "admin", ID: "394051"},
		To:          messaging.MessageRecipient{Type: "user", ID: "536e564f316c83104c000020"},
	})
	if err != nil {
		t.Fatalf("Messages.CreateRaw returned error: %v", err)
	}
	if result.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
	msg, err := messaging.ParseCreateResult(result)
	if err != nil {
		t.Fatalf("ParseCreateResult returned error: %v", err)
	}
	if msg.ID != "19" {
		t.Errorf("ID = %v, want 19", msg.ID)
	}
	if msg.Body != "Hey there!" {
		t.Errorf("Body = %v, want Hey there!", msg.Body)
	}
}

func TestMessagesService_CreateRaw_Unauthorized(t *testing.T) {
	caller, mux, teardown := newTestCaller()
	defer teardown()
	svc := messaging.NewMessagesService(caller)

	mux.HandleFunc("/messages", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{
			"type": "error.list",
			"request_id": "req-123",
			"errors": [{"code": "unauthorized", "message": "Access Token Invalid"}]
		}`)
	})

	ctx := context.Background()
	result, err := svc.CreateRaw(ctx, &messaging.CreateMessageRequest{
		MessageType: "in_app",
		Body:        "Hello",
		From:        messaging.MessageSender{Type: "admin", ID: "394051"},
		To:          messaging.MessageRecipient{Type: "user", ID: "123"},
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
	caller, mux, teardown := newTestCaller()
	defer teardown()
	svc := messaging.NewMessagesService(caller)

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
	msg, err := svc.Create(ctx, &messaging.CreateMessageRequest{
		MessageType: "conversation",
		Body:        "Hello",
		From:        messaging.MessageSender{Type: "admin", ID: "admin-1"},
		To:          messaging.MessageRecipient{Type: "user", ID: "user-1"},
		CC:          []messaging.MessageRecipient{{Type: "user", ID: "cc-user-1"}},
		BCC:         []messaging.MessageRecipient{{Type: "user", ID: "bcc-user-1"}},
	})
	if err != nil {
		t.Fatalf("Messages.Create returned error: %v", err)
	}
	if msg.ID != "msg-1" {
		t.Errorf("Message.ID = %v, want msg-1", msg.ID)
	}
}

// --- Emails Tests ---

func TestEmailsService_Get(t *testing.T) {
	caller, mux, teardown := newTestCaller()
	defer teardown()
	svc := messaging.NewEmailsService(caller)

	mux.HandleFunc("/emails/email_123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testHeader(t, r, "Authorization", "Bearer test-token")
		fmt.Fprint(w, `{
			"type":"email_setting",
			"id":"email_123",
			"email":"support@example.com",
			"verified":true,
			"domain":"example.com",
			"brand_id":"brand_1",
			"forwarding_enabled":true,
			"created_at":1700000000,
			"updated_at":1700000100
		}`)
	})

	ctx := context.Background()
	email, err := svc.Get(ctx, "email_123")
	if err != nil {
		t.Fatalf("Emails.Get returned error: %v", err)
	}
	if email.ID != "email_123" {
		t.Errorf("ID = %v, want email_123", email.ID)
	}
	if email.Email != "support@example.com" {
		t.Errorf("Email = %v, want support@example.com", email.Email)
	}
	if !email.Verified {
		t.Error("Verified = false, want true")
	}
	if email.Domain != "example.com" {
		t.Errorf("Domain = %v, want example.com", email.Domain)
	}
	if !email.ForwardingEnabled {
		t.Error("ForwardingEnabled = false, want true")
	}
}

func TestEmailsService_Get_NotFound(t *testing.T) {
	caller, mux, teardown := newTestCaller()
	defer teardown()
	svc := messaging.NewEmailsService(caller)

	mux.HandleFunc("/emails/nonexistent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"Email not found"}]}`)
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

func TestEmailsService_List(t *testing.T) {
	caller, mux, teardown := newTestCaller()
	defer teardown()
	svc := messaging.NewEmailsService(caller)

	mux.HandleFunc("/emails", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"list",
			"data":[
				{"type":"email_setting","id":"email_1","email":"support@example.com","verified":true},
				{"type":"email_setting","id":"email_2","email":"sales@example.com","verified":false}
			]
		}`)
	})

	ctx := context.Background()
	result, err := svc.List(ctx)
	if err != nil {
		t.Fatalf("Emails.List returned error: %v", err)
	}
	if result.Type != "list" {
		t.Errorf("Type = %v, want list", result.Type)
	}
	if len(result.Data) != 2 {
		t.Fatalf("Data length = %d, want 2", len(result.Data))
	}
	if result.Data[0].Email != "support@example.com" {
		t.Errorf("Data[0].Email = %v, want support@example.com", result.Data[0].Email)
	}
}

func TestEmailsService_GetRaw_Success(t *testing.T) {
	caller, mux, teardown := newTestCaller()
	defer teardown()
	svc := messaging.NewEmailsService(caller)

	mux.HandleFunc("/emails/email_123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-email")
		fmt.Fprint(w, `{"type":"email_setting","id":"email_123","email":"support@example.com","verified":true}`)
	})

	ctx := context.Background()
	result, err := svc.GetRaw(ctx, "email_123")
	if err != nil {
		t.Fatalf("GetRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-email" {
		t.Errorf("Header X-Request-Id = %q, want req-email", result.Header.Get("X-Request-Id"))
	}
	if result.Error != nil {
		t.Errorf("Result.Error = %v, want nil", result.Error)
	}
	email, err := messaging.ParseGetResult(result)
	if err != nil {
		t.Fatalf("ParseGetResult returned error: %v", err)
	}
	if email.ID != "email_123" {
		t.Errorf("ID = %v, want email_123", email.ID)
	}
	if email.Email != "support@example.com" {
		t.Errorf("Email = %v, want support@example.com", email.Email)
	}
}

func TestEmailsService_GetRaw_NotFound(t *testing.T) {
	caller, mux, teardown := newTestCaller()
	defer teardown()
	svc := messaging.NewEmailsService(caller)

	mux.HandleFunc("/emails/nonexistent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"Email not found"}]}`)
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

func TestEmailsService_ListRaw_Success(t *testing.T) {
	caller, mux, teardown := newTestCaller()
	defer teardown()
	svc := messaging.NewEmailsService(caller)

	mux.HandleFunc("/emails", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-list")
		fmt.Fprint(w, `{"type":"list","data":[{"type":"email_setting","id":"email_1","email":"support@example.com"}]}`)
	})

	ctx := context.Background()
	result, err := svc.ListRaw(ctx)
	if err != nil {
		t.Fatalf("ListRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	emails, err := messaging.ParseListResult(result)
	if err != nil {
		t.Fatalf("ParseListResult returned error: %v", err)
	}
	if len(emails.Data) != 1 {
		t.Fatalf("Data length = %d, want 1", len(emails.Data))
	}
	if emails.Data[0].Email != "support@example.com" {
		t.Errorf("Data[0].Email = %v, want support@example.com", emails.Data[0].Email)
	}
}

// --- Subscription Types Tests ---

func TestSubscriptionsService_List(t *testing.T) {
	caller, mux, teardown := newTestCaller()
	defer teardown()
	svc := messaging.NewSubscriptionsService(caller)

	mux.HandleFunc("/subscription_types", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testHeader(t, r, "Authorization", "Bearer test-token")
		fmt.Fprint(w, `{
			"type":"list",
			"data":[
				{
					"type":"subscription",
					"id":"135",
					"state":"live",
					"consent_type":"opt_out",
					"default_translation":{
						"name":"Newsletters",
						"description":"Lorem ipsum dolor sit amet",
						"locale":"en"
					},
					"translations":[
						{
							"name":"Newsletters",
							"description":"Lorem ipsum dolor sit amet",
							"locale":"en"
						}
					],
					"content_types":["email"]
				},
				{
					"type":"subscription",
					"id":"136",
					"state":"draft",
					"consent_type":"opt_in",
					"default_translation":{
						"name":"Product Updates",
						"description":"Get the latest product news",
						"locale":"en"
					},
					"translations":[
						{
							"name":"Product Updates",
							"description":"Get the latest product news",
							"locale":"en"
						},
						{
							"name":"Mises à jour produit",
							"description":"Recevez les dernières nouvelles",
							"locale":"fr"
						}
					],
					"content_types":["email","sms_message"]
				}
			]
		}`)
	})

	ctx := context.Background()
	result, err := svc.List(ctx)
	if err != nil {
		t.Fatalf("SubscriptionTypes.List returned error: %v", err)
	}
	if result.Type != "list" {
		t.Errorf("Type = %v, want list", result.Type)
	}
	if len(result.Data) != 2 {
		t.Fatalf("Data length = %d, want 2", len(result.Data))
	}

	st := result.Data[0]
	if st.Type != "subscription" {
		t.Errorf("Data[0].Type = %v, want subscription", st.Type)
	}
	if st.ID != "135" {
		t.Errorf("Data[0].ID = %v, want 135", st.ID)
	}
	if st.State != "live" {
		t.Errorf("Data[0].State = %v, want live", st.State)
	}
	if st.ConsentType != "opt_out" {
		t.Errorf("Data[0].ConsentType = %v, want opt_out", st.ConsentType)
	}
	if st.DefaultTranslation.Name != "Newsletters" {
		t.Errorf("Data[0].DefaultTranslation.Name = %v, want Newsletters", st.DefaultTranslation.Name)
	}
	if st.DefaultTranslation.Description != "Lorem ipsum dolor sit amet" {
		t.Errorf("Data[0].DefaultTranslation.Description = %v, want Lorem ipsum dolor sit amet", st.DefaultTranslation.Description)
	}
	if st.DefaultTranslation.Locale != "en" {
		t.Errorf("Data[0].DefaultTranslation.Locale = %v, want en", st.DefaultTranslation.Locale)
	}
	if len(st.Translations) != 1 {
		t.Fatalf("Data[0].Translations length = %d, want 1", len(st.Translations))
	}
	if len(st.ContentTypes) != 1 {
		t.Fatalf("Data[0].ContentTypes length = %d, want 1", len(st.ContentTypes))
	}
	if st.ContentTypes[0] != "email" {
		t.Errorf("Data[0].ContentTypes[0] = %v, want email", st.ContentTypes[0])
	}

	st2 := result.Data[1]
	if st2.ID != "136" {
		t.Errorf("Data[1].ID = %v, want 136", st2.ID)
	}
	if st2.State != "draft" {
		t.Errorf("Data[1].State = %v, want draft", st2.State)
	}
	if st2.ConsentType != "opt_in" {
		t.Errorf("Data[1].ConsentType = %v, want opt_in", st2.ConsentType)
	}
	if len(st2.Translations) != 2 {
		t.Fatalf("Data[1].Translations length = %d, want 2", len(st2.Translations))
	}
	if st2.Translations[1].Locale != "fr" {
		t.Errorf("Data[1].Translations[1].Locale = %v, want fr", st2.Translations[1].Locale)
	}
	if len(st2.ContentTypes) != 2 {
		t.Fatalf("Data[1].ContentTypes length = %d, want 2", len(st2.ContentTypes))
	}
	if st2.ContentTypes[1] != "sms_message" {
		t.Errorf("Data[1].ContentTypes[1] = %v, want sms_message", st2.ContentTypes[1])
	}
}

func TestSubscriptionsService_ListRaw_Success(t *testing.T) {
	caller, mux, teardown := newTestCaller()
	defer teardown()
	svc := messaging.NewSubscriptionsService(caller)

	mux.HandleFunc("/subscription_types", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-sub-list")
		fmt.Fprint(w, `{"type":"list","data":[{"type":"subscription","id":"135","state":"live","consent_type":"opt_out"}]}`)
	})

	ctx := context.Background()
	result, err := svc.ListRaw(ctx)
	if err != nil {
		t.Fatalf("ListRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-sub-list" {
		t.Errorf("Header X-Request-Id = %q, want req-sub-list", result.Header.Get("X-Request-Id"))
	}
	stl, err := messaging.ParseSubscriptionListResult(result)
	if err != nil {
		t.Fatalf("ParseSubscriptionListResult returned error: %v", err)
	}
	if len(stl.Data) != 1 {
		t.Fatalf("Data length = %d, want 1", len(stl.Data))
	}
	if stl.Data[0].ID != "135" {
		t.Errorf("Data[0].ID = %v, want 135", stl.Data[0].ID)
	}
}

func TestSubscriptionsService_List_Unauthorized(t *testing.T) {
	caller, mux, teardown := newTestCaller()
	defer teardown()
	svc := messaging.NewSubscriptionsService(caller)

	mux.HandleFunc("/subscription_types", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{
			"type":"error.list",
			"request_id":"req-789",
			"errors":[{"code":"unauthorized","message":"Access Token Invalid"}]
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
