package admins_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rassakhatsky/intercom-go-sdk/admins"
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

func setupAdmins() (svc *admins.Service, mux *http.ServeMux, teardown func()) {
	mux = http.NewServeMux()
	server := httptest.NewServer(mux)
	caller := &testCaller{baseURL: server.URL, client: server.Client()}
	svc = admins.NewService(caller)
	return svc, mux, server.Close
}

func setupTeams() (svc *admins.TeamsService, mux *http.ServeMux, teardown func()) {
	mux = http.NewServeMux()
	server := httptest.NewServer(mux)
	caller := &testCaller{baseURL: server.URL, client: server.Client()}
	svc = admins.NewTeamsService(caller)
	return svc, mux, server.Close
}

func setupAway() (svc *admins.AwayStatusReasonsService, mux *http.ServeMux, teardown func()) {
	mux = http.NewServeMux()
	server := httptest.NewServer(mux)
	caller := &testCaller{baseURL: server.URL, client: server.Client()}
	svc = admins.NewAwayStatusReasonsService(caller)
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
		t.Errorf("Header %v = %q, want %q", header, got, want)
	}
}

// --- Admins Service Tests ---

func TestService_Me(t *testing.T) {
	svc, mux, teardown := setupAdmins()
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
	admin, err := svc.Me(ctx)
	if err != nil {
		t.Fatalf("Me returned error: %v", err)
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

func TestService_Get(t *testing.T) {
	svc, mux, teardown := setupAdmins()
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
	admin, err := svc.Get(ctx, "12345")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
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

func TestService_Get_NotFound(t *testing.T) {
	svc, mux, teardown := setupAdmins()
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
	_, err := svc.Get(ctx, "nonexistent")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !api.IsNotFound(err) {
		t.Errorf("Expected IsNotFound, got: %v", err)
	}
}

func TestService_List(t *testing.T) {
	svc, mux, teardown := setupAdmins()
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
	list, err := svc.List(ctx)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if list.Type != "admin.list" {
		t.Errorf("List.Type = %v, want admin.list", list.Type)
	}
	if len(list.Admins) != 2 {
		t.Fatalf("List.Admins length = %d, want 2", len(list.Admins))
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

func TestService_SetAway(t *testing.T) {
	svc, mux, teardown := setupAdmins()
	defer teardown()

	mux.HandleFunc("/admins/12345/away", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		var body admins.SetAwayRequest
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
	req := &admins.SetAwayRequest{
		AwayModeEnabled:  true,
		AwayModeReassign: true,
	}
	admin, err := svc.SetAway(ctx, "12345", req)
	if err != nil {
		t.Fatalf("SetAway returned error: %v", err)
	}
	if !admin.AwayModeEnabled {
		t.Errorf("Admin.AwayModeEnabled = false, want true")
	}
	if !admin.AwayModeReassign {
		t.Errorf("Admin.AwayModeReassign = false, want true")
	}
}

func TestService_SetAway_WithReason(t *testing.T) {
	svc, mux, teardown := setupAdmins()
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
	req := &admins.SetAwayRequest{
		AwayModeEnabled:    true,
		AwayModeReassign:   true,
		AwayStatusReasonID: &reasonID,
	}
	admin, err := svc.SetAway(ctx, "12345", req)
	if err != nil {
		t.Fatalf("SetAway returned error: %v", err)
	}
	if admin.AwayStatusReasonID == nil || *admin.AwayStatusReasonID != 99 {
		t.Errorf("Admin.AwayStatusReasonID = %v, want 99", admin.AwayStatusReasonID)
	}
}

func TestService_ListActivityLogs(t *testing.T) {
	svc, mux, teardown := setupAdmins()
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
	opts := &admins.ActivityLogListOptions{
		CreatedAtAfter:  "1677253093",
		CreatedAtBefore: "1677861493",
	}
	list, err := svc.ListActivityLogs(ctx, opts)
	if err != nil {
		t.Fatalf("ListActivityLogs returned error: %v", err)
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

func TestService_ListActivityLogs_OnlyRequired(t *testing.T) {
	svc, mux, teardown := setupAdmins()
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
	opts := &admins.ActivityLogListOptions{
		CreatedAtAfter: "1677253093",
	}
	list, err := svc.ListActivityLogs(ctx, opts)
	if err != nil {
		t.Fatalf("ListActivityLogs returned error: %v", err)
	}
	if len(list.ActivityLogs) != 0 {
		t.Errorf("Expected empty activity logs, got %d", len(list.ActivityLogs))
	}
}

// --- Raw companion method tests ---

func TestService_MeRaw_Success(t *testing.T) {
	svc, mux, teardown := setupAdmins()
	defer teardown()

	mux.HandleFunc("/me", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-me")
		fmt.Fprint(w, `{"type":"admin","id":"12345","name":"Joe Admin","email":"joe@example.com"}`)
	})

	ctx := context.Background()
	result, err := svc.MeRaw(ctx)
	if err != nil {
		t.Fatalf("MeRaw returned error: %v", err)
	}
	admin, err := admins.ParseMeResult(result)
	if err != nil {
		t.Fatalf("ParseMeResult returned error: %v", err)
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

func TestService_GetRaw_Success(t *testing.T) {
	svc, mux, teardown := setupAdmins()
	defer teardown()

	mux.HandleFunc("/admins/12345", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-get")
		fmt.Fprint(w, `{"type":"admin","id":"12345","name":"Joe Admin","email":"joe@example.com"}`)
	})

	ctx := context.Background()
	result, err := svc.GetRaw(ctx, "12345")
	if err != nil {
		t.Fatalf("GetRaw returned error: %v", err)
	}
	admin, err := admins.ParseGetResult(result)
	if err != nil {
		t.Fatalf("ParseGetResult returned error: %v", err)
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

func TestService_GetRaw_NotFound(t *testing.T) {
	svc, mux, teardown := setupAdmins()
	defer teardown()

	mux.HandleFunc("/admins/nonexistent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"admin_not_found","message":"Admin not found"}]}`)
	})

	ctx := context.Background()
	result, err := svc.GetRaw(ctx, "nonexistent")
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

func TestService_ListRaw_Success(t *testing.T) {
	svc, mux, teardown := setupAdmins()
	defer teardown()

	mux.HandleFunc("/admins", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-list")
		fmt.Fprint(w, `{"type":"admin.list","admins":[{"type":"admin","id":"1","name":"Admin One"},{"type":"admin","id":"2","name":"Admin Two"}]}`)
	})

	ctx := context.Background()
	result, err := svc.ListRaw(ctx)
	if err != nil {
		t.Fatalf("ListRaw returned error: %v", err)
	}
	list, err := admins.ParseListResult(result)
	if err != nil {
		t.Fatalf("ParseListResult returned error: %v", err)
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

func TestService_SetAwayRaw_Success(t *testing.T) {
	svc, mux, teardown := setupAdmins()
	defer teardown()

	mux.HandleFunc("/admins/12345/away", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		w.Header().Set("X-Request-Id", "req-away")
		var body admins.SetAwayRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if !body.AwayModeEnabled {
			t.Errorf("SetAway body away_mode_enabled = false, want true")
		}
		fmt.Fprint(w, `{"type":"admin","id":"12345","away_mode_enabled":true,"away_mode_reassign":true}`)
	})

	ctx := context.Background()
	req := &admins.SetAwayRequest{
		AwayModeEnabled:  true,
		AwayModeReassign: true,
	}
	result, err := svc.SetAwayRaw(ctx, "12345", req)
	if err != nil {
		t.Fatalf("SetAwayRaw returned error: %v", err)
	}
	admin, err := admins.ParseSetAwayResult(result)
	if err != nil {
		t.Fatalf("ParseSetAwayResult returned error: %v", err)
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

func TestService_ListActivityLogsRaw_Success(t *testing.T) {
	svc, mux, teardown := setupAdmins()
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
	opts := &admins.ActivityLogListOptions{
		CreatedAtAfter: "1677253093",
	}
	result, err := svc.ListActivityLogsRaw(ctx, opts)
	if err != nil {
		t.Fatalf("ListActivityLogsRaw returned error: %v", err)
	}
	logList, err := admins.ParseListActivityLogsResult(result)
	if err != nil {
		t.Fatalf("ParseListActivityLogsResult returned error: %v", err)
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

// --- Teams Service Tests ---

func TestTeamsService_Get(t *testing.T) {
	svc, mux, teardown := setupTeams()
	defer teardown()

	mux.HandleFunc("/teams/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testHeader(t, r, "Authorization", "Bearer test-token")
		fmt.Fprint(w, `{
			"type":"team",
			"id":"123",
			"name":"Sales Team",
			"admin_ids":[493881,12345],
			"assignment_limit":10,
			"distribution_method":"round_robin"
		}`)
	})

	ctx := context.Background()
	team, err := svc.Get(ctx, "123")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
	}
	if team.ID != "123" {
		t.Errorf("Team.ID = %v, want 123", team.ID)
	}
	if team.Name != "Sales Team" {
		t.Errorf("Team.Name = %v, want Sales Team", team.Name)
	}
	if team.Type != "team" {
		t.Errorf("Team.Type = %v, want team", team.Type)
	}
	if len(team.AdminIDs) != 2 {
		t.Fatalf("Team.AdminIDs length = %d, want 2", len(team.AdminIDs))
	}
	if team.AdminIDs[0] != 493881 {
		t.Errorf("Team.AdminIDs[0] = %v, want 493881", team.AdminIDs[0])
	}
	if team.AssignmentLimit == nil || *team.AssignmentLimit != 10 {
		t.Errorf("Team.AssignmentLimit = %v, want 10", team.AssignmentLimit)
	}
	if team.DistributionMethod == nil || *team.DistributionMethod != "round_robin" {
		t.Errorf("Team.DistributionMethod = %v, want round_robin", team.DistributionMethod)
	}
}

func TestTeamsService_Get_NotFound(t *testing.T) {
	svc, mux, teardown := setupTeams()
	defer teardown()

	mux.HandleFunc("/teams/nonexistent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{
			"type":"error.list",
			"request_id":"req-123",
			"errors":[{"code":"team_not_found","message":"Team not found"}]
		}`)
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

func TestTeamsService_GetRaw_Success(t *testing.T) {
	svc, mux, teardown := setupTeams()
	defer teardown()

	mux.HandleFunc("/teams/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-team-get")
		fmt.Fprint(w, `{"type":"team","id":"123","name":"Sales Team","admin_ids":[493881]}`)
	})

	ctx := context.Background()
	result, err := svc.GetRaw(ctx, "123")
	if err != nil {
		t.Fatalf("GetRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-team-get" {
		t.Errorf("Header X-Request-Id = %q, want req-team-get", result.Header.Get("X-Request-Id"))
	}
	if result.Error != nil {
		t.Errorf("Result.Error = %v, want nil", result.Error)
	}
	team, err := admins.ParseTeamGetResult(result)
	if err != nil {
		t.Fatalf("ParseTeamGetResult returned error: %v", err)
	}
	if team.ID != "123" {
		t.Errorf("ID = %v, want 123", team.ID)
	}
	if team.Name != "Sales Team" {
		t.Errorf("Name = %v, want Sales Team", team.Name)
	}
}

func TestTeamsService_GetRaw_NotFound(t *testing.T) {
	svc, mux, teardown := setupTeams()
	defer teardown()

	mux.HandleFunc("/teams/nonexistent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"Team not found"}]}`)
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

func TestTeamsService_ListRaw_Success(t *testing.T) {
	svc, mux, teardown := setupTeams()
	defer teardown()

	mux.HandleFunc("/teams", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-team-list")
		fmt.Fprint(w, `{"type":"team.list","teams":[{"type":"team","id":"1","name":"Sales"},{"type":"team","id":"2","name":"Support"}]}`)
	})

	ctx := context.Background()
	result, err := svc.ListRaw(ctx)
	if err != nil {
		t.Fatalf("ListRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-team-list" {
		t.Errorf("Header X-Request-Id = %q, want req-team-list", result.Header.Get("X-Request-Id"))
	}
	teams, err := admins.ParseTeamListResult(result)
	if err != nil {
		t.Fatalf("ParseTeamListResult returned error: %v", err)
	}
	if len(teams.Teams) != 2 {
		t.Fatalf("Teams length = %d, want 2", len(teams.Teams))
	}
	if teams.Teams[0].Name != "Sales" {
		t.Errorf("Teams[0].Name = %v, want Sales", teams.Teams[0].Name)
	}
}

func TestTeamsService_List(t *testing.T) {
	svc, mux, teardown := setupTeams()
	defer teardown()

	mux.HandleFunc("/teams", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"team.list",
			"teams":[
				{"type":"team","id":"1","name":"Sales","admin_ids":[100,200]},
				{"type":"team","id":"2","name":"Support","admin_ids":[300],"assignment_limit":5,"distribution_method":"load_balanced"}
			]
		}`)
	})

	ctx := context.Background()
	result, err := svc.List(ctx)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if result.Type != "team.list" {
		t.Errorf("Type = %v, want team.list", result.Type)
	}
	if len(result.Teams) != 2 {
		t.Fatalf("Teams length = %d, want 2", len(result.Teams))
	}
	if result.Teams[0].Name != "Sales" {
		t.Errorf("Teams[0].Name = %v, want Sales", result.Teams[0].Name)
	}
	if result.Teams[1].Name != "Support" {
		t.Errorf("Teams[1].Name = %v, want Support", result.Teams[1].Name)
	}
	if len(result.Teams[0].AdminIDs) != 2 {
		t.Errorf("Teams[0].AdminIDs length = %d, want 2", len(result.Teams[0].AdminIDs))
	}
}

// --- Away Status Reasons Service Tests ---

func TestAwayStatusReasonsService_List(t *testing.T) {
	svc, mux, teardown := setupAway()
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
	reasons, err := svc.List(ctx)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
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

func TestAwayStatusReasonsService_ListRaw_Success(t *testing.T) {
	svc, mux, teardown := setupAway()
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
	result, err := svc.ListRaw(ctx)
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
	reasons, err := admins.ParseAwayStatusReasonListResult(result)
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
