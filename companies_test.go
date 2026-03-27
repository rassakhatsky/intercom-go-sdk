package intercom

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"
)

func TestCompaniesService_Get(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/companies/abc123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		testHeader(t, r, "Authorization", "Bearer test-token")
		fmt.Fprint(w, `{
			"type":"company",
			"id":"abc123",
			"company_id":"remote-1",
			"name":"Acme Inc",
			"size":100,
			"website":"https://acme.com",
			"industry":"Software",
			"monthly_spend":5000,
			"session_count":42,
			"user_count":10
		}`)
	})

	ctx := context.Background()
	company, err := client.Companies.Get(ctx, "abc123")
	if err != nil {
		t.Fatalf("Companies.Get returned error: %v", err)
	}
	if company.ID != "abc123" {
		t.Errorf("Company.ID = %v, want abc123", company.ID)
	}
	if company.Name != "Acme Inc" {
		t.Errorf("Company.Name = %v, want Acme Inc", company.Name)
	}
	if company.Size != 100 {
		t.Errorf("Company.Size = %v, want 100", company.Size)
	}
	if company.Website != "https://acme.com" {
		t.Errorf("Company.Website = %v, want https://acme.com", company.Website)
	}
}

func TestCompaniesService_Create(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/companies", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body CreateOrUpdateCompanyRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.CompanyID != "remote-1" {
			t.Errorf("Create body company_id = %v, want remote-1", body.CompanyID)
		}
		if body.Name != "Acme Inc" {
			t.Errorf("Create body name = %v, want Acme Inc", body.Name)
		}
		fmt.Fprint(w, `{"type":"company","id":"abc123","company_id":"remote-1","name":"Acme Inc"}`)
	})

	ctx := context.Background()
	company, err := client.Companies.Create(ctx, &CreateOrUpdateCompanyRequest{
		CompanyID: "remote-1",
		Name:      "Acme Inc",
	})
	if err != nil {
		t.Fatalf("Companies.Create returned error: %v", err)
	}
	if company.ID != "abc123" {
		t.Errorf("Company.ID = %v, want abc123", company.ID)
	}
	if company.CompanyID != "remote-1" {
		t.Errorf("Company.CompanyID = %v, want remote-1", company.CompanyID)
	}
}

func TestCompaniesService_Update(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/companies/abc123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		var body UpdateCompanyRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.Name != "Acme Updated" {
			t.Errorf("Update body name = %v, want Acme Updated", body.Name)
		}
		fmt.Fprint(w, `{"type":"company","id":"abc123","name":"Acme Updated"}`)
	})

	ctx := context.Background()
	company, err := client.Companies.Update(ctx, "abc123", &UpdateCompanyRequest{
		Name: "Acme Updated",
	})
	if err != nil {
		t.Fatalf("Companies.Update returned error: %v", err)
	}
	if company.Name != "Acme Updated" {
		t.Errorf("Company.Name = %v, want Acme Updated", company.Name)
	}
}

func TestCompaniesService_Delete(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/companies/abc123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{"id":"abc123","object":"company","deleted":true}`)
	})

	ctx := context.Background()
	deleted, err := client.Companies.Delete(ctx, "abc123")
	if err != nil {
		t.Fatalf("Companies.Delete returned error: %v", err)
	}
	if !deleted.Deleted {
		t.Error("CompanyDeleted.Deleted = false, want true")
	}
	if deleted.ID != "abc123" {
		t.Errorf("CompanyDeleted.ID = %v, want abc123", deleted.ID)
	}
	if deleted.Object != "company" {
		t.Errorf("CompanyDeleted.Object = %v, want company", deleted.Object)
	}
}

func TestCompaniesService_List(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/companies", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"list",
			"data":[
				{"type":"company","id":"c1","name":"Acme"},
				{"type":"company","id":"c2","name":"Globex"}
			],
			"total_count":2,
			"pages":{"type":"pages","page":1,"per_page":50,"total_pages":1}
		}`)
	})

	ctx := context.Background()
	result, err := client.Companies.List(ctx, nil)
	if err != nil {
		t.Fatalf("Companies.List returned error: %v", err)
	}
	if len(result.Data) != 2 {
		t.Errorf("List returned %d companies, want 2", len(result.Data))
	}
	if result.TotalCount != 2 {
		t.Errorf("TotalCount = %d, want 2", result.TotalCount)
	}
}

func TestCompaniesService_ListAll(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	callCount := 0
	mux.HandleFunc("/companies", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		callCount++
		if callCount == 1 {
			fmt.Fprint(w, `{
				"type":"list",
				"data":[{"type":"company","id":"c1"}],
				"total_count":2,
				"pages":{"type":"pages","page":1,"per_page":1,"total_pages":2,"next":{"per_page":1,"starting_after":"cursor1"}}
			}`)
		} else {
			fmt.Fprint(w, `{
				"type":"list",
				"data":[{"type":"company","id":"c2"}],
				"total_count":2,
				"pages":{"type":"pages","page":2,"per_page":1,"total_pages":2}
			}`)
		}
	})

	ctx := context.Background()
	iter := client.Companies.ListAll(ctx, &ListOptions{PerPage: 1})
	var ids []string
	for iter.Next() {
		ids = append(ids, iter.Current().ID)
	}
	if err := iter.Err(); err != nil {
		t.Fatalf("ListAll returned error: %v", err)
	}
	if len(ids) != 2 {
		t.Fatalf("ListAll returned %d items, want 2", len(ids))
	}
	if ids[0] != "c1" || ids[1] != "c2" {
		t.Errorf("ListAll ids = %v, want [c1 c2]", ids)
	}
}

func TestCompaniesService_Scroll(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/companies/scroll", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		scrollParam := r.URL.Query().Get("scroll_param")
		if scrollParam == "" {
			// First request - no scroll param
			fmt.Fprint(w, `{
				"type":"list",
				"data":[{"type":"company","id":"c1","name":"Acme"}],
				"scroll_param":"scroll-token-1",
				"total_count":2,
				"pages":{"type":"pages"}
			}`)
		} else if scrollParam == "scroll-token-1" {
			// Second request with scroll param
			fmt.Fprint(w, `{
				"type":"list",
				"data":[{"type":"company","id":"c2","name":"Globex"}],
				"scroll_param":"scroll-token-2",
				"total_count":2,
				"pages":{"type":"pages"}
			}`)
		} else {
			// End of scroll
			fmt.Fprint(w, `{
				"type":"list",
				"data":[],
				"total_count":2,
				"pages":{"type":"pages"}
			}`)
		}
	})

	ctx := context.Background()

	// First request with empty scroll param
	result, err := client.Companies.Scroll(ctx, "")
	if err != nil {
		t.Fatalf("Companies.Scroll (first) returned error: %v", err)
	}
	if len(result.Data) != 1 {
		t.Fatalf("Scroll returned %d companies, want 1", len(result.Data))
	}
	if result.Data[0].ID != "c1" {
		t.Errorf("Scroll first company ID = %v, want c1", result.Data[0].ID)
	}
	if result.ScrollParam != "scroll-token-1" {
		t.Errorf("ScrollParam = %v, want scroll-token-1", result.ScrollParam)
	}

	// Second request with scroll param
	result, err = client.Companies.Scroll(ctx, result.ScrollParam)
	if err != nil {
		t.Fatalf("Companies.Scroll (second) returned error: %v", err)
	}
	if len(result.Data) != 1 {
		t.Fatalf("Scroll returned %d companies, want 1", len(result.Data))
	}
	if result.Data[0].ID != "c2" {
		t.Errorf("Scroll second company ID = %v, want c2", result.Data[0].ID)
	}
}

func TestCompaniesService_Scroll_EmptyStart(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/companies/scroll", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		scrollParam := r.URL.Query().Get("scroll_param")
		if scrollParam != "" {
			t.Errorf("Expected empty scroll_param on first call, got %v", scrollParam)
		}
		fmt.Fprint(w, `{
			"type":"list",
			"data":[],
			"total_count":0,
			"pages":{"type":"pages"}
		}`)
	})

	ctx := context.Background()
	result, err := client.Companies.Scroll(ctx, "")
	if err != nil {
		t.Fatalf("Companies.Scroll returned error: %v", err)
	}
	if len(result.Data) != 0 {
		t.Errorf("Scroll returned %d companies, want 0", len(result.Data))
	}
}

func TestCompaniesService_ListContacts(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/companies/abc123/contacts", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"list",
			"data":[{"type":"contact","id":"ct1","email":"joe@example.com"}],
			"total_count":1,
			"pages":{"type":"pages","page":1,"per_page":50,"total_pages":1}
		}`)
	})

	ctx := context.Background()
	result, err := client.Companies.ListContacts(ctx, "abc123", nil)
	if err != nil {
		t.Fatalf("Companies.ListContacts returned error: %v", err)
	}
	if len(result.Data) != 1 {
		t.Errorf("ListContacts returned %d contacts, want 1", len(result.Data))
	}
	if result.Data[0].Email != "joe@example.com" {
		t.Errorf("Contact.Email = %v, want joe@example.com", result.Data[0].Email)
	}
}

func TestCompaniesService_ListSegments(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/companies/abc123/segments", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"list",
			"data":[{"type":"segment","id":"s1","name":"Active"}]
		}`)
	})

	ctx := context.Background()
	result, err := client.Companies.ListSegments(ctx, "abc123")
	if err != nil {
		t.Fatalf("Companies.ListSegments returned error: %v", err)
	}
	if len(result.Data) != 1 {
		t.Errorf("ListSegments returned %d segments, want 1", len(result.Data))
	}
	if result.Data[0].Name != "Active" {
		t.Errorf("Segment.Name = %v, want Active", result.Data[0].Name)
	}
}

func TestCompaniesService_ListNotes(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/companies/abc123/notes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"list",
			"data":[{"type":"note","id":"n1","body":"<p>Company note</p>","created_at":1674589321}],
			"total_count":1,
			"pages":{"type":"pages","page":1,"per_page":50,"total_pages":1}
		}`)
	})

	ctx := context.Background()
	result, err := client.Companies.ListNotes(ctx, "abc123")
	if err != nil {
		t.Fatalf("Companies.ListNotes returned error: %v", err)
	}
	if len(result.Data) != 1 {
		t.Errorf("ListNotes returned %d notes, want 1", len(result.Data))
	}
	if result.Data[0].Body != "<p>Company note</p>" {
		t.Errorf("Note.Body = %v, want <p>Company note</p>", result.Data[0].Body)
	}
}

func TestCompaniesService_CompanyList(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/companies/list", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body map[string]any
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("failed to decode request body: %v", err)
		}
		if body["page"] != float64(1) {
			t.Errorf("page = %v, want 1", body["page"])
		}
		if body["per_page"] != float64(15) {
			t.Errorf("per_page = %v, want 15", body["per_page"])
		}
		fmt.Fprint(w, `{
			"type":"list",
			"data":[
				{"type":"company","id":"c1","name":"Acme"},
				{"type":"company","id":"c2","name":"Globex"}
			],
			"total_count":2,
			"pages":{"type":"pages","page":1,"per_page":15,"total_pages":1}
		}`)
	})

	ctx := context.Background()
	result, err := client.Companies.CompanyList(ctx, &CompanyListOptions{
		Page:    1,
		PerPage: 15,
	})
	if err != nil {
		t.Fatalf("Companies.CompanyList returned error: %v", err)
	}
	if len(result.Data) != 2 {
		t.Errorf("CompanyList returned %d companies, want 2", len(result.Data))
	}
	if result.TotalCount != 2 {
		t.Errorf("TotalCount = %d, want 2", result.TotalCount)
	}
}

func TestCompaniesService_Get_NotFound(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/companies/999", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"company_not_found","message":"Company Not Found"}]}`)
	})

	ctx := context.Background()
	_, err := client.Companies.Get(ctx, "999")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !IsNotFound(err) {
		t.Errorf("IsNotFound = false, want true")
	}
}

// --- Raw companion method tests ---

func TestCompaniesService_GetRaw_Success(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/companies/abc123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-abc")
		fmt.Fprint(w, `{"type":"company","id":"abc123","name":"Acme Inc"}`)
	})

	ctx := context.Background()
	result, err := client.Companies.GetRaw(ctx, "abc123")
	if err != nil {
		t.Fatalf("GetRaw returned error: %v", err)
	}
	company, err := ParseCompanyGetResult(result)
	if err != nil {
		t.Fatalf("ParseCompanyGetResult returned error: %v", err)
	}
	if company.ID != "abc123" {
		t.Errorf("Company.ID = %v, want abc123", company.ID)
	}
	if company.Name != "Acme Inc" {
		t.Errorf("Company.Name = %v, want Acme Inc", company.Name)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-Request-Id") != "req-abc" {
		t.Errorf("Header X-Request-Id = %q, want req-abc", result.Header.Get("X-Request-Id"))
	}
	if result.Error != nil {
		t.Errorf("Result.Error = %v, want nil", result.Error)
	}
}

func TestCompaniesService_GetRaw_NotFound(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/companies/999", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"company_not_found","message":"Company Not Found"}]}`)
	})

	ctx := context.Background()
	result, err := client.Companies.GetRaw(ctx, "999")
	if err != nil {
		t.Fatalf("GetRaw returned Go error: %v", err)
	}
	if result.Error == nil {
		t.Fatal("Result.Error is nil, want non-nil")
	}
	if result.Error.Code != "company_not_found" {
		t.Errorf("Error.Code = %q, want company_not_found", result.Error.Code)
	}
	if result.StatusCode != 404 {
		t.Errorf("StatusCode = %d, want 404", result.StatusCode)
	}
}

func TestCompaniesService_ListRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/companies", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"list",
			"data":[{"type":"company","id":"c1"},{"type":"company","id":"c2"}],
			"total_count":2,
			"pages":{"type":"pages","page":1,"per_page":50,"total_pages":1}
		}`)
	})

	ctx := context.Background()
	result, err := client.Companies.ListRaw(ctx, nil)
	if err != nil {
		t.Fatalf("ListRaw returned error: %v", err)
	}
	paged, err := ParseCompanyListResult(result)
	if err != nil {
		t.Fatalf("ParseCompanyListResult returned error: %v", err)
	}
	if len(paged.Data) != 2 {
		t.Errorf("Data length = %d, want 2", len(paged.Data))
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestCompaniesService_CompanyListRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/companies/list", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{
			"type":"list",
			"data":[{"type":"company","id":"c1"}],
			"total_count":1,
			"pages":{"type":"pages","page":1,"per_page":15,"total_pages":1}
		}`)
	})

	ctx := context.Background()
	result, err := client.Companies.CompanyListRaw(ctx, &CompanyListOptions{Page: 1, PerPage: 15})
	if err != nil {
		t.Fatalf("CompanyListRaw returned error: %v", err)
	}
	paged, err := ParseCompanyCompanyListResult(result)
	if err != nil {
		t.Fatalf("ParseCompanyCompanyListResult returned error: %v", err)
	}
	if len(paged.Data) != 1 {
		t.Errorf("Data length = %d, want 1", len(paged.Data))
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestCompaniesService_CreateRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/companies", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"company","id":"abc123","company_id":"remote-1","name":"Acme Inc"}`)
	})

	ctx := context.Background()
	result, err := client.Companies.CreateRaw(ctx, &CreateOrUpdateCompanyRequest{
		CompanyID: "remote-1",
		Name:      "Acme Inc",
	})
	if err != nil {
		t.Fatalf("CreateRaw returned error: %v", err)
	}
	company, err := ParseCompanyCreateResult(result)
	if err != nil {
		t.Fatalf("ParseCompanyCreateResult returned error: %v", err)
	}
	if company.ID != "abc123" {
		t.Errorf("Company.ID = %v, want abc123", company.ID)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestCompaniesService_UpdateRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/companies/abc123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `{"type":"company","id":"abc123","name":"Acme Updated"}`)
	})

	ctx := context.Background()
	result, err := client.Companies.UpdateRaw(ctx, "abc123", &UpdateCompanyRequest{Name: "Acme Updated"})
	if err != nil {
		t.Fatalf("UpdateRaw returned error: %v", err)
	}
	company, err := ParseCompanyUpdateResult(result)
	if err != nil {
		t.Fatalf("ParseCompanyUpdateResult returned error: %v", err)
	}
	if company.Name != "Acme Updated" {
		t.Errorf("Company.Name = %v, want Acme Updated", company.Name)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestCompaniesService_DeleteRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/companies/abc123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{"id":"abc123","object":"company","deleted":true}`)
	})

	ctx := context.Background()
	result, err := client.Companies.DeleteRaw(ctx, "abc123")
	if err != nil {
		t.Fatalf("DeleteRaw returned error: %v", err)
	}
	deleted, err := ParseCompanyDeleteResult(result)
	if err != nil {
		t.Fatalf("ParseCompanyDeleteResult returned error: %v", err)
	}
	if !deleted.Deleted {
		t.Error("Deleted = false, want true")
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestCompaniesService_ScrollRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/companies/scroll", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-RateLimit-Limit", "100")
		fmt.Fprint(w, `{
			"type":"list",
			"data":[{"type":"company","id":"c1","name":"Acme"}],
			"scroll_param":"scroll-token-1",
			"total_count":1,
			"pages":{"type":"pages"}
		}`)
	})

	ctx := context.Background()
	result, err := client.Companies.ScrollRaw(ctx, "")
	if err != nil {
		t.Fatalf("ScrollRaw returned error: %v", err)
	}
	scroll, err := ParseCompanyScrollResult(result)
	if err != nil {
		t.Fatalf("ParseCompanyScrollResult returned error: %v", err)
	}
	if len(scroll.Data) != 1 {
		t.Errorf("Data length = %d, want 1", len(scroll.Data))
	}
	if scroll.ScrollParam != "scroll-token-1" {
		t.Errorf("ScrollParam = %v, want scroll-token-1", scroll.ScrollParam)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
	if result.Header.Get("X-RateLimit-Limit") != "100" {
		t.Errorf("Header X-RateLimit-Limit = %q, want 100", result.Header.Get("X-RateLimit-Limit"))
	}
}

func TestCompaniesService_ListContactsRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/companies/abc123/contacts", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"list",
			"data":[{"type":"contact","id":"ct1","email":"joe@example.com"}],
			"total_count":1,
			"pages":{"type":"pages","page":1,"per_page":50,"total_pages":1}
		}`)
	})

	ctx := context.Background()
	result, err := client.Companies.ListContactsRaw(ctx, "abc123", nil)
	if err != nil {
		t.Fatalf("ListContactsRaw returned error: %v", err)
	}
	paged, err := ParseCompanyListContactsResult(result)
	if err != nil {
		t.Fatalf("ParseCompanyListContactsResult returned error: %v", err)
	}
	if len(paged.Data) != 1 {
		t.Errorf("Data length = %d, want 1", len(paged.Data))
	}
	if paged.Data[0].Email != "joe@example.com" {
		t.Errorf("Contact.Email = %v, want joe@example.com", paged.Data[0].Email)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestCompaniesService_ListSegmentsRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/companies/abc123/segments", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"list",
			"data":[{"type":"segment","id":"s1","name":"Active"}]
		}`)
	})

	ctx := context.Background()
	result, err := client.Companies.ListSegmentsRaw(ctx, "abc123")
	if err != nil {
		t.Fatalf("ListSegmentsRaw returned error: %v", err)
	}
	segments, err := ParseCompanyListSegmentsResult(result)
	if err != nil {
		t.Fatalf("ParseCompanyListSegmentsResult returned error: %v", err)
	}
	if len(segments.Data) != 1 {
		t.Errorf("Data length = %d, want 1", len(segments.Data))
	}
	if segments.Data[0].Name != "Active" {
		t.Errorf("Segment.Name = %v, want Active", segments.Data[0].Name)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestCompaniesService_ListNotesRaw(t *testing.T) {
	client, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/companies/abc123/notes", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"list",
			"data":[{"type":"note","id":"n1","body":"<p>Company note</p>","created_at":1674589321}],
			"total_count":1,
			"pages":{"type":"pages","page":1,"per_page":50,"total_pages":1}
		}`)
	})

	ctx := context.Background()
	result, err := client.Companies.ListNotesRaw(ctx, "abc123")
	if err != nil {
		t.Fatalf("ListNotesRaw returned error: %v", err)
	}
	notes, err := ParseCompanyListNotesResult(result)
	if err != nil {
		t.Fatalf("ParseCompanyListNotesResult returned error: %v", err)
	}
	if len(notes.Data) != 1 {
		t.Errorf("Data length = %d, want 1", len(notes.Data))
	}
	if notes.Data[0].Body != "<p>Company note</p>" {
		t.Errorf("Note.Body = %v, want <p>Company note</p>", notes.Data[0].Body)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}
