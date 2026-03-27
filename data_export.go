package intercom

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// DataExportService handles communication with the data export related
// methods of the Intercom API.
type DataExportService service

// DataExport represents an Intercom content data export job.
type DataExport struct {
	JobIdentifier     string `json:"job_identifier"`
	Status            string `json:"status"`
	DownloadURL       string `json:"download_url"`
	DownloadExpiresAt string `json:"download_expires_at"`
}

// CreateDataExportRequest represents the request body for creating a data export.
type CreateDataExportRequest struct {
	CreatedAtAfter  int64 `json:"created_at_after"`
	CreatedAtBefore int64 `json:"created_at_before"`
}

// --- Parse Functions ---

// ParseDataExportCreateResult decodes a Result into a DataExport.
func ParseDataExportCreateResult(r *Result) (*DataExport, error) {
	return Decode[DataExport](r)
}

// ParseDataExportGetStatusResult decodes a Result into a DataExport.
func ParseDataExportGetStatusResult(r *Result) (*DataExport, error) {
	return Decode[DataExport](r)
}

// ParseDataExportCancelResult decodes a Result into a DataExport.
func ParseDataExportCancelResult(r *Result) (*DataExport, error) {
	return Decode[DataExport](r)
}

// --- Regular Methods ---

// Create starts a new content data export job for the given time range.
func (s *DataExportService) Create(ctx context.Context, body *CreateDataExportRequest) (*DataExport, error) {
	result, err := s.CreateRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseDataExportCreateResult(result)
}

// GetStatus retrieves the status of a data export job.
func (s *DataExportService) GetStatus(ctx context.Context, jobID string) (*DataExport, error) {
	result, err := s.GetStatusRaw(ctx, jobID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseDataExportGetStatusResult(result)
}

// Cancel cancels an in-progress data export job.
func (s *DataExportService) Cancel(ctx context.Context, jobID string) (*DataExport, error) {
	result, err := s.CancelRaw(ctx, jobID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseDataExportCancelResult(result)
}

// --- Raw Methods ---

// CreateRaw starts a new content data export job and returns the full HTTP result.
func (s *DataExportService) CreateRaw(ctx context.Context, body *CreateDataExportRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "export/content/data", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// GetStatusRaw retrieves the status of a data export job with the full HTTP result.
func (s *DataExportService) GetStatusRaw(ctx context.Context, jobID string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("export/content/data/%s", url.PathEscape(jobID)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// CancelRaw cancels an in-progress data export job and returns the full HTTP result.
func (s *DataExportService) CancelRaw(ctx context.Context, jobID string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, fmt.Sprintf("export/cancel/%s", url.PathEscape(jobID)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// Download writes the exported data to w. The data is typically gzipped CSV.
// Unlike other methods, Download handles the HTTP response directly to stream
// binary data rather than decoding JSON.
func (s *DataExportService) Download(ctx context.Context, jobID string, w io.Writer) error {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("download/content/data/%s", url.PathEscape(jobID)), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/octet-stream")
	req = req.WithContext(ctx)

	s.client.logger.Debug("http request", "method", req.Method, "url", req.URL)

	resp, err := s.client.httpClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 400 {
		body, readErr := io.ReadAll(resp.Body)
		if readErr != nil {
			return fmt.Errorf("HTTP %d: failed to read error body: %w", resp.StatusCode, readErr)
		}
		result := buildResult(resp, body)
		return resultError(result)
	}

	_, err = io.Copy(w, resp.Body)
	return err
}
