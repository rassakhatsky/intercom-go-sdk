package workflows

import (
	"context"
	"fmt"
	"net/http"
	"net/url"

	"github.com/rassakhatsky/intercom-go-sdk/api"
)

// Service handles communication with the workflow
// related methods of the Intercom API.
type Service struct {
	client api.Caller
}

// NewService creates a new workflows service.
func NewService(c api.Caller) *Service {
	return &Service{client: c}
}

// Export represents a workflow export containing the complete
// workflow configuration.
type Export struct {
	ExportVersion string    `json:"export_version"`
	ExportedAt    string    `json:"exported_at"`
	AppID         int       `json:"app_id"`
	Workflow      *Workflow `json:"workflow"`
}

// Workflow represents a workflow configuration.
type Workflow struct {
	ID               string           `json:"id"`
	Title            string           `json:"title"`
	Description      *string          `json:"description"`
	TriggerType      string           `json:"trigger_type"`
	State            string           `json:"state"`
	TargetChannels   []string         `json:"target_channels"`
	PreferredDevices []string         `json:"preferred_devices"`
	CreatedAt        string           `json:"created_at"`
	UpdatedAt        string           `json:"updated_at"`
	Targeting        map[string]any   `json:"targeting"`
	Snapshot         map[string]any   `json:"snapshot"`
	Attributes       []map[string]any `json:"attributes"`
	EmbeddedRules    []map[string]any `json:"embedded_rules"`
}

// --- Parse Functions ---

// ParseExportResult decodes a Result into an Export.
func ParseExportResult(r *api.Result) (*Export, error) {
	return api.Decode[Export](r)
}

// --- Regular Methods ---

// Export retrieves the complete workflow configuration by its ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/workflows/exportworkflow
func (s *Service) Export(ctx context.Context, id string) (*Export, error) {
	result, err := s.ExportRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, api.ResultError(result)
	}
	return ParseExportResult(result)
}

// --- Raw Methods ---

// ExportRaw retrieves the complete workflow configuration by its ID
// and returns the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/workflows/exportworkflow
func (s *Service) ExportRaw(ctx context.Context, id string) (*api.Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("export/workflows/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
