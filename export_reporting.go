package intercom

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
)

// ExportReportingService handles communication with the reporting data export
// related methods of the Intercom API.
type ExportReportingService service

// ExportReportingJob represents a reporting data export job.
type ExportReportingJob struct {
	JobIdentifier     string `json:"job_identifier"`
	Status            string `json:"status"`
	DownloadURL       string `json:"download_url"`
	DownloadExpiresAt string `json:"download_expires_at"`
}

// EnqueueExportReportingRequest represents the request body for enqueueing
// a reporting data export job.
type EnqueueExportReportingRequest struct {
	DatasetID    string   `json:"dataset_id"`
	AttributeIDs []string `json:"attribute_ids"`
	StartTime    int64    `json:"start_time"`
	EndTime      int64    `json:"end_time"`
}

// ReportingDataset represents an available reporting dataset.
type ReportingDataset struct {
	ID         string                      `json:"id"`
	Name       string                      `json:"name"`
	Attributes []ReportingDatasetAttribute `json:"attributes"`
}

// ReportingDatasetAttribute represents an attribute within a reporting dataset.
type ReportingDatasetAttribute struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Type string `json:"type"`
}

// ReportingDatasetsResponse wraps the datasets list API response.
type ReportingDatasetsResponse struct {
	Type string             `json:"type"`
	Data []ReportingDataset `json:"data"`
}

// --- Parse Functions ---

// ParseExportReportingEnqueueResult decodes a Result into an ExportReportingJob.
func ParseExportReportingEnqueueResult(r *Result) (*ExportReportingJob, error) {
	return Decode[ExportReportingJob](r)
}

// ParseExportReportingGetStatusResult decodes a Result into an ExportReportingJob.
func ParseExportReportingGetStatusResult(r *Result) (*ExportReportingJob, error) {
	return Decode[ExportReportingJob](r)
}

// ParseExportReportingGetDatasetsResult decodes a Result into a ReportingDatasetsResponse.
func ParseExportReportingGetDatasetsResult(r *Result) (*ReportingDatasetsResponse, error) {
	return Decode[ReportingDatasetsResponse](r)
}

// --- Regular Methods ---

// Enqueue starts a new reporting data export job.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/export/paths/~1export~1reporting_data~1enqueue/post
func (s *ExportReportingService) Enqueue(ctx context.Context, body *EnqueueExportReportingRequest) (*ExportReportingJob, error) {
	result, err := s.EnqueueRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseExportReportingEnqueueResult(result)
}

// GetStatus retrieves the status of a reporting data export job.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/export/paths/~1export~1reporting_data~1{job_identifier}/get
func (s *ExportReportingService) GetStatus(ctx context.Context, jobIdentifier string) (*ExportReportingJob, error) {
	result, err := s.GetStatusRaw(ctx, jobIdentifier)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseExportReportingGetStatusResult(result)
}

// GetDatasets returns the list of available reporting datasets and their attributes.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/export/paths/~1export~1reporting_data~1get_datasets/get
func (s *ExportReportingService) GetDatasets(ctx context.Context) ([]ReportingDataset, error) {
	result, err := s.GetDatasetsRaw(ctx)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	resp, err := ParseExportReportingGetDatasetsResult(result)
	if err != nil {
		return nil, err
	}
	return resp.Data, nil
}

// --- Raw Methods ---

// EnqueueRaw starts a new reporting data export job and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/export/paths/~1export~1reporting_data~1enqueue/post
func (s *ExportReportingService) EnqueueRaw(ctx context.Context, body *EnqueueExportReportingRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "export/reporting_data/enqueue", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// GetStatusRaw retrieves the status of a reporting data export job with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/export/paths/~1export~1reporting_data~1{job_identifier}/get
func (s *ExportReportingService) GetStatusRaw(ctx context.Context, jobIdentifier string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("export/reporting_data/%s", url.PathEscape(jobIdentifier)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// GetDatasetsRaw returns the list of available reporting datasets with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/export/paths/~1export~1reporting_data~1get_datasets/get
func (s *ExportReportingService) GetDatasetsRaw(ctx context.Context) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, "export/reporting_data/get_datasets", nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// Download writes the reporting export data to w. The data is typically CSV.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/export/paths/~1download~1reporting_data~1{job_identifier}/get
func (s *ExportReportingService) Download(ctx context.Context, jobIdentifier string, w io.Writer) error {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("download/reporting_data/%s", url.PathEscape(jobIdentifier)), nil)
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
