package intercom

import (
	"context"
	"fmt"
	"net/http"
	"testing"
)

func TestTeamsService_Get(t *testing.T) {
	client, mux, teardown := setup()
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
	team, err := client.Teams.Get(ctx, "123")
	if err != nil {
		t.Fatalf("Teams.Get returned error: %v", err)
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
	client, mux, teardown := setup()
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
	_, err := client.Teams.Get(ctx, "nonexistent")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("Expected IsNotFound, got: %v", err)
	}
}

func TestTeamsService_GetRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/teams/123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-team-get")
		fmt.Fprint(w, `{"type":"team","id":"123","name":"Sales Team","admin_ids":[493881]}`)
	})

	ctx := context.Background()
	result, err := client.Teams.GetRaw(ctx, "123")
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
	team, err := ParseTeamGetResult(result)
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
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/teams/nonexistent", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"not_found","message":"Team not found"}]}`)
	})

	ctx := context.Background()
	result, err := client.Teams.GetRaw(ctx, "nonexistent")
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
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/teams", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-team-list")
		fmt.Fprint(w, `{"type":"team.list","teams":[{"type":"team","id":"1","name":"Sales"},{"type":"team","id":"2","name":"Support"}]}`)
	})

	ctx := context.Background()
	result, err := client.Teams.ListRaw(ctx)
	if err != nil {
		t.Fatalf("ListRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-team-list" {
		t.Errorf("Header X-Request-Id = %q, want req-team-list", result.Header.Get("X-Request-Id"))
	}
	teams, err := ParseTeamListResult(result)
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
	client, mux, teardown := setup()
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
	result, err := client.Teams.List(ctx)
	if err != nil {
		t.Fatalf("Teams.List returned error: %v", err)
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
