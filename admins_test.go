package intercom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestAdminsService_Me(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/me", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testHeader(t, r, "Authorization", "Bearer test-token")
		fmt.Fprint(w, `{
			"type":"admin",
			"id":"12345",
			"name":"Joe Admin",
			"email":"joe@example.com",
			"job_title":"Support Lead",
			"away_mode_enabled":false,
			"away_mode_reassign":false,
			"has_inbox_seat":true,
			"team_ids":[814865],
			"avatar":{"type":"avatar","image_url":"https://example.com/avatar.png"},
			"email_verified":true,
			"app":{
				"type":"app",
				"id_code":"xyz789",
				"name":"ACME",
				"region":"US",
				"timezone":"America/Los_Angeles",
				"created_at":1671465577,
				"identity_verification":false
			}
		}`)
	})

	ctx := context.Background()
	admin, err := client.Admins.Me(ctx)
	if err != nil {
		t.Fatalf("Admins.Me returned error: %v", err)
	}
	if admin.ID != "12345" {
		t.Errorf("Admin.ID = %v, want 12345", admin.ID)
	}
	if admin.Name != "Joe Admin" {
		t.Errorf("Admin.Name = %v, want Joe Admin", admin.Name)
	}
	if admin.Email != "joe@example.com" {
		t.Errorf("Admin.Email = %v, want joe@example.com", admin.Email)
	}
	if admin.App == nil {
		t.Fatal("Admin.App is nil, expected non-nil")
	}
	if admin.App.IDCode != "xyz789" {
		t.Errorf("Admin.App.IDCode = %v, want xyz789", admin.App.IDCode)
	}
	if admin.App.Name != "ACME" {
		t.Errorf("Admin.App.Name = %v, want ACME", admin.App.Name)
	}
	if admin.App.Region != "US" {
		t.Errorf("Admin.App.Region = %v, want US", admin.App.Region)
	}
	if admin.EmailVerified == nil || !*admin.EmailVerified {
		t.Errorf("Admin.EmailVerified = %v, want true", admin.EmailVerified)
	}
}

func TestAdminsService_Get(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/admins/12345", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"admin",
			"id":"12345",
			"name":"Joe Admin",
			"email":"joe@example.com",
			"job_title":"Support Lead",
			"away_mode_enabled":false,
			"away_mode_reassign":false,
			"has_inbox_seat":true,
			"team_ids":[814865]
		}`)
	})

	ctx := context.Background()
	admin, err := client.Admins.Get(ctx, "12345")
	if err != nil {
		t.Fatalf("Admins.Get returned error: %v", err)
	}
	if admin.ID != "12345" {
		t.Errorf("Admin.ID = %v, want 12345", admin.ID)
	}
	if admin.Name != "Joe Admin" {
		t.Errorf("Admin.Name = %v, want Joe Admin", admin.Name)
	}
	if admin.JobTitle != "Support Lead" {
		t.Errorf("Admin.JobTitle = %v, want Support Lead", admin.JobTitle)
	}
	if !admin.HasInboxSeat {
		t.Errorf("Admin.HasInboxSeat = false, want true")
	}
	if len(admin.TeamIDs) != 1 || admin.TeamIDs[0] != 814865 {
		t.Errorf("Admin.TeamIDs = %v, want [814865]", admin.TeamIDs)
	}
}

func TestAdminsService_Get_NotFound(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/admins/nonexistent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{
			"type":"error.list",
			"request_id":"req-123",
			"errors":[{"code":"admin_not_found","message":"Admin not found"}]
		}`)
	})

	ctx := context.Background()
	_, err := client.Admins.Get(ctx, "nonexistent")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("Expected IsNotFound, got: %v", err)
	}
}

func TestAdminsService_List(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/admins", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"admin.list",
			"admins":[
				{
					"type":"admin",
					"id":"1",
					"name":"Admin One",
					"email":"admin1@example.com",
					"away_mode_enabled":false,
					"away_mode_reassign":false,
					"has_inbox_seat":true
				},
				{
					"type":"admin",
					"id":"2",
					"name":"Admin Two",
					"email":"admin2@example.com",
					"away_mode_enabled":true,
					"away_mode_reassign":true,
					"has_inbox_seat":false
				}
			]
		}`)
	})

	ctx := context.Background()
	list, err := client.Admins.List(ctx)
	if err != nil {
		t.Fatalf("Admins.List returned error: %v", err)
	}
	if list.Type != "admin.list" {
		t.Errorf("AdminList.Type = %v, want admin.list", list.Type)
	}
	if len(list.Admins) != 2 {
		t.Fatalf("AdminList.Admins length = %d, want 2", len(list.Admins))
	}
	if list.Admins[0].Name != "Admin One" {
		t.Errorf("Admins[0].Name = %v, want Admin One", list.Admins[0].Name)
	}
	if list.Admins[1].Name != "Admin Two" {
		t.Errorf("Admins[1].Name = %v, want Admin Two", list.Admins[1].Name)
	}
	if !list.Admins[1].AwayModeEnabled {
		t.Errorf("Admins[1].AwayModeEnabled = false, want true")
	}
}

func TestAdminsService_SetAway(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/admins/12345/away", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		var body SetAwayRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if !body.AwayModeEnabled {
			t.Errorf("SetAway body away_mode_enabled = false, want true")
		}
		if !body.AwayModeReassign {
			t.Errorf("SetAway body away_mode_reassign = false, want true")
		}
		fmt.Fprint(w, `{
			"type":"admin",
			"id":"12345",
			"name":"Joe Admin",
			"email":"joe@example.com",
			"away_mode_enabled":true,
			"away_mode_reassign":true,
			"away_status_reason_id":12345
		}`)
	})

	ctx := context.Background()
	req := &SetAwayRequest{
		AwayModeEnabled:  true,
		AwayModeReassign: true,
	}
	admin, err := client.Admins.SetAway(ctx, "12345", req)
	if err != nil {
		t.Fatalf("Admins.SetAway returned error: %v", err)
	}
	if !admin.AwayModeEnabled {
		t.Errorf("Admin.AwayModeEnabled = false, want true")
	}
	if !admin.AwayModeReassign {
		t.Errorf("Admin.AwayModeReassign = false, want true")
	}
}

func TestAdminsService_SetAway_WithReason(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/admins/12345/away", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body["away_status_reason_id"] != float64(99) {
			t.Errorf("SetAway body away_status_reason_id = %v, want 99", body["away_status_reason_id"])
		}
		fmt.Fprint(w, `{
			"type":"admin",
			"id":"12345",
			"away_mode_enabled":true,
			"away_mode_reassign":true,
			"away_status_reason_id":99
		}`)
	})

	ctx := context.Background()
	reasonID := 99
	req := &SetAwayRequest{
		AwayModeEnabled:    true,
		AwayModeReassign:   true,
		AwayStatusReasonID: &reasonID,
	}
	admin, err := client.Admins.SetAway(ctx, "12345", req)
	if err != nil {
		t.Fatalf("Admins.SetAway returned error: %v", err)
	}
	if admin.AwayStatusReasonID == nil || *admin.AwayStatusReasonID != 99 {
		t.Errorf("Admin.AwayStatusReasonID = %v, want 99", admin.AwayStatusReasonID)
	}
}

func TestAdminsService_ListActivityLogs(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/admins/activity_logs", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		if got := r.URL.Query().Get("created_at_after"); got != "1677253093" {
			t.Errorf("created_at_after = %v, want 1677253093", got)
		}
		if got := r.URL.Query().Get("created_at_before"); got != "1677861493" {
			t.Errorf("created_at_before = %v, want 1677861493", got)
		}
		fmt.Fprint(w, `{
			"type":"activity_log.list",
			"pages":{
				"type":"pages",
				"page":1,
				"per_page":20,
				"total_pages":1
			},
			"activity_logs":[
				{
					"id":"log-1",
					"performed_by":{
						"type":"admin",
						"id":"12345",
						"email":"joe@example.com",
						"ip":"198.51.100.255"
					},
					"metadata":{"before":"before","after":"after"},
					"created_at":1734537253,
					"activity_type":"app_name_change",
					"activity_description":"Admin changed app name from before to after."
				},
				{
					"id":"log-2",
					"performed_by":{
						"type":"admin",
						"id":"12345",
						"email":"joe@example.com",
						"ip":"198.51.100.255"
					},
					"created_at":1734537254,
					"activity_type":"admin_login_success",
					"activity_description":"Admin logged in."
				}
			]
		}`)
	})

	ctx := context.Background()
	opts := &ActivityLogListOptions{
		CreatedAtAfter:  "1677253093",
		CreatedAtBefore: "1677861493",
	}
	list, err := client.Admins.ListActivityLogs(ctx, opts)
	if err != nil {
		t.Fatalf("Admins.ListActivityLogs returned error: %v", err)
	}
	if list.Type != "activity_log.list" {
		t.Errorf("ActivityLogList.Type = %v, want activity_log.list", list.Type)
	}
	if len(list.ActivityLogs) != 2 {
		t.Fatalf("ActivityLogList.ActivityLogs length = %d, want 2", len(list.ActivityLogs))
	}
	log := list.ActivityLogs[0]
	if log.ID != "log-1" {
		t.Errorf("ActivityLog.ID = %v, want log-1", log.ID)
	}
	if log.ActivityType != "app_name_change" {
		t.Errorf("ActivityLog.ActivityType = %v, want app_name_change", log.ActivityType)
	}
	if log.PerformedBy.Email != "joe@example.com" {
		t.Errorf("ActivityLog.PerformedBy.Email = %v, want joe@example.com", log.PerformedBy.Email)
	}
	if log.PerformedBy.IP != "198.51.100.255" {
		t.Errorf("ActivityLog.PerformedBy.IP = %v, want 198.51.100.255", log.PerformedBy.IP)
	}
}

func TestAdminsService_ListActivityLogs_OnlyRequired(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/admins/activity_logs", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		if got := r.URL.Query().Get("created_at_after"); got != "1677253093" {
			t.Errorf("created_at_after = %v, want 1677253093", got)
		}
		if got := r.URL.Query().Get("created_at_before"); got != "" {
			t.Errorf("created_at_before should be empty, got %v", got)
		}
		fmt.Fprint(w, `{
			"type":"activity_log.list",
			"pages":{"type":"pages","page":1,"per_page":20,"total_pages":1},
			"activity_logs":[]
		}`)
	})

	ctx := context.Background()
	opts := &ActivityLogListOptions{
		CreatedAtAfter: "1677253093",
	}
	list, err := client.Admins.ListActivityLogs(ctx, opts)
	if err != nil {
		t.Fatalf("Admins.ListActivityLogs returned error: %v", err)
	}
	if len(list.ActivityLogs) != 0 {
		t.Errorf("Expected empty activity logs, got %d", len(list.ActivityLogs))
	}
}

// --- Raw companion method tests ---

func TestAdminsService_MeRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/me", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-me")
		fmt.Fprint(w, `{"type":"admin","id":"12345","name":"Joe Admin","email":"joe@example.com"}`)
	})

	ctx := context.Background()
	result, err := client.Admins.MeRaw(ctx)
	if err != nil {
		t.Fatalf("MeRaw returned error: %v", err)
	}
	admin, err := ParseAdminMeResult(result)
	if err != nil {
		t.Fatalf("ParseAdminMeResult returned error: %v", err)
	}
	if admin.ID != "12345" {
		t.Errorf("Admin.ID = %v, want 12345", admin.ID)
	}
	if admin.Name != "Joe Admin" {
		t.Errorf("Admin.Name = %v, want Joe Admin", admin.Name)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-me" {
		t.Errorf("Header X-Request-Id = %q, want req-me", result.Header.Get("X-Request-Id"))
	}
	if result.Error != nil {
		t.Errorf("Result.Error = %v, want nil", result.Error)
	}
}

func TestAdminsService_GetRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/admins/12345", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-get")
		fmt.Fprint(w, `{"type":"admin","id":"12345","name":"Joe Admin","email":"joe@example.com"}`)
	})

	ctx := context.Background()
	result, err := client.Admins.GetRaw(ctx, "12345")
	if err != nil {
		t.Fatalf("GetRaw returned error: %v", err)
	}
	admin, err := ParseAdminGetResult(result)
	if err != nil {
		t.Fatalf("ParseAdminGetResult returned error: %v", err)
	}
	if admin.ID != "12345" {
		t.Errorf("Admin.ID = %v, want 12345", admin.ID)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-get" {
		t.Errorf("Header X-Request-Id = %q, want req-get", result.Header.Get("X-Request-Id"))
	}
	if result.Error != nil {
		t.Errorf("Result.Error = %v, want nil", result.Error)
	}
}

func TestAdminsService_GetRaw_NotFound(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/admins/nonexistent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"admin_not_found","message":"Admin not found"}]}`)
	})

	ctx := context.Background()
	result, err := client.Admins.GetRaw(ctx, "nonexistent")
	if err != nil {
		t.Fatalf("GetRaw returned Go error: %v", err)
	}
	if result.Error == nil {
		t.Fatal("Result.Error is nil, want non-nil")
	}
	if result.Error.Code != "admin_not_found" {
		t.Errorf("Error.Code = %q, want admin_not_found", result.Error.Code)
	}
	if result.StatusCode != 404 {
		t.Errorf("StatusCode = %d, want 404", result.StatusCode)
	}
}

func TestAdminsService_ListRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/admins", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-list")
		fmt.Fprint(w, `{"type":"admin.list","admins":[{"type":"admin","id":"1","name":"Admin One"},{"type":"admin","id":"2","name":"Admin Two"}]}`)
	})

	ctx := context.Background()
	result, err := client.Admins.ListRaw(ctx)
	if err != nil {
		t.Fatalf("ListRaw returned error: %v", err)
	}
	list, err := ParseAdminListResult(result)
	if err != nil {
		t.Fatalf("ParseAdminListResult returned error: %v", err)
	}
	if len(list.Admins) != 2 {
		t.Fatalf("Admins length = %d, want 2", len(list.Admins))
	}
	if list.Admins[0].Name != "Admin One" {
		t.Errorf("Admins[0].Name = %v, want Admin One", list.Admins[0].Name)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-list" {
		t.Errorf("Header X-Request-Id = %q, want req-list", result.Header.Get("X-Request-Id"))
	}
}

func TestAdminsService_SetAwayRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/admins/12345/away", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		w.Header().Set("X-Request-Id", "req-away")
		var body SetAwayRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if !body.AwayModeEnabled {
			t.Errorf("SetAway body away_mode_enabled = false, want true")
		}
		fmt.Fprint(w, `{"type":"admin","id":"12345","away_mode_enabled":true,"away_mode_reassign":true}`)
	})

	ctx := context.Background()
	req := &SetAwayRequest{
		AwayModeEnabled:  true,
		AwayModeReassign: true,
	}
	result, err := client.Admins.SetAwayRaw(ctx, "12345", req)
	if err != nil {
		t.Fatalf("SetAwayRaw returned error: %v", err)
	}
	admin, err := ParseAdminSetAwayResult(result)
	if err != nil {
		t.Fatalf("ParseAdminSetAwayResult returned error: %v", err)
	}
	if !admin.AwayModeEnabled {
		t.Errorf("Admin.AwayModeEnabled = false, want true")
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-away" {
		t.Errorf("Header X-Request-Id = %q, want req-away", result.Header.Get("X-Request-Id"))
	}
}

func TestAdminsService_ListActivityLogsRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/admins/activity_logs", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-logs")
		if got := r.URL.Query().Get("created_at_after"); got != "1677253093" {
			t.Errorf("created_at_after = %v, want 1677253093", got)
		}
		fmt.Fprint(w, `{"type":"activity_log.list","activity_logs":[{"id":"log-1","performed_by":{"type":"admin","id":"12345","email":"joe@example.com"},"created_at":1734537253,"activity_type":"app_name_change"}]}`)
	})

	ctx := context.Background()
	opts := &ActivityLogListOptions{
		CreatedAtAfter: "1677253093",
	}
	result, err := client.Admins.ListActivityLogsRaw(ctx, opts)
	if err != nil {
		t.Fatalf("ListActivityLogsRaw returned error: %v", err)
	}
	logList, err := ParseAdminListActivityLogsResult(result)
	if err != nil {
		t.Fatalf("ParseAdminListActivityLogsResult returned error: %v", err)
	}
	if len(logList.ActivityLogs) != 1 {
		t.Fatalf("ActivityLogs length = %d, want 1", len(logList.ActivityLogs))
	}
	if logList.ActivityLogs[0].ID != "log-1" {
		t.Errorf("ActivityLogs[0].ID = %v, want log-1", logList.ActivityLogs[0].ID)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-logs" {
		t.Errorf("Header X-Request-Id = %q, want req-logs", result.Header.Get("X-Request-Id"))
	}
}
