package intercom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestIPAllowlistService_Get(t *testing.T) {
	client, mux, teardown := setup()
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
	result, err := client.IPAllowlist.Get(ctx)
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
	client, mux, teardown := setup()
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
	result, err := client.IPAllowlist.Update(ctx, &UpdateIPAllowlistRequest{
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
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ip_allowlist", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-ip-get")
		fmt.Fprint(w, `{"type":"ip_allowlist","enabled":true,"ip_allowlist":["192.168.1.0/24"]}`)
	})

	ctx := context.Background()
	result, err := client.IPAllowlist.GetRaw(ctx)
	if err != nil {
		t.Fatalf("GetRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Error != nil {
		t.Errorf("Result.Error = %v, want nil", result.Error)
	}
	settings, err := ParseIPAllowlistGetResult(result)
	if err != nil {
		t.Fatalf("ParseIPAllowlistGetResult returned error: %v", err)
	}
	if !settings.Enabled {
		t.Error("Enabled = false, want true")
	}
	if len(settings.IPAllowlist) != 1 {
		t.Fatalf("IPAllowlist length = %d, want 1", len(settings.IPAllowlist))
	}
}

func TestIPAllowlistService_UpdateRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/ip_allowlist", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		w.Header().Set("X-Request-Id", "req-ip-update")
		fmt.Fprint(w, `{"type":"ip_allowlist","enabled":true,"ip_allowlist":["10.0.0.0/8"]}`)
	})

	ctx := context.Background()
	result, err := client.IPAllowlist.UpdateRaw(ctx, &UpdateIPAllowlistRequest{
		Enabled:     true,
		IPAllowlist: []string{"10.0.0.0/8"},
	})
	if err != nil {
		t.Fatalf("UpdateRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	settings, err := ParseIPAllowlistUpdateResult(result)
	if err != nil {
		t.Fatalf("ParseIPAllowlistUpdateResult returned error: %v", err)
	}
	if !settings.Enabled {
		t.Error("Enabled = false, want true")
	}
}
