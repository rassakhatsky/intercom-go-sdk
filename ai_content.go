package intercom

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// AIContentService handles communication with the AI content related
// methods of the Intercom API, including content import sources and
// external pages.
type AIContentService service

// ContentImportSource represents an Intercom content import source.
type ContentImportSource struct {
	Type         string `json:"type"`
	ID           int    `json:"id"`
	URL          string `json:"url"`
	SyncBehavior string `json:"sync_behavior"`
	Status       string `json:"status"`
	LastSyncedAt int64  `json:"last_synced_at"`
	CreatedAt    int64  `json:"created_at"`
	UpdatedAt    int64  `json:"updated_at"`
}

// ExternalPage represents an Intercom external page.
type ExternalPage struct {
	Type                  string `json:"type"`
	ID                    string `json:"id"`
	Title                 string `json:"title"`
	HTML                  string `json:"html"`
	URL                   string `json:"url,omitempty"`
	AIAgentAvailability   bool   `json:"ai_agent_availability"`
	AICopilotAvailability bool   `json:"ai_copilot_availability"`
	FinAvailability       *bool  `json:"fin_availability,omitempty"`
	Locale                string `json:"locale"`
	SourceID              int    `json:"source_id"`
	ExternalID            string `json:"external_id"`
	CreatedAt             int64  `json:"created_at"`
	UpdatedAt             int64  `json:"updated_at"`
	LastIngestedAt        int64  `json:"last_ingested_at"`
}

// CreateContentImportSourceRequest represents the request body for creating a content import source.
type CreateContentImportSourceRequest struct {
	SyncBehavior string `json:"sync_behavior"`
	URL          string `json:"url"`
	Status       string `json:"status,omitempty"`
}

// UpdateContentImportSourceRequest represents the request body for updating a content import source.
type UpdateContentImportSourceRequest struct {
	SyncBehavior string `json:"sync_behavior"`
	URL          string `json:"url"`
	Status       string `json:"status,omitempty"`
}

// CreateExternalPageRequest represents the request body for creating an external page.
type CreateExternalPageRequest struct {
	Title                 string `json:"title"`
	HTML                  string `json:"html"`
	URL                   string `json:"url,omitempty"`
	AIAgentAvailability   *bool  `json:"ai_agent_availability,omitempty"`
	AICopilotAvailability *bool  `json:"ai_copilot_availability,omitempty"`
	Locale                string `json:"locale"`
	SourceID              int    `json:"source_id"`
	ExternalID            string `json:"external_id"`
}

// UpdateExternalPageRequest represents the request body for updating an external page.
type UpdateExternalPageRequest struct {
	Title           string `json:"title,omitempty"`
	HTML            string `json:"html,omitempty"`
	URL             string `json:"url,omitempty"`
	FinAvailability *bool  `json:"fin_availability,omitempty"`
	Locale          string `json:"locale,omitempty"`
	SourceID        int    `json:"source_id,omitempty"`
	ExternalID      string `json:"external_id,omitempty"`
}

// --- Parse Functions ---

// ParseAIContentListContentImportSourcesResult decodes a Result into a PagedResult[ContentImportSource].
func ParseAIContentListContentImportSourcesResult(r *Result) (*PagedResult[ContentImportSource], error) {
	return Decode[PagedResult[ContentImportSource]](r)
}

// ParseAIContentGetContentImportSourceResult decodes a Result into a ContentImportSource.
func ParseAIContentGetContentImportSourceResult(r *Result) (*ContentImportSource, error) {
	return Decode[ContentImportSource](r)
}

// ParseAIContentCreateContentImportSourceResult decodes a Result into a ContentImportSource.
func ParseAIContentCreateContentImportSourceResult(r *Result) (*ContentImportSource, error) {
	return Decode[ContentImportSource](r)
}

// ParseAIContentUpdateContentImportSourceResult decodes a Result into a ContentImportSource.
func ParseAIContentUpdateContentImportSourceResult(r *Result) (*ContentImportSource, error) {
	return Decode[ContentImportSource](r)
}

// ParseAIContentDeleteContentImportSourceResult decodes a Result into an Empty.
func ParseAIContentDeleteContentImportSourceResult(r *Result) (*Empty, error) {
	return Decode[Empty](r)
}

// ParseAIContentListExternalPagesResult decodes a Result into a PagedResult[ExternalPage].
func ParseAIContentListExternalPagesResult(r *Result) (*PagedResult[ExternalPage], error) {
	return Decode[PagedResult[ExternalPage]](r)
}

// ParseAIContentGetExternalPageResult decodes a Result into an ExternalPage.
func ParseAIContentGetExternalPageResult(r *Result) (*ExternalPage, error) {
	return Decode[ExternalPage](r)
}

// ParseAIContentCreateExternalPageResult decodes a Result into an ExternalPage.
func ParseAIContentCreateExternalPageResult(r *Result) (*ExternalPage, error) {
	return Decode[ExternalPage](r)
}

// ParseAIContentUpdateExternalPageResult decodes a Result into an ExternalPage.
func ParseAIContentUpdateExternalPageResult(r *Result) (*ExternalPage, error) {
	return Decode[ExternalPage](r)
}

// ParseAIContentDeleteExternalPageResult decodes a Result into an ExternalPage.
func ParseAIContentDeleteExternalPageResult(r *Result) (*ExternalPage, error) {
	return Decode[ExternalPage](r)
}

// --- Content Import Sources ---

// ListContentImportSources returns all content import sources.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ai-content/listcontentimportsources
func (s *AIContentService) ListContentImportSources(ctx context.Context) (*PagedResult[ContentImportSource], error) {
	result, err := s.ListContentImportSourcesRaw(ctx)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseAIContentListContentImportSourcesResult(result)
}

// GetContentImportSource retrieves a content import source by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ai-content/getcontentimportsource
func (s *AIContentService) GetContentImportSource(ctx context.Context, id string) (*ContentImportSource, error) {
	result, err := s.GetContentImportSourceRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseAIContentGetContentImportSourceResult(result)
}

// CreateContentImportSource creates a new content import source.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ai-content/createcontentimportsource
func (s *AIContentService) CreateContentImportSource(ctx context.Context, body *CreateContentImportSourceRequest) (*ContentImportSource, error) {
	result, err := s.CreateContentImportSourceRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseAIContentCreateContentImportSourceResult(result)
}

// UpdateContentImportSource updates an existing content import source.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ai-content/updatecontentimportsource
func (s *AIContentService) UpdateContentImportSource(ctx context.Context, id string, body *UpdateContentImportSourceRequest) (*ContentImportSource, error) {
	result, err := s.UpdateContentImportSourceRaw(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseAIContentUpdateContentImportSourceResult(result)
}

// DeleteContentImportSource deletes a content import source by ID.
// This also deletes all external pages imported from this source.
// The API returns 204 No Content on success.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ai-content/deletecontentimportsource
func (s *AIContentService) DeleteContentImportSource(ctx context.Context, id string) error {
	result, err := s.DeleteContentImportSourceRaw(ctx, id)
	if err != nil {
		return err
	}
	if result.Error != nil {
		return resultError(result)
	}
	return nil
}

// --- External Pages ---

// ListExternalPages returns all external pages.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ai-content/listexternalpages
func (s *AIContentService) ListExternalPages(ctx context.Context) (*PagedResult[ExternalPage], error) {
	result, err := s.ListExternalPagesRaw(ctx)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseAIContentListExternalPagesResult(result)
}

// GetExternalPage retrieves an external page by ID.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ai-content/getexternalpage
func (s *AIContentService) GetExternalPage(ctx context.Context, id string) (*ExternalPage, error) {
	result, err := s.GetExternalPageRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseAIContentGetExternalPageResult(result)
}

// CreateExternalPage creates a new external page.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ai-content/createexternalpage
func (s *AIContentService) CreateExternalPage(ctx context.Context, body *CreateExternalPageRequest) (*ExternalPage, error) {
	result, err := s.CreateExternalPageRaw(ctx, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseAIContentCreateExternalPageResult(result)
}

// UpdateExternalPage updates an existing external page.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ai-content/updateexternalpage
func (s *AIContentService) UpdateExternalPage(ctx context.Context, id string, body *UpdateExternalPageRequest) (*ExternalPage, error) {
	result, err := s.UpdateExternalPageRaw(ctx, id, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseAIContentUpdateExternalPageResult(result)
}

// DeleteExternalPage deletes an external page by ID.
// Returns the deleted external page object.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ai-content/deleteexternalpage
func (s *AIContentService) DeleteExternalPage(ctx context.Context, id string) (*ExternalPage, error) {
	result, err := s.DeleteExternalPageRaw(ctx, id)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseAIContentDeleteExternalPageResult(result)
}

// --- Raw Content Import Sources ---

// ListContentImportSourcesRaw returns all content import sources with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ai-content/listcontentimportsources
func (s *AIContentService) ListContentImportSourcesRaw(ctx context.Context) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, "ai/content_import_sources", nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// GetContentImportSourceRaw retrieves a content import source by ID with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ai-content/getcontentimportsource
func (s *AIContentService) GetContentImportSourceRaw(ctx context.Context, id string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("ai/content_import_sources/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// CreateContentImportSourceRaw creates a new content import source with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ai-content/createcontentimportsource
func (s *AIContentService) CreateContentImportSourceRaw(ctx context.Context, body *CreateContentImportSourceRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "ai/content_import_sources", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// UpdateContentImportSourceRaw updates an existing content import source with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ai-content/updatecontentimportsource
func (s *AIContentService) UpdateContentImportSourceRaw(ctx context.Context, id string, body *UpdateContentImportSourceRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPut, fmt.Sprintf("ai/content_import_sources/%s", url.PathEscape(id)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// DeleteContentImportSourceRaw deletes a content import source by ID with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ai-content/deletecontentimportsource
func (s *AIContentService) DeleteContentImportSourceRaw(ctx context.Context, id string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodDelete, fmt.Sprintf("ai/content_import_sources/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// --- Raw External Pages ---

// ListExternalPagesRaw returns all external pages with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ai-content/listexternalpages
func (s *AIContentService) ListExternalPagesRaw(ctx context.Context) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, "ai/external_pages", nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// GetExternalPageRaw retrieves an external page by ID with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ai-content/getexternalpage
func (s *AIContentService) GetExternalPageRaw(ctx context.Context, id string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodGet, fmt.Sprintf("ai/external_pages/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// CreateExternalPageRaw creates a new external page with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ai-content/createexternalpage
func (s *AIContentService) CreateExternalPageRaw(ctx context.Context, body *CreateExternalPageRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPost, "ai/external_pages", body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// UpdateExternalPageRaw updates an existing external page with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ai-content/updateexternalpage
func (s *AIContentService) UpdateExternalPageRaw(ctx context.Context, id string, body *UpdateExternalPageRequest) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodPut, fmt.Sprintf("ai/external_pages/%s", url.PathEscape(id)), body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// DeleteExternalPageRaw deletes an external page by ID with the full HTTP result.
//
// See: https://developers.intercom.com/docs/references/rest-api/api.intercom.io/ai-content/deleteexternalpage
func (s *AIContentService) DeleteExternalPageRaw(ctx context.Context, id string) (*Result, error) {
	req, err := s.client.NewRequest(http.MethodDelete, fmt.Sprintf("ai/external_pages/%s", url.PathEscape(id)), nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
