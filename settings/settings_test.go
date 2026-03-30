package settings_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rassakhatsky/intercom-go-sdk/api"
	"github.com/rassakhatsky/intercom-go-sdk/settings"
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
	noRedirectClient := &http.Client{
		Transport: tc.client.Transport,
		Timeout:   tc.client.Timeout,
		CheckRedirect: func(req *http.Request, via []*http.Request) error {
			return http.ErrUseLastResponse
		},
	}
	req = req.WithContext(ctx)
	resp, err := noRedirectClient.Do(req)
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

type services struct {
	brands        *settings.BrandsService
	ipAllowlist   *settings.IPAllowlistService
	channelEvents *settings.ChannelEventsService
	jobs          *settings.JobsService
	notes         *settings.NotesService
}

func setup() (svcs services, mux *http.ServeMux, teardown func()) {
	mux = http.NewServeMux()
	server := httptest.NewServer(mux)
	caller := &testCaller{baseURL: server.URL, client: server.Client()}
	svcs = services{
		brands:        settings.NewBrandsService(caller),
		ipAllowlist:   settings.NewIPAllowlistService(caller),
		channelEvents: settings.NewChannelEventsService(caller),
		jobs:          settings.NewJobsService(caller),
		notes:         settings.NewNotesService(caller),
	}
	return svcs, mux, server.Close
}

// --- Brands Tests ---

func TestBrandsService_Get(t *testing.T) {
	svcs, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/brands/brand_123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testHeader(t, r, "Authorization", "Bearer test-token")
		fmt.Fprint(w, `{
			"type":"brand",
			"id":"brand_123",
			"name":"My Brand",
			"is_default":true,
			"created_at":1700000000,
			"updated_at":1700000100,
			"help_center_id":"hc_1",
			"default_address_settings_id":"das_1"
		}`)
	})

	ctx := context.Background()
	brand, err := svcs.brands.Get(ctx, "brand_123")
	if err != nil {
		t.Fatalf("Brands.Get returned error: %v", err)
	}
	if brand.ID != "brand_123" {
		t.Errorf("Brand.ID = %v, want brand_123", brand.ID)
	}
	if brand.Name != "My Brand" {
		t.Errorf("Brand.Name = %v, want My Brand", brand.Name)
	}
	if !brand.IsDefault {
		t.Error("Brand.IsDefault = false, want true")
	}
	if brand.HelpCenterID != "hc_1" {
		t.Errorf("Brand.HelpCenterID = %v, want hc_1", brand.HelpCenterID)
	}
}

func TestBrandsService_Get_NotFound(t *testing.T) {
	svcs, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/brands/nonexistent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"Brand not found"}]}`)
	})

	ctx := context.Background()
	_, err := svcs.brands.Get(ctx, "nonexistent")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !api.IsNotFound(err) {
		t.Errorf("Expected IsNotFound, got: %v", err)
	}
}

func TestBrandsService_List(t *testing.T) {
	svcs, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/brands", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"list",
			"data":[
				{"type":"brand","id":"brand_1","name":"Default Brand","is_default":true},
				{"type":"brand","id":"brand_2","name":"Secondary","is_default":false}
			]
		}`)
	})

	ctx := context.Background()
	result, err := svcs.brands.List(ctx)
	if err != nil {
		t.Fatalf("Brands.List returned error: %v", err)
	}
	if result.Type != "list" {
		t.Errorf("Type = %v, want list", result.Type)
	}
	if len(result.Data) != 2 {
		t.Fatalf("Data length = %d, want 2", len(result.Data))
	}
	if result.Data[0].Name != "Default Brand" {
		t.Errorf("Data[0].Name = %v, want Default Brand", result.Data[0].Name)
	}
	if result.Data[1].IsDefault {
		t.Error("Data[1].IsDefault = true, want false")
	}
}

func TestBrandsService_GetRaw_Success(t *testing.T) {
	svcs, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/brands/brand_123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-brand-get")
		fmt.Fprint(w, `{"type":"brand","id":"brand_123","name":"My Brand","is_default":true}`)
	})

	ctx := context.Background()
	result, err := svcs.brands.GetRaw(ctx, "brand_123")
	if err != nil {
		t.Fatalf("GetRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-brand-get" {
		t.Errorf("Header X-Request-Id = %q, want req-brand-get", result.Header.Get("X-Request-Id"))
	}
	if result.Error != nil {
		t.Errorf("Result.Error = %v, want nil", result.Error)
	}
	brand, err := settings.ParseGetResult(result)
	if err != nil {
		t.Fatalf("ParseGetResult returned error: %v", err)
	}
	if brand.ID != "brand_123" {
		t.Errorf("Brand.ID = %v, want brand_123", brand.ID)
	}
	if brand.Name != "My Brand" {
		t.Errorf("Brand.Name = %v, want My Brand", brand.Name)
	}
}

func TestBrandsService_GetRaw_NotFound(t *testing.T) {
	svcs, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/brands/nonexistent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"Brand not found"}]}`)
	})

	ctx := context.Background()
	result, err := svcs.brands.GetRaw(ctx, "nonexistent")
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

func TestBrandsService_ListRaw_Success(t *testing.T) {
	svcs, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/brands", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-brand-list")
		fmt.Fprint(w, `{"type":"list","data":[{"type":"brand","id":"brand_1","name":"Default Brand"},{"type":"brand","id":"brand_2","name":"Secondary"}]}`)
	})

	ctx := context.Background()
	result, err := svcs.brands.ListRaw(ctx)
	if err != nil {
		t.Fatalf("ListRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-brand-list" {
		t.Errorf("Header X-Request-Id = %q, want req-brand-list", result.Header.Get("X-Request-Id"))
	}
	brands, err := settings.ParseListResult(result)
	if err != nil {
		t.Fatalf("ParseListResult returned error: %v", err)
	}
	if len(brands.Data) != 2 {
		t.Fatalf("Data length = %d, want 2", len(brands.Data))
	}
	if brands.Data[0].Name != "Default Brand" {
		t.Errorf("Data[0].Name = %v, want Default Brand", brands.Data[0].Name)
	}
}

// --- IP Allowlist Tests ---

func TestIPAllowlistService_Get(t *testing.T) {
	svcs, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ip_allowlist", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testHeader(t, r, "Authorization", "Bearer test-token")
		fmt.Fprint(w, `{
			"type":"ip_allowlist",
			"enabled":true,
			"ip_allowlist":["192.168.1.0/24","10.0.0.1"]
		}`)
	})

	ctx := context.Background()
	result, err := svcs.ipAllowlist.Get(ctx)
	if err != nil {
		t.Fatalf("IPAllowlist.Get returned error: %v", err)
	}
	if result.Type != "ip_allowlist" {
		t.Errorf("Type = %v, want ip_allowlist", result.Type)
	}
	if !result.Enabled {
		t.Error("Enabled = false, want true")
	}
	if len(result.IPAllowlist) != 2 {
		t.Fatalf("IPAllowlist length = %d, want 2", len(result.IPAllowlist))
	}
	if result.IPAllowlist[0] != "192.168.1.0/24" {
		t.Errorf("IPAllowlist[0] = %v, want 192.168.1.0/24", result.IPAllowlist[0])
	}
}

func TestIPAllowlistService_Update(t *testing.T) {
	svcs, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ip_allowlist", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["enabled"] != true {
			t.Errorf("enabled = %v, want true", body["enabled"])
		}
		ips := body["ip_allowlist"].([]any)
		if len(ips) != 1 {
			t.Errorf("ip_allowlist length = %d, want 1", len(ips))
		}
		fmt.Fprint(w, `{
			"type":"ip_allowlist",
			"enabled":true,
			"ip_allowlist":["10.0.0.0/8"]
		}`)
	})

	ctx := context.Background()
	result, err := svcs.ipAllowlist.Update(ctx, &settings.UpdateIPAllowlistRequest{
		Enabled:     true,
		IPAllowlist: []string{"10.0.0.0/8"},
	})
	if err != nil {
		t.Fatalf("IPAllowlist.Update returned error: %v", err)
	}
	if !result.Enabled {
		t.Error("Enabled = false, want true")
	}
	if len(result.IPAllowlist) != 1 {
		t.Fatalf("IPAllowlist length = %d, want 1", len(result.IPAllowlist))
	}
}

func TestIPAllowlistService_GetRaw_Success(t *testing.T) {
	svcs, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ip_allowlist", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-ip-get")
		fmt.Fprint(w, `{"type":"ip_allowlist","enabled":true,"ip_allowlist":["192.168.1.0/24"]}`)
	})

	ctx := context.Background()
	result, err := svcs.ipAllowlist.GetRaw(ctx)
	if err != nil {
		t.Fatalf("GetRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Error != nil {
		t.Errorf("Result.Error = %v, want nil", result.Error)
	}
	s, err := settings.ParseIPAllowlistGetResult(result)
	if err != nil {
		t.Fatalf("ParseIPAllowlistGetResult returned error: %v", err)
	}
	if !s.Enabled {
		t.Error("Enabled = false, want true")
	}
	if len(s.IPAllowlist) != 1 {
		t.Fatalf("IPAllowlist length = %d, want 1", len(s.IPAllowlist))
	}
}

func TestIPAllowlistService_UpdateRaw_Success(t *testing.T) {
	svcs, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ip_allowlist", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		w.Header().Set("X-Request-Id", "req-ip-update")
		fmt.Fprint(w, `{"type":"ip_allowlist","enabled":true,"ip_allowlist":["10.0.0.0/8"]}`)
	})

	ctx := context.Background()
	result, err := svcs.ipAllowlist.UpdateRaw(ctx, &settings.UpdateIPAllowlistRequest{
		Enabled:     true,
		IPAllowlist: []string{"10.0.0.0/8"},
	})
	if err != nil {
		t.Fatalf("UpdateRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	s, err := settings.ParseIPAllowlistUpdateResult(result)
	if err != nil {
		t.Fatalf("ParseIPAllowlistUpdateResult returned error: %v", err)
	}
	if !s.Enabled {
		t.Error("Enabled = false, want true")
	}
}

func TestIPAllowlistService_Update_WithType(t *testing.T) {
	svcs, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ip_allowlist", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["type"] != "ip_allowlist" {
			t.Errorf("type = %v, want ip_allowlist", body["type"])
		}
		fmt.Fprint(w, `{"type":"ip_allowlist","enabled":true,"ip_allowlist":["10.0.0.1"]}`)
	})

	ctx := context.Background()
	s, err := svcs.ipAllowlist.Update(ctx, &settings.UpdateIPAllowlistRequest{
		Type:        "ip_allowlist",
		Enabled:     true,
		IPAllowlist: []string{"10.0.0.1"},
	})
	if err != nil {
		t.Fatalf("IPAllowlist.Update returned error: %v", err)
	}
	if !s.Enabled {
		t.Error("Enabled = false, want true")
	}
}

// --- Custom Channel Events Tests ---

func TestChannelEventsService_NotifyNewConversation(t *testing.T) {
	svcs, mux, teardown := setup()
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
	resp, err := svcs.channelEvents.NotifyNewConversation(ctx, &settings.BaseEvent{
		EventID:                "evt_1",
		ExternalConversationID: "ext_conv_1",
		Contact: settings.Contact{
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

func TestChannelEventsService_NotifyNewMessage(t *testing.T) {
	svcs, mux, teardown := setup()
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
	resp, err := svcs.channelEvents.NotifyNewMessage(ctx, &settings.MessageEvent{
		BaseEvent: settings.BaseEvent{
			EventID:                "evt_2",
			ExternalConversationID: "ext_conv_1",
			Contact: settings.Contact{
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

func TestChannelEventsService_NotifyQuickReply(t *testing.T) {
	svcs, mux, teardown := setup()
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
	resp, err := svcs.channelEvents.NotifyQuickReply(ctx, &settings.QuickReplyEvent{
		BaseEvent: settings.BaseEvent{
			EventID:                "evt_3",
			ExternalConversationID: "ext_conv_1",
			Contact: settings.Contact{
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

func TestChannelEventsService_NotifyAttributeCollected(t *testing.T) {
	svcs, mux, teardown := setup()
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
	resp, err := svcs.channelEvents.NotifyAttributeCollected(ctx, &settings.AttributeEvent{
		BaseEvent: settings.BaseEvent{
			EventID:                "evt_4",
			ExternalConversationID: "ext_conv_1",
			Contact: settings.Contact{
				Type:       "user",
				ExternalID: "ext_contact_1",
			},
		},
		Attribute: settings.Attribute{
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

func TestChannelEventsService_NotifyNewConversationRaw_Success(t *testing.T) {
	svcs, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/custom_channel_events/notify_new_conversation", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.Header().Set("X-Request-Id", "req-cce-conv")
		fmt.Fprint(w, `{"external_conversation_id":"ext_conv_1","conversation_id":"conv_123","external_contact_id":"ext_contact_1","contact_id":"contact_123"}`)
	})

	ctx := context.Background()
	result, err := svcs.channelEvents.NotifyNewConversationRaw(ctx, &settings.BaseEvent{
		EventID:                "evt_1",
		ExternalConversationID: "ext_conv_1",
		Contact:                settings.Contact{Type: "user", ExternalID: "ext_contact_1"},
	})
	if err != nil {
		t.Fatalf("NotifyNewConversationRaw returned error: %v", err)
	}
	data, err := settings.ParseNotifyNewConversationResult(result)
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

func TestChannelEventsService_NotifyNewMessageRaw_Success(t *testing.T) {
	svcs, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/custom_channel_events/notify_new_message", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.Header().Set("X-Request-Id", "req-cce-msg")
		fmt.Fprint(w, `{"external_conversation_id":"ext_conv_1","conversation_id":"conv_123","external_contact_id":"ext_contact_1","contact_id":"contact_123"}`)
	})

	ctx := context.Background()
	result, err := svcs.channelEvents.NotifyNewMessageRaw(ctx, &settings.MessageEvent{
		BaseEvent: settings.BaseEvent{
			EventID:                "evt_2",
			ExternalConversationID: "ext_conv_1",
			Contact:                settings.Contact{Type: "user", ExternalID: "ext_contact_1"},
		},
		Body: "Hello there",
	})
	if err != nil {
		t.Fatalf("NotifyNewMessageRaw returned error: %v", err)
	}
	data, err := settings.ParseNotifyNewMessageResult(result)
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

func TestChannelEventsService_NotifyQuickReplyRaw_Success(t *testing.T) {
	svcs, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/custom_channel_events/notify_quick_reply_selected", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.Header().Set("X-Request-Id", "req-cce-qr")
		fmt.Fprint(w, `{"external_conversation_id":"ext_conv_1","conversation_id":"conv_123","external_contact_id":"ext_contact_1","contact_id":"contact_123"}`)
	})

	ctx := context.Background()
	result, err := svcs.channelEvents.NotifyQuickReplyRaw(ctx, &settings.QuickReplyEvent{
		BaseEvent: settings.BaseEvent{
			EventID:                "evt_3",
			ExternalConversationID: "ext_conv_1",
			Contact:                settings.Contact{Type: "lead", ExternalID: "ext_contact_1"},
		},
		QuickReplyOptionID: "opt_1",
	})
	if err != nil {
		t.Fatalf("NotifyQuickReplyRaw returned error: %v", err)
	}
	data, err := settings.ParseNotifyQuickReplyResult(result)
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

func TestChannelEventsService_NotifyAttributeCollectedRaw_Success(t *testing.T) {
	svcs, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/custom_channel_events/notify_attribute_collected", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		w.Header().Set("X-Request-Id", "req-cce-attr")
		fmt.Fprint(w, `{"external_conversation_id":"ext_conv_1","conversation_id":"conv_123","external_contact_id":"ext_contact_1","contact_id":"contact_123"}`)
	})

	ctx := context.Background()
	result, err := svcs.channelEvents.NotifyAttributeCollectedRaw(ctx, &settings.AttributeEvent{
		BaseEvent: settings.BaseEvent{
			EventID:                "evt_4",
			ExternalConversationID: "ext_conv_1",
			Contact:                settings.Contact{Type: "user", ExternalID: "ext_contact_1"},
		},
		Attribute: settings.Attribute{ID: "attr_1", Value: "some_value"},
	})
	if err != nil {
		t.Fatalf("NotifyAttributeCollectedRaw returned error: %v", err)
	}
	data, err := settings.ParseNotifyAttributeCollectedResult(result)
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

// --- Jobs Tests ---

func TestJobsService_GetStatus(t *testing.T) {
	svcs, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/jobs/status/job_123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testHeader(t, r, "Authorization", "Bearer test-token")
		fmt.Fprint(w, `{
			"type":"job",
			"id":"job_123",
			"url":"https://api.intercom.io/jobs/status/job_123",
			"status":"success",
			"resource_type":"ticket",
			"resource_id":"ticket_456",
			"resource_url":"https://api.intercom.io/tickets/ticket_456"
		}`)
	})

	ctx := context.Background()
	job, err := svcs.jobs.GetStatus(ctx, "job_123")
	if err != nil {
		t.Fatalf("Jobs.GetStatus returned error: %v", err)
	}
	if job.ID != "job_123" {
		t.Errorf("ID = %v, want job_123", job.ID)
	}
	if job.Status != "success" {
		t.Errorf("Status = %v, want success", job.Status)
	}
	if job.ResourceType != "ticket" {
		t.Errorf("ResourceType = %v, want ticket", job.ResourceType)
	}
	if job.ResourceID != "ticket_456" {
		t.Errorf("ResourceID = %v, want ticket_456", job.ResourceID)
	}
}

func TestJobsService_GetStatus_NotFound(t *testing.T) {
	svcs, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/jobs/status/nonexistent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"Job not found"}]}`)
	})

	ctx := context.Background()
	_, err := svcs.jobs.GetStatus(ctx, "nonexistent")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !api.IsNotFound(err) {
		t.Errorf("Expected IsNotFound, got: %v", err)
	}
}

func TestJobsService_GetStatus_Pending(t *testing.T) {
	svcs, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/jobs/status/job_pending", func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{
			"type":"job",
			"id":"job_pending",
			"url":"https://api.intercom.io/jobs/status/job_pending",
			"status":"pending",
			"resource_type":"ticket",
			"resource_id":null,
			"resource_url":null
		}`)
	})

	ctx := context.Background()
	job, err := svcs.jobs.GetStatus(ctx, "job_pending")
	if err != nil {
		t.Fatalf("Jobs.GetStatus returned error: %v", err)
	}
	if job.Status != "pending" {
		t.Errorf("Status = %v, want pending", job.Status)
	}
	if job.ResourceID != "" {
		t.Errorf("ResourceID = %v, want empty", job.ResourceID)
	}
}

func TestJobsService_GetStatusRaw(t *testing.T) {
	svcs, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/jobs/status/job_123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"job",
			"id":"job_123",
			"status":"success",
			"resource_type":"ticket",
			"resource_id":"ticket_456"
		}`)
	})

	ctx := context.Background()
	result, err := svcs.jobs.GetStatusRaw(ctx, "job_123")
	if err != nil {
		t.Fatalf("Jobs.GetStatusRaw returned error: %v", err)
	}
	if result.StatusCode != http.StatusOK {
		t.Errorf("StatusCode = %d, want %d", result.StatusCode, http.StatusOK)
	}
	job, err := settings.ParseJobGetStatusResult(result)
	if err != nil {
		t.Fatalf("ParseJobGetStatusResult returned error: %v", err)
	}
	if job.ID != "job_123" {
		t.Errorf("ID = %v, want job_123", job.ID)
	}
	if job.Status != "success" {
		t.Errorf("Status = %v, want success", job.Status)
	}
}

func TestJobsService_GetStatusRaw_NotFound(t *testing.T) {
	svcs, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/jobs/status/nonexistent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"Job not found"}]}`)
	})

	ctx := context.Background()
	result, err := svcs.jobs.GetStatusRaw(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("Jobs.GetStatusRaw returned error: %v", err)
	}
	if result.Error == nil {
		t.Fatal("Expected Error to be populated")
	}
	if result.Error.Code != "not_found" {
		t.Errorf("Error.Code = %v, want not_found", result.Error.Code)
	}
}

// --- Notes Tests ---

func TestNotesService_Get(t *testing.T) {
	svcs, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/notes/34", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testHeader(t, r, "Authorization", "Bearer test-token")
		fmt.Fprint(w, `{
			"type": "note",
			"id": "34",
			"created_at": 1733846617,
			"contact": {
				"type": "contact",
				"id": "6762f2591bb69f9f2193bc1f"
			},
			"author": {
				"type": "admin",
				"id": "991267864",
				"name": "Ciaran Lee",
				"email": "admin@email.com"
			},
			"body": "<p>This is a note.</p>"
		}`)
	})

	ctx := context.Background()
	note, err := svcs.notes.Get(ctx, "34")
	if err != nil {
		t.Fatalf("Notes.Get returned error: %v", err)
	}
	if note.Type != "note" {
		t.Errorf("Note.Type = %v, want note", note.Type)
	}
	if note.ID != "34" {
		t.Errorf("Note.ID = %v, want 34", note.ID)
	}
	if note.CreatedAt != 1733846617 {
		t.Errorf("Note.CreatedAt = %v, want 1733846617", note.CreatedAt)
	}
	if note.Body != "<p>This is a note.</p>" {
		t.Errorf("Note.Body = %v, want <p>This is a note.</p>", note.Body)
	}
	if note.Contact == nil {
		t.Fatal("Note.Contact is nil")
	}
	if note.Contact.Type != "contact" {
		t.Errorf("Note.Contact.Type = %v, want contact", note.Contact.Type)
	}
	if note.Contact.ID != "6762f2591bb69f9f2193bc1f" {
		t.Errorf("Note.Contact.ID = %v, want 6762f2591bb69f9f2193bc1f", note.Contact.ID)
	}
	if note.Author == nil {
		t.Fatal("Note.Author is nil")
	}
	if note.Author.Type != "admin" {
		t.Errorf("Note.Author.Type = %v, want admin", note.Author.Type)
	}
	if note.Author.ID != "991267864" {
		t.Errorf("Note.Author.ID = %v, want 991267864", note.Author.ID)
	}
	if note.Author.Name != "Ciaran Lee" {
		t.Errorf("Note.Author.Name = %v, want Ciaran Lee", note.Author.Name)
	}
	if note.Author.Email != "admin@email.com" {
		t.Errorf("Note.Author.Email = %v, want admin@email.com", note.Author.Email)
	}
}

func TestNotesService_Get_NotFound(t *testing.T) {
	svcs, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/notes/nonexistent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{
			"type": "error.list",
			"request_id": "req-123",
			"errors": [{"code": "not_found", "message": "Note not found"}]
		}`)
	})

	ctx := context.Background()
	_, err := svcs.notes.Get(ctx, "nonexistent")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !api.IsNotFound(err) {
		t.Errorf("Expected IsNotFound, got: %v", err)
	}
}

func TestNotesService_GetRaw_Success(t *testing.T) {
	svcs, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/notes/34", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-note")
		fmt.Fprint(w, `{"type":"note","id":"34","created_at":1733846617,"body":"<p>This is a note.</p>","contact":{"type":"contact","id":"abc123"},"author":{"type":"admin","id":"991267864","name":"Ciaran Lee","email":"admin@email.com"}}`)
	})

	ctx := context.Background()
	result, err := svcs.notes.GetRaw(ctx, "34")
	if err != nil {
		t.Fatalf("GetRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-note" {
		t.Errorf("Header X-Request-Id = %q, want req-note", result.Header.Get("X-Request-Id"))
	}
	if result.Error != nil {
		t.Errorf("Result.Error = %v, want nil", result.Error)
	}
	note, err := settings.ParseNoteGetResult(result)
	if err != nil {
		t.Fatalf("ParseNoteGetResult returned error: %v", err)
	}
	if note.ID != "34" {
		t.Errorf("Data.ID = %v, want 34", note.ID)
	}
	if note.Body != "<p>This is a note.</p>" {
		t.Errorf("Data.Body = %v, want <p>This is a note.</p>", note.Body)
	}
}
