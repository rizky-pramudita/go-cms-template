package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// ContentType represents a content type definition
type ContentType struct {
	ID           uuid.UUID       `json:"id"`
	Name         string          `json:"name"`
	Slug         string          `json:"slug"`
	SchemaFields json.RawMessage `json:"schema_fields,omitempty"`
	IsActive     bool            `json:"is_active"`
	DisplayOrder int             `json:"display_order"`
	CreatedAt    time.Time       `json:"created_at"`
	UpdatedAt    time.Time       `json:"updated_at"`
}

// CreateContentTypeRequest represents the request to create a content type
type CreateContentTypeRequest struct {
	Name         string          `json:"name"`
	Slug         string          `json:"slug"`
	SchemaFields json.RawMessage `json:"schema_fields,omitempty"`
	IsActive     *bool           `json:"is_active,omitempty"`
	DisplayOrder *int            `json:"display_order,omitempty"`
}

// UpdateContentTypeRequest represents the request to update a content type
type UpdateContentTypeRequest struct {
	Name         *string          `json:"name,omitempty"`
	Slug         *string          `json:"slug,omitempty"`
	SchemaFields *json.RawMessage `json:"schema_fields,omitempty"`
	IsActive     *bool            `json:"is_active,omitempty"`
	DisplayOrder *int             `json:"display_order,omitempty"`
}

// ContentTypeFilter represents filter options for content types
type ContentTypeFilter struct {
	IsActive *bool
	PaginationParams
}
