package intercom

import (
	"context"
	"fmt"
	"net/http"
	"testing"
)

func TestSubscriptionTypesService_List(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

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
	result, err := client.SubscriptionTypes.List(ctx)
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

func TestSubscriptionTypesService_ListRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/subscription_types", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-sub-list")
		fmt.Fprint(w, `{"type":"list","data":[{"type":"subscription","id":"135","state":"live","consent_type":"opt_out"}]}`)
	})

	ctx := context.Background()
	result, err := client.SubscriptionTypes.ListRaw(ctx)
	if err != nil {
		t.Fatalf("ListRaw returned error: %v", err)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-sub-list" {
		t.Errorf("Header X-Request-Id = %q, want req-sub-list", result.Header.Get("X-Request-Id"))
	}
	stl, err := ParseSubscriptionTypeListResult(result)
	if err != nil {
		t.Fatalf("ParseSubscriptionTypeListResult returned error: %v", err)
	}
	if len(stl.Data) != 1 {
		t.Fatalf("Data length = %d, want 1", len(stl.Data))
	}
	if stl.Data[0].ID != "135" {
		t.Errorf("Data[0].ID = %v, want 135", stl.Data[0].ID)
	}
}

func TestSubscriptionTypesService_List_Unauthorized(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/subscription_types", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		fmt.Fprint(w, `{
			"type":"error.list",
			"request_id":"req-789",
			"errors":[{"code":"unauthorized","message":"Access Token Invalid"}]
		}`)
	})

	ctx := context.Background()
	_, err := client.SubscriptionTypes.List(ctx)
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !IsUnauthorized(err) {
		t.Errorf("Expected IsUnauthorized, got: %v", err)
	}
}
