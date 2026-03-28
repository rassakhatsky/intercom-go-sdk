package settings

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/rassakhatsky/intercom-go-sdk/internal/api"
)

// JobsService handles communication with the job status related methods
// of the Intercom API.
type JobsService struct {
	client api.Caller
}

// NewJobsService creates a new JobsService.
func NewJobsService(c api.Caller) *JobsService {
	return &JobsService{client: c}
}

// Job represents an Intercom async job status.
type Job struct {
	Type         string `json:"type"`
	ID           string `json:"id"`
	URL          string `json:"url,omitempty"`
	Status       string `json:"status,omitempty"`
	ResourceType string `json:"resource_type,omitempty"`
	ResourceID   string `json:"resource_id,omitempty"`
	ResourceURL  string `json:"resource_url,omitempty"`
}

// --- Parse Functions ---

// ParseJobGetStatusResult decodes a Result into a Job.
func ParseJobGetStatusResult(r *api.Result) (*Job, error) {
	return api.Decode[Job](r)
}

// --- Regular Methods ---

// GetStatus retrieves the status of a job by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/jobs/jobsstatus
func (s *JobsService) GetStatus(ctx context.Context, jobID string) (*Job, error) {
	result, err := s.GetStatusRaw(ctx, jobID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseJobGetStatusResult(result)
}

// --- Raw Methods ---

// GetStatusRaw retrieves the status of a job by ID with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/jobs/jobsstatus
func (s *JobsService) GetStatusRaw(ctx context.Context, jobID string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("jobs/status/%s", url.PathEscape(jobID)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
