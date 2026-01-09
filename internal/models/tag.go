package models

import (
	"time"

	"github.com/google/uuid"
)

// Tag represents a content tag
type Tag struct {
	ID        uuid.UUID `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	CreatedAt time.Time `json:"created_at"`
}

// CreateTagRequest represents the request to create a tag
type CreateTagRequest struct {
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// UpdateTagRequest represents the request to update a tag
type UpdateTagRequest struct {
	Name *string `json:"name,omitempty"`
	Slug *string `json:"slug,omitempty"`
}

// TagFilter represents filter options for tags
type TagFilter struct {
	Search string
	PaginationParams
}
