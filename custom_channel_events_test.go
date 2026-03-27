package intercom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestCustomChannelEventsService_NotifyNewConversation(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/custom_channel_events/notify_new_conversation", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["event_id"] != "evt_1" {
			t.Errorf("event_id = %v, want evt_1", body["event_id"])
		}
		if body["external_conversation_id"] != "ext_conv_1" {
			t.Errorf("external_conversation_id = %v, want ext_conv_1", body["external_conversation_id"])
		}
		fmt.Fprint(w, `{
			"external_conversation_id":"ext_conv_1",
			"conversation_id":"conv_123",
			"external_contact_id":"ext_contact_1",
			"contact_id":"contact_123"
		}`)
	})

	ctx := context.Background()
	resp, err := client.CustomChannelEvents.NotifyNewConversation(ctx, &CustomChannelBaseEvent{
		EventID:                "evt_1",
		ExternalConversationID: "ext_conv_1",
		Contact: CustomChannelContact{
			Type:       "user",
			ExternalID: "ext_contact_1",
			Name:       "Test User",
		},
	})
	if err != nil {
		t.Fatalf("NotifyNewConversation returned error: %v", err)
	}
	if resp.ConversationID != "conv_123" {
		t.Errorf("ConversationID = %v, want conv_123", resp.ConversationID)
	}
	if resp.ContactID != "contact_123" {
		t.Errorf("ContactID = %v, want contact_123", resp.ContactID)
	}
}

func TestCustomChannelEventsService_NotifyNewMessage(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/custom_channel_events/notify_new_message", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["body"] != "Hello there" {
			t.Errorf("body = %v, want Hello there", body["body"])
		}
		fmt.Fprint(w, `{
			"external_conversation_id":"ext_conv_1",
			"conversation_id":"conv_123",
			"external_contact_id":"ext_contact_1",
			"contact_id":"contact_123"
		}`)
	})

	ctx := context.Background()
	resp, err := client.CustomChannelEvents.NotifyNewMessage(ctx, &CustomChannelMessageEvent{
		CustomChannelBaseEvent: CustomChannelBaseEvent{
			EventID:                "evt_2",
			ExternalConversationID: "ext_conv_1",
			Contact: CustomChannelContact{
				Type:       "user",
				ExternalID: "ext_contact_1",
			},
		},
		Body: "Hello there",
	})
	if err != nil {
		t.Fatalf("NotifyNewMessage returned error: %v", err)
	}
	if resp.ConversationID != "conv_123" {
		t.Errorf("ConversationID = %v, want conv_123", resp.ConversationID)
	}
}

func TestCustomChannelEventsService_NotifyQuickReply(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/custom_channel_events/notify_quick_reply_selected", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["quick_reply_option_id"] != "opt_1" {
			t.Errorf("quick_reply_option_id = %v, want opt_1", body["quick_reply_option_id"])
		}
		fmt.Fprint(w, `{
			"external_conversation_id":"ext_conv_1",
			"conversation_id":"conv_123",
			"external_contact_id":"ext_contact_1",
			"contact_id":"contact_123"
		}`)
	})

	ctx := context.Background()
	resp, err := client.CustomChannelEvents.NotifyQuickReply(ctx, &CustomChannelQuickReplyEvent{
		CustomChannelBaseEvent: CustomChannelBaseEvent{
			EventID:                "evt_3",
			ExternalConversationID: "ext_conv_1",
			Contact: CustomChannelContact{
				Type:       "lead",
				ExternalID: "ext_contact_1",
			},
		},
		QuickReplyOptionID: "opt_1",
	})
	if err != nil {
		t.Fatalf("NotifyQuickReply returned error: %v", err)
	}
	if resp.ExternalConversationID != "ext_conv_1" {
		t.Errorf("ExternalConversationID = %v, want ext_conv_1", resp.ExternalConversationID)
	}
}

func TestCustomChannelEventsService_NotifyAttributeCollected(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/custom_channel_events/notify_attribute_collected", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		attr := body["attribute"].(map[string]any)
		if attr["id"] != "attr_1" {
			t.Errorf("attribute.id = %v, want attr_1", attr["id"])
		}
		if attr["value"] != "some_value" {
			t.Errorf("attribute.value = %v, want some_value", attr["value"])
		}
		fmt.Fprint(w, `{
			"external_conversation_id":"ext_conv_1",
			"conversation_id":"conv_123",
			"external_contact_id":"ext_contact_1",
			"contact_id":"contact_123"
		}`)
	})

	ctx := context.Background()
	resp, err := client.CustomChannelEvents.NotifyAttributeCollected(ctx, &CustomChannelAttributeEvent{
		CustomChannelBaseEvent: CustomChannelBaseEvent{
			EventID:                "evt_4",
			ExternalConversationID: "ext_conv_1",
			Contact: CustomChannelContact{
				Type:       "user",
				ExternalID: "ext_contact_1",
			},
		},
		Attribute: CustomChannelAttribute{
			ID:    "attr_1",
			Value: "some_value",
		},
	})
	if err != nil {
		t.Fatalf("NotifyAttributeCollected returned error: %v", err)
	}
	if resp.ContactID != "contact_123" {
		t.Errorf("ContactID = %v, want contact_123", resp.ContactID)
	}
}

// --- Raw companion method tests ---

func TestCustomChannelEventsService_NotifyNewConversationRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/custom_channel_events/notify_new_conversation", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.Header().Set("X-Request-Id", "req-cce-conv")
		fmt.Fprint(w, `{"external_conversation_id":"ext_conv_1","conversation_id":"conv_123","external_contact_id":"ext_contact_1","contact_id":"contact_123"}`)
	})

	ctx := context.Background()
	result, err := client.CustomChannelEvents.NotifyNewConversationRaw(ctx, &CustomChannelBaseEvent{
		EventID:                "evt_1",
		ExternalConversationID: "ext_conv_1",
		Contact:                CustomChannelContact{Type: "user", ExternalID: "ext_contact_1"},
	})
	if err != nil {
		t.Fatalf("NotifyNewConversationRaw returned error: %v", err)
	}
	data, err := ParseCustomChannelEventNotifyNewConversationResult(result)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if data.ConversationID != "conv_123" {
		t.Errorf("ConversationID = %v, want conv_123", data.ConversationID)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-cce-conv" {
		t.Errorf("Header X-Request-Id = %q, want req-cce-conv", result.Header.Get("X-Request-Id"))
	}
}

func TestCustomChannelEventsService_NotifyNewMessageRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/custom_channel_events/notify_new_message", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.Header().Set("X-Request-Id", "req-cce-msg")
		fmt.Fprint(w, `{"external_conversation_id":"ext_conv_1","conversation_id":"conv_123","external_contact_id":"ext_contact_1","contact_id":"contact_123"}`)
	})

	ctx := context.Background()
	result, err := client.CustomChannelEvents.NotifyNewMessageRaw(ctx, &CustomChannelMessageEvent{
		CustomChannelBaseEvent: CustomChannelBaseEvent{
			EventID:                "evt_2",
			ExternalConversationID: "ext_conv_1",
			Contact:                CustomChannelContact{Type: "user", ExternalID: "ext_contact_1"},
		},
		Body: "Hello there",
	})
	if err != nil {
		t.Fatalf("NotifyNewMessageRaw returned error: %v", err)
	}
	data, err := ParseCustomChannelEventNotifyNewMessageResult(result)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if data.ConversationID != "conv_123" {
		t.Errorf("ConversationID = %v, want conv_123", data.ConversationID)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestCustomChannelEventsService_NotifyQuickReplyRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/custom_channel_events/notify_quick_reply_selected", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.Header().Set("X-Request-Id", "req-cce-qr")
		fmt.Fprint(w, `{"external_conversation_id":"ext_conv_1","conversation_id":"conv_123","external_contact_id":"ext_contact_1","contact_id":"contact_123"}`)
	})

	ctx := context.Background()
	result, err := client.CustomChannelEvents.NotifyQuickReplyRaw(ctx, &CustomChannelQuickReplyEvent{
		CustomChannelBaseEvent: CustomChannelBaseEvent{
			EventID:                "evt_3",
			ExternalConversationID: "ext_conv_1",
			Contact:                CustomChannelContact{Type: "lead", ExternalID: "ext_contact_1"},
		},
		QuickReplyOptionID: "opt_1",
	})
	if err != nil {
		t.Fatalf("NotifyQuickReplyRaw returned error: %v", err)
	}
	data, err := ParseCustomChannelEventNotifyQuickReplyResult(result)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if data.ExternalConversationID != "ext_conv_1" {
		t.Errorf("ExternalConversationID = %v, want ext_conv_1", data.ExternalConversationID)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestCustomChannelEventsService_NotifyAttributeCollectedRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/custom_channel_events/notify_attribute_collected", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.Header().Set("X-Request-Id", "req-cce-attr")
		fmt.Fprint(w, `{"external_conversation_id":"ext_conv_1","conversation_id":"conv_123","external_contact_id":"ext_contact_1","contact_id":"contact_123"}`)
	})

	ctx := context.Background()
	result, err := client.CustomChannelEvents.NotifyAttributeCollectedRaw(ctx, &CustomChannelAttributeEvent{
		CustomChannelBaseEvent: CustomChannelBaseEvent{
			EventID:                "evt_4",
			ExternalConversationID: "ext_conv_1",
			Contact:                CustomChannelContact{Type: "user", ExternalID: "ext_contact_1"},
		},
		Attribute: CustomChannelAttribute{ID: "attr_1", Value: "some_value"},
	})
	if err != nil {
		t.Fatalf("NotifyAttributeCollectedRaw returned error: %v", err)
	}
	data, err := ParseCustomChannelEventNotifyAttributeCollectedResult(result)
	if err != nil {
		t.Fatalf("Parse returned error: %v", err)
	}
	if data.ContactID != "contact_123" {
		t.Errorf("ContactID = %v, want contact_123", data.ContactID)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-cce-attr" {
		t.Errorf("Header X-Request-Id = %q, want req-cce-attr", result.Header.Get("X-Request-Id"))
	}
}
