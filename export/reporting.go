package export

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/rassakhatsky/intercom-go-sdk/internal/api"
)

// ReportingService handles communication with the reporting data export
// related methods of the Intercom API.
type ReportingService struct {
	client api.Caller
}

// NewReportingService creates a new export reporting Service.
func NewReportingService(c api.Caller) *ReportingService {
	return &ReportingService{client: c}
}

// Job represents a reporting data export job.
type Job struct {
	JobIdentifier     string `json:"job_identifier"`
	Status            string `json:"status"`
	DownloadURL       string `json:"download_url"`
	DownloadExpiresAt string `json:"download_expires_at"`
}

// EnqueueRequest represents the request body for enqueueing
// a reporting data export job.
type EnqueueRequest struct {
	DatasetID    string   `json:"dataset_id"`
	AttributeIDs []string `json:"attribute_ids"`
	StartTime    int64    `json:"start_time"`
	EndTime      int64    `json:"end_time"`
}

// Dataset represents an available reporting dataset.
type Dataset struct {
	ID         string             `json:"id"`
	Name       string             `json:"name"`
	Attributes []DatasetAttribute `json:"attributes"`
}

// DatasetAttribute represents an attribute within a reporting dataset.
type DatasetAttribute struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// DatasetsResponse wraps the datasets list API response.
type DatasetsResponse struct {
	Type string    `json:"type"`
	Data []Dataset `json:"data"`
}

// --- Parse Functions ---

// ParseEnqueueResult decodes a Result into a Job.
func ParseEnqueueResult(r *api.Result) (*Job, error) {
	return api.Decode[Job](r)
}

// ParseGetStatusResult decodes a Result into a Job.
func ParseGetStatusResult(r *api.Result) (*Job, error) {
	return api.Decode[Job](r)
}

// ParseGetDatasetsResult decodes a Result into a DatasetsResponse.
func ParseGetDatasetsResult(r *api.Result) (*DatasetsResponse, error) {
	return api.Decode[DatasetsResponse](r)
}

// --- Regular Methods ---

// Enqueue starts a new reporting data export job.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/export/enqueuereportingdataexport
func (s *ReportingService) Enqueue(ctx context.Context, body *EnqueueRequest) (*Job, error) {
	result, err := s.EnqueueRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseEnqueueResult(result)
}

// GetStatus retrieves the status of a reporting data export job.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/export/getreportingdataexport
func (s *ReportingService) GetStatus(ctx context.Context, jobIdentifier string) (*Job, error) {
	result, err := s.GetStatusRaw(ctx, jobIdentifier)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseGetStatusResult(result)
}

// GetDatasets returns the list of available reporting datasets and their attributes.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/export/getreportingdatasets
func (s *ReportingService) GetDatasets(ctx context.Context) ([]Dataset, error) {
	result, err := s.GetDatasetsRaw(ctx)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	resp, err := ParseGetDatasetsResult(result)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// --- Raw Methods ---

// EnqueueRaw starts a new reporting data export job and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/export/enqueuereportingdataexport
func (s *ReportingService) EnqueueRaw(ctx context.Context, body *EnqueueRequest) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "export/reporting_data/enqueue", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// GetStatusRaw retrieves the status of a reporting data export job with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/export/getreportingdataexport
func (s *ReportingService) GetStatusRaw(ctx context.Context, jobIdentifier string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("export/reporting_data/%s", url.PathEscape(jobIdentifier)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// GetDatasetsRaw returns the list of available reporting datasets with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/export/getreportingdatasets
func (s *ReportingService) GetDatasetsRaw(ctx context.Context) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, "export/reporting_data/get_datasets", nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// Download writes the reporting export data to w. The data is typically CSV.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/export/downloadreportingdataexport
func (s *ReportingService) Download(ctx context.Context, jobIdentifier string, w io.Writer) error {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("download/reporting_data/%s", url.PathEscape(jobIdentifier)), nil)
	if err != nil {
		return err
	}
	req.Header.Set("Accept", "application/octet-stream")
	return s.client.DoDownload(ctx, req, w)
}
