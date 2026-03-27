package intercom

import (
	"context"
	"fmt"
	"net/http"
	"testing"
)

func TestAwayStatusReasonsService_List(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/away_status_reasons", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testHeader(t, r, "Authorization", "Bearer test-token")
		fmt.Fprint(w, `[
			{
				"type":"away_status_reason",
				"id":"asr_1",
				"label":"On a break",
				"emoji":"☕",
				"order":1,
				"deleted":false,
				"created_at":1700000000,
				"updated_at":1700000100
			},
			{
				"type":"away_status_reason",
				"id":"asr_2",
				"label":"In a meeting",
				"emoji":"📅",
				"order":2,
				"deleted":false,
				"created_at":1700000000,
				"updated_at":1700000100
			},
			{
				"type":"away_status_reason",
				"id":"asr_3",
				"label":"Old reason",
				"emoji":"🗑️",
				"order":3,
				"deleted":true,
				"created_at":1699000000,
				"updated_at":1700000000
			}
		]`)
	})

	ctx := context.Background()
	reasons, err := client.AwayStatusReasons.List(ctx)
	if err != nil {
		t.Fatalf("AwayStatusReasons.List returned error: %v", err)
	}
	if len(reasons) != 3 {
		t.Fatalf("length = %d, want 3", len(reasons))
	}
	if reasons[0].ID != "asr_1" {
		t.Errorf("reasons[0].ID = %v, want asr_1", reasons[0].ID)
	}
	if reasons[0].Label != "On a break" {
		t.Errorf("reasons[0].Label = %v, want On a break", reasons[0].Label)
	}
	if reasons[0].Emoji != "☕" {
		t.Errorf("reasons[0].Emoji = %v, want ☕", reasons[0].Emoji)
	}
	if reasons[0].Order != 1 {
		t.Errorf("reasons[0].Order = %v, want 1", reasons[0].Order)
	}
	if reasons[0].Deleted {
		t.Error("reasons[0].Deleted = true, want false")
	}
	if !reasons[2].Deleted {
		t.Error("reasons[2].Deleted = false, want true")
	}
}

// --- Raw companion method tests ---

func TestAwayStatusReasonsService_ListRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/away_status_reasons", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-asr")
		fmt.Fprint(w, `[
			{"type":"away_status_reason","id":"asr_1","label":"On a break","emoji":"☕","order":1},
			{"type":"away_status_reason","id":"asr_2","label":"In a meeting","emoji":"📅","order":2}
		]`)
	})

	ctx := context.Background()
	result, err := client.AwayStatusReasons.ListRaw(ctx)
	if err != nil {
		t.Fatalf("ListRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-asr" {
		t.Errorf("Header X-Request-Id = %q, want req-asr", result.Header.Get("X-Request-Id"))
	}
	if result.Error != nil {
		t.Errorf("Result.Error = %v, want nil", result.Error)
	}
	reasons, err := ParseAwayStatusReasonListResult(result)
	if err != nil {
		t.Fatalf("ParseAwayStatusReasonListResult returned error: %v", err)
	}
	if len(reasons) != 2 {
		t.Fatalf("Data length = %d, want 2", len(reasons))
	}
	if reasons[0].ID != "asr_1" {
		t.Errorf("Data[0].ID = %v, want asr_1", reasons[0].ID)
	}
}
