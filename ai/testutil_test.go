package ai_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rassakhatsky/intercom-go-sdk/ai"
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

var sharedCaller *testCaller

func setup() (svc *ai.ContentService, mux *http.ServeMux, teardown func()) {
	mux = http.NewServeMux()
	server := httptest.NewServer(mux)
	sharedCaller = &testCaller{baseURL: server.URL, client: server.Client()}
	svc = ai.NewContentService(sharedCaller)
	return svc, mux, server.Close
}

func setupVoice() *ai.VoiceService {
	return ai.NewVoiceService(sharedCaller)
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
		t.Errorf("Header %v = %v, want %v", header, got, want)
	}
}
