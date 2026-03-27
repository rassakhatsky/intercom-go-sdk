package intercom

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
)

// CustomObjectsService handles communication with the custom object instance
// related methods of the Intercom API.
type CustomObjectsService service

// CustomObjectInstance represents an Intercom custom object instance.
type CustomObjectInstance struct {
	ID                string            `json:"id"`
	ExternalID        string            `json:"external_id,omitempty"`
	Type              string            `json:"type,omitempty"`
	ExternalCreatedAt *int64            `json:"external_created_at,omitempty"`
	ExternalUpdatedAt *int64            `json:"external_updated_at,omitempty"`
	CreatedAt         int64             `json:"created_at,omitempty"`
	UpdatedAt         int64             `json:"updated_at,omitempty"`
	CustomAttributes  map[string]string `json:"custom_attributes,omitempty"`
}

// CustomObjectInstanceDeleted represents the response from deleting a custom object instance.
type CustomObjectInstanceDeleted struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Deleted bool   `json:"deleted"`
}

// CreateOrUpdateCustomObjectRequest represents a request to create or update a custom object instance.
type CreateOrUpdateCustomObjectRequest struct {
	ExternalID        string            `json:"external_id,omitempty"`
	ExternalCreatedAt *int64            `json:"external_created_at,omitempty"`
	ExternalUpdatedAt *int64            `json:"external_updated_at,omitempty"`
	CustomAttributes  map[string]string `json:"custom_attributes,omitempty"`
}

// --- Parse Functions ---

// ParseCustomObjectGetResult decodes a Result into a CustomObjectInstance.
func ParseCustomObjectGetResult(r *Result) (*CustomObjectInstance, error) {
	return Decode[CustomObjectInstance](r)
}

// ParseCustomObjectGetByExternalIDResult decodes a Result into a CustomObjectInstance.
func ParseCustomObjectGetByExternalIDResult(r *Result) (*CustomObjectInstance, error) {
	return Decode[CustomObjectInstance](r)
}

// ParseCustomObjectCreateOrUpdateResult decodes a Result into a CustomObjectInstance.
func ParseCustomObjectCreateOrUpdateResult(r *Result) (*CustomObjectInstance, error) {
	return Decode[CustomObjectInstance](r)
}

// ParseCustomObjectDeleteResult decodes a Result into a CustomObjectInstanceDeleted.
func ParseCustomObjectDeleteResult(r *Result) (*CustomObjectInstanceDeleted, error) {
	return Decode[CustomObjectInstanceDeleted](r)
}

// ParseCustomObjectDeleteByExternalIDResult decodes a Result into a CustomObjectInstanceDeleted.
func ParseCustomObjectDeleteByExternalIDResult(r *Result) (*CustomObjectInstanceDeleted, error) {
	return Decode[CustomObjectInstanceDeleted](r)
}

// --- Regular Methods ---

// Get retrieves a custom object instance by type identifier and instance ID.
func (s *CustomObjectsService) Get(ctx context.Context, typeIdentifier, instanceID string) (*CustomObjectInstance, error) {
	result, err := s.GetRaw(ctx, typeIdentifier, instanceID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseCustomObjectGetResult(result)
}

// GetByExternalID retrieves a custom object instance by type identifier and external ID.
func (s *CustomObjectsService) GetByExternalID(ctx context.Context, typeIdentifier, externalID string) (*CustomObjectInstance, error) {
	result, err := s.GetByExternalIDRaw(ctx, typeIdentifier, externalID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseCustomObjectGetByExternalIDResult(result)
}

// CreateOrUpdate creates or updates a custom object instance (upsert by external_id).
func (s *CustomObjectsService) CreateOrUpdate(ctx context.Context, typeIdentifier string, body *CreateOrUpdateCustomObjectRequest) (*CustomObjectInstance, error) {
	result, err := s.CreateOrUpdateRaw(ctx, typeIdentifier, body)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseCustomObjectCreateOrUpdateResult(result)
}

// Delete deletes a custom object instance by type identifier and instance ID.
func (s *CustomObjectsService) Delete(ctx context.Context, typeIdentifier, instanceID string) (*CustomObjectInstanceDeleted, error) {
	result, err := s.DeleteRaw(ctx, typeIdentifier, instanceID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseCustomObjectDeleteResult(result)
}

// DeleteByExternalID deletes a custom object instance by type identifier and external ID.
func (s *CustomObjectsService) DeleteByExternalID(ctx context.Context, typeIdentifier, externalID string) (*CustomObjectInstanceDeleted, error) {
	result, err := s.DeleteByExternalIDRaw(ctx, typeIdentifier, externalID)
	if err != nil {
		return nil, err
	}
	if result.Error != nil {
		return nil, resultError(result)
	}
	return ParseCustomObjectDeleteByExternalIDResult(result)
}

// --- Raw Methods ---

// GetRaw retrieves a custom object instance by type identifier and instance ID with the full HTTP result.
func (s *CustomObjectsService) GetRaw(ctx context.Context, typeIdentifier, instanceID string) (*Result, error) {
	path := fmt.Sprintf("custom_object_instances/%s/%s", url.PathEscape(typeIdentifier), url.PathEscape(instanceID))
	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// GetByExternalIDRaw retrieves a custom object instance by type identifier and external ID with the full HTTP result.
func (s *CustomObjectsService) GetByExternalIDRaw(ctx context.Context, typeIdentifier, externalID string) (*Result, error) {
	path := fmt.Sprintf("custom_object_instances/%s?external_id=%s", url.PathEscape(typeIdentifier), url.QueryEscape(externalID))
	req, err := s.client.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// CreateOrUpdateRaw creates or updates a custom object instance with the full HTTP result.
func (s *CustomObjectsService) CreateOrUpdateRaw(ctx context.Context, typeIdentifier string, body *CreateOrUpdateCustomObjectRequest) (*Result, error) {
	path := fmt.Sprintf("custom_object_instances/%s", url.PathEscape(typeIdentifier))
	req, err := s.client.NewRequest(http.MethodPost, path, body)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// DeleteRaw deletes a custom object instance by type identifier and instance ID with the full HTTP result.
func (s *CustomObjectsService) DeleteRaw(ctx context.Context, typeIdentifier, instanceID string) (*Result, error) {
	path := fmt.Sprintf("custom_object_instances/%s/%s", url.PathEscape(typeIdentifier), url.PathEscape(instanceID))
	req, err := s.client.NewRequest(http.MethodDelete, path, nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}

// DeleteByExternalIDRaw deletes a custom object instance by type identifier and external ID with the full HTTP result.
func (s *CustomObjectsService) DeleteByExternalIDRaw(ctx context.Context, typeIdentifier, externalID string) (*Result, error) {
	path := fmt.Sprintf("custom_object_instances/%s?external_id=%s", url.PathEscape(typeIdentifier), url.QueryEscape(externalID))
	req, err := s.client.NewRequest(http.MethodDelete, path, nil)
	if err != nil {
		return nil, err
	}
	return s.client.DoRaw(ctx, req)
}
