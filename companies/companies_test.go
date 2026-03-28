package companies_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rassakhatsky/intercom-go-sdk/companies"
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

func setup() (svc *companies.Service, mux *http.ServeMux, teardown func()) {
	mux = http.NewServeMux()
	server := httptest.NewServer(mux)
	caller := &testCaller{baseURL: server.URL, client: server.Client()}
	svc = companies.NewService(caller)
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

func TestService_Get(t *testing.T) {
	svc, mux, teardown := setup()
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
	company, err := svc.Get(ctx, "abc123")
	if err != nil {
		t.Fatalf("Get returned error: %v", err)
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

func TestService_Create(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/companies", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		var body companies.CreateOrUpdateRequest
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
	company, err := svc.Create(ctx, &companies.CreateOrUpdateRequest{
		CompanyID: "remote-1",
		Name:      "Acme Inc",
	})
	if err != nil {
		t.Fatalf("Create returned error: %v", err)
	}
	if company.ID != "abc123" {
		t.Errorf("Company.ID = %v, want abc123", company.ID)
	}
	if company.CompanyID != "remote-1" {
		t.Errorf("Company.CompanyID = %v, want remote-1", company.CompanyID)
	}
}

func TestService_Update(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/companies/abc123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		var body companies.UpdateRequest
		if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
			t.Fatalf("decode request body: %v", err)
		}
		if body.Name != "Acme Updated" {
			t.Errorf("Update body name = %v, want Acme Updated", body.Name)
		}
		fmt.Fprint(w, `{"type":"company","id":"abc123","name":"Acme Updated"}`)
	})

	ctx := context.Background()
	company, err := svc.Update(ctx, "abc123", &companies.UpdateRequest{
		Name: "Acme Updated",
	})
	if err != nil {
		t.Fatalf("Update returned error: %v", err)
	}
	if company.Name != "Acme Updated" {
		t.Errorf("Company.Name = %v, want Acme Updated", company.Name)
	}
}

func TestService_Delete(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/companies/abc123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{"id":"abc123","object":"company","deleted":true}`)
	})

	ctx := context.Background()
	deleted, err := svc.Delete(ctx, "abc123")
	if err != nil {
		t.Fatalf("Delete returned error: %v", err)
	}
	if !deleted.Deleted {
		t.Error("Deleted.Deleted = false, want true")
	}
	if deleted.ID != "abc123" {
		t.Errorf("Deleted.ID = %v, want abc123", deleted.ID)
	}
	if deleted.Object != "company" {
		t.Errorf("Deleted.Object = %v, want company", deleted.Object)
	}
}

func TestService_List(t *testing.T) {
	svc, mux, teardown := setup()
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
	result, err := svc.List(ctx, nil)
	if err != nil {
		t.Fatalf("List returned error: %v", err)
	}
	if len(result.Data) != 2 {
		t.Errorf("List returned %d companies, want 2", len(result.Data))
	}
	if result.TotalCount != 2 {
		t.Errorf("TotalCount = %d, want 2", result.TotalCount)
	}
}

func TestService_ListAll(t *testing.T) {
	svc, mux, teardown := setup()
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
	iter := svc.ListAll(ctx, &api.ListOptions{PerPage: 1})
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

func TestService_Scroll(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/companies/scroll", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		scrollParam := r.URL.Query().Get("scroll_param")
		if scrollParam == "" {
			fmt.Fprint(w, `{
				"type":"list",
				"data":[{"type":"company","id":"c1","name":"Acme"}],
				"scroll_param":"scroll-token-1",
				"total_count":2,
				"pages":{"type":"pages"}
			}`)
		} else if scrollParam == "scroll-token-1" {
			fmt.Fprint(w, `{
				"type":"list",
				"data":[{"type":"company","id":"c2","name":"Globex"}],
				"scroll_param":"scroll-token-2",
				"total_count":2,
				"pages":{"type":"pages"}
			}`)
		} else {
			fmt.Fprint(w, `{
				"type":"list",
				"data":[],
				"total_count":2,
				"pages":{"type":"pages"}
			}`)
		}
	})

	ctx := context.Background()

	result, err := svc.Scroll(ctx, "")
	if err != nil {
		t.Fatalf("Scroll (first) returned error: %v", err)
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

	result, err = svc.Scroll(ctx, result.ScrollParam)
	if err != nil {
		t.Fatalf("Scroll (second) returned error: %v", err)
	}
	if len(result.Data) != 1 {
		t.Fatalf("Scroll returned %d companies, want 1", len(result.Data))
	}
	if result.Data[0].ID != "c2" {
		t.Errorf("Scroll second company ID = %v, want c2", result.Data[0].ID)
	}
}

func TestService_Scroll_EmptyStart(t *testing.T) {
	svc, mux, teardown := setup()
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
	result, err := svc.Scroll(ctx, "")
	if err != nil {
		t.Fatalf("Scroll returned error: %v", err)
	}
	if len(result.Data) != 0 {
		t.Errorf("Scroll returned %d companies, want 0", len(result.Data))
	}
}

func TestService_ListContacts(t *testing.T) {
	svc, mux, teardown := setup()
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
	result, err := svc.ListContacts(ctx, "abc123", nil)
	if err != nil {
		t.Fatalf("ListContacts returned error: %v", err)
	}
	if len(result.Data) != 1 {
		t.Errorf("ListContacts returned %d contacts, want 1", len(result.Data))
	}
	if result.Data[0].Email != "joe@example.com" {
		t.Errorf("Contact.Email = %v, want joe@example.com", result.Data[0].Email)
	}
}

func TestService_ListSegments(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/companies/abc123/segments", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"list",
			"data":[{"type":"segment","id":"s1","name":"Active"}]
		}`)
	})

	ctx := context.Background()
	result, err := svc.ListSegments(ctx, "abc123")
	if err != nil {
		t.Fatalf("ListSegments returned error: %v", err)
	}
	if len(result.Data) != 1 {
		t.Errorf("ListSegments returned %d segments, want 1", len(result.Data))
	}
	if result.Data[0].Name != "Active" {
		t.Errorf("Segment.Name = %v, want Active", result.Data[0].Name)
	}
}

func TestService_ListNotes(t *testing.T) {
	svc, mux, teardown := setup()
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
	result, err := svc.ListNotes(ctx, "abc123")
	if err != nil {
		t.Fatalf("ListNotes returned error: %v", err)
	}
	if len(result.Data) != 1 {
		t.Errorf("ListNotes returned %d notes, want 1", len(result.Data))
	}
	if result.Data[0].Body != "<p>Company note</p>" {
		t.Errorf("Note.Body = %v, want <p>Company note</p>", result.Data[0].Body)
	}
}

func TestService_CompanyList(t *testing.T) {
	svc, mux, teardown := setup()
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
	result, err := svc.CompanyList(ctx, &companies.ListOptions{
		Page:    1,
		PerPage: 15,
	})
	if err != nil {
		t.Fatalf("CompanyList returned error: %v", err)
	}
	if len(result.Data) != 2 {
		t.Errorf("CompanyList returned %d companies, want 2", len(result.Data))
	}
	if result.TotalCount != 2 {
		t.Errorf("TotalCount = %d, want 2", result.TotalCount)
	}
}

func TestService_Get_NotFound(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/companies/999", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"company_not_found","message":"Company Not Found"}]}`)
	})

	ctx := context.Background()
	_, err := svc.Get(ctx, "999")
	if err == nil {
		t.Fatal("Expected error, got nil")
	}
	if !api.IsNotFound(err) {
		t.Errorf("IsNotFound = false, want true")
	}
}

// --- Raw companion method tests ---

func TestService_GetRaw_Success(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/companies/abc123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		w.Header().Set("X-Request-Id", "req-abc")
		fmt.Fprint(w, `{"type":"company","id":"abc123","name":"Acme Inc"}`)
	})

	ctx := context.Background()
	result, err := svc.GetRaw(ctx, "abc123")
	if err != nil {
		t.Fatalf("GetRaw returned error: %v", err)
	}
	company, err := companies.ParseGetResult(result)
	if err != nil {
		t.Fatalf("ParseGetResult returned error: %v", err)
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

func TestService_GetRaw_NotFound(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/companies/999", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		fmt.Fprint(w, `{"type":"error.list","request_id":"req-1","errors":[{"code":"company_not_found","message":"Company Not Found"}]}`)
	})

	ctx := context.Background()
	result, err := svc.GetRaw(ctx, "999")
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

func TestService_ListRaw(t *testing.T) {
	svc, mux, teardown := setup()
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
	result, err := svc.ListRaw(ctx, nil)
	if err != nil {
		t.Fatalf("ListRaw returned error: %v", err)
	}
	paged, err := companies.ParseListResult(result)
	if err != nil {
		t.Fatalf("ParseListResult returned error: %v", err)
	}
	if len(paged.Data) != 2 {
		t.Errorf("Data length = %d, want 2", len(paged.Data))
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestService_CompanyListRaw(t *testing.T) {
	svc, mux, teardown := setup()
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
	result, err := svc.CompanyListRaw(ctx, &companies.ListOptions{Page: 1, PerPage: 15})
	if err != nil {
		t.Fatalf("CompanyListRaw returned error: %v", err)
	}
	paged, err := companies.ParseCompanyListResult(result)
	if err != nil {
		t.Fatalf("ParseCompanyListResult returned error: %v", err)
	}
	if len(paged.Data) != 1 {
		t.Errorf("Data length = %d, want 1", len(paged.Data))
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestService_CreateRaw(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/companies", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPost)
		fmt.Fprint(w, `{"type":"company","id":"abc123","company_id":"remote-1","name":"Acme Inc"}`)
	})

	ctx := context.Background()
	result, err := svc.CreateRaw(ctx, &companies.CreateOrUpdateRequest{
		CompanyID: "remote-1",
		Name:      "Acme Inc",
	})
	if err != nil {
		t.Fatalf("CreateRaw returned error: %v", err)
	}
	company, err := companies.ParseCreateResult(result)
	if err != nil {
		t.Fatalf("ParseCreateResult returned error: %v", err)
	}
	if company.ID != "abc123" {
		t.Errorf("Company.ID = %v, want abc123", company.ID)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestService_UpdateRaw(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/companies/abc123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodPut)
		fmt.Fprint(w, `{"type":"company","id":"abc123","name":"Acme Updated"}`)
	})

	ctx := context.Background()
	result, err := svc.UpdateRaw(ctx, "abc123", &companies.UpdateRequest{Name: "Acme Updated"})
	if err != nil {
		t.Fatalf("UpdateRaw returned error: %v", err)
	}
	company, err := companies.ParseUpdateResult(result)
	if err != nil {
		t.Fatalf("ParseUpdateResult returned error: %v", err)
	}
	if company.Name != "Acme Updated" {
		t.Errorf("Company.Name = %v, want Acme Updated", company.Name)
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestService_DeleteRaw(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/companies/abc123", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodDelete)
		fmt.Fprint(w, `{"id":"abc123","object":"company","deleted":true}`)
	})

	ctx := context.Background()
	result, err := svc.DeleteRaw(ctx, "abc123")
	if err != nil {
		t.Fatalf("DeleteRaw returned error: %v", err)
	}
	deleted, err := companies.ParseDeleteResult(result)
	if err != nil {
		t.Fatalf("ParseDeleteResult returned error: %v", err)
	}
	if !deleted.Deleted {
		t.Error("Deleted = false, want true")
	}
	if result.StatusCode != 200 {
		t.Errorf("StatusCode = %d, want 200", result.StatusCode)
	}
}

func TestService_ScrollRaw(t *testing.T) {
	svc, mux, teardown := setup()
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
	result, err := svc.ScrollRaw(ctx, "")
	if err != nil {
		t.Fatalf("ScrollRaw returned error: %v", err)
	}
	scroll, err := companies.ParseScrollResult(result)
	if err != nil {
		t.Fatalf("ParseScrollResult returned error: %v", err)
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

func TestService_ListContactsRaw(t *testing.T) {
	svc, mux, teardown := setup()
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
	result, err := svc.ListContactsRaw(ctx, "abc123", nil)
	if err != nil {
		t.Fatalf("ListContactsRaw returned error: %v", err)
	}
	paged, err := companies.ParseListContactsResult(result)
	if err != nil {
		t.Fatalf("ParseListContactsResult returned error: %v", err)
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

func TestService_ListSegmentsRaw(t *testing.T) {
	svc, mux, teardown := setup()
	defer teardown()

	mux.HandleFunc("/companies/abc123/segments", func(w http.ResponseWriter, r *http.Request) {
		testMethod(t, r, http.MethodGet)
		fmt.Fprint(w, `{
			"type":"list",
			"data":[{"type":"segment","id":"s1","name":"Active"}]
		}`)
	})

	ctx := context.Background()
	result, err := svc.ListSegmentsRaw(ctx, "abc123")
	if err != nil {
		t.Fatalf("ListSegmentsRaw returned error: %v", err)
	}
	segments, err := companies.ParseListSegmentsResult(result)
	if err != nil {
		t.Fatalf("ParseListSegmentsResult returned error: %v", err)
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

func TestService_ListNotesRaw(t *testing.T) {
	svc, mux, teardown := setup()
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
	result, err := svc.ListNotesRaw(ctx, "abc123")
	if err != nil {
		t.Fatalf("ListNotesRaw returned error: %v", err)
	}
	notes, err := companies.ParseListNotesResult(result)
	if err != nil {
		t.Fatalf("ParseListNotesResult returned error: %v", err)
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
