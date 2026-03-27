package intercom

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// JobsService handles communication with the job status related methods
// of the Intercom API.
type JobsService service

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
func ParseJobGetStatusResult(r *Result) (*Job, error) {
	return Decode[Job](r)
}

// --- Regular Methods ---

// GetStatus retrieves the status of a job by ID.
func (s *JobsService) GetStatus(ctx context.Context, jobID string) (*Job, error) {
	result, err := s.GetStatusRaw(ctx, jobID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseJobGetStatusResult(result)
}

// --- Raw Methods ---

// GetStatusRaw retrieves the status of a job by ID with the full HTTP result.
func (s *JobsService) GetStatusRaw(ctx context.Context, jobID string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("jobs/status/%s", url.PathEscape(jobID)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
