package export

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/rassakhatsky/intercom-go-sdk/internal/api"
)

// DataService handles communication with the data export related
// methods of the Intercom API.
type DataService struct {
	client api.Caller
}

// NewDataService creates a new data export Service.
func NewDataService(c api.Caller) *DataService {
	return &DataService{client: c}
}

// DataExport represents an Intercom content data export job.
type DataExport struct {
	JobIdentifier     string `json:"job_identifier"`
	Status            string `json:"status"`
	DownloadURL       string `json:"download_url"`
	DownloadExpiresAt string `json:"download_expires_at"`
}

// CreateRequest represents the request body for creating a data export.
type CreateRequest struct {
	CreatedAtAfter  int64 `json:"created_at_after"`
	CreatedAtBefore int64 `json:"created_at_before"`
}

// --- Parse Functions ---

// ParseCreateResult decodes a Result into a DataExport.
func ParseCreateResult(r *api.Result) (*DataExport, error) {
	return api.Decode[DataExport](r)
}

// ParseGetStatusResult decodes a Result into a DataExport.
func ParseDataGetStatusResult(r *api.Result) (*DataExport, error) {
	return api.Decode[DataExport](r)
}

// ParseCancelResult decodes a Result into a DataExport.
func ParseCancelResult(r *api.Result) (*DataExport, error) {
	return api.Decode[DataExport](r)
}

// --- Regular Methods ---

// Create starts a new content data export job for the given time range.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/data-export/createdataexport
func (s *DataService) Create(ctx context.Context, body *CreateRequest) (*DataExport, error) {
	result, err := s.CreateRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseCreateResult(result)
}

// GetStatus retrieves the status of a data export job.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/data-export/getdataexport
func (s *DataService) GetStatus(ctx context.Context, jobID string) (*DataExport, error) {
	result, err := s.GetStatusRaw(ctx, jobID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseDataGetStatusResult(result)
}

// Cancel cancels an in-progress data export job.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/data-export/canceldataexport
func (s *DataService) Cancel(ctx context.Context, jobID string) (*DataExport, error) {
	result, err := s.CancelRaw(ctx, jobID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseCancelResult(result)
}

// --- Raw Methods ---

// CreateRaw starts a new content data export job and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/data-export/createdataexport
func (s *DataService) CreateRaw(ctx context.Context, body *CreateRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "export/content/data", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// GetStatusRaw retrieves the status of a data export job with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/data-export/getdataexport
func (s *DataService) GetStatusRaw(ctx context.Context, jobID string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("export/content/data/%s", url.PathEscape(jobID)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// CancelRaw cancels an in-progress data export job and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/data-export/canceldataexport
func (s *DataService) CancelRaw(ctx context.Context, jobID string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, fmt.Sprintf("export/cancel/%s", url.PathEscape(jobID)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// Download writes the exported data to w. The data is typically gzipped CSV.
// Unlike other methods, Download streams binary data rather than decoding JSON.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/data-export/downloaddataexport
func (s *DataService) Download(ctx context.Context, jobID string, w io.Writer) error {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("download/content/data/%s", url.PathEscape(jobID)), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/octet-stream")
	return s.client.DoDownload(ctx, req, w)
}
