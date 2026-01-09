package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// PostStatus represents the status of a content post
type PostStatus int16

const (
	PostStatusDraft     PostStatus = 1
	PostStatusPublished PostStatus = 2
	PostStatusArchived  PostStatus = 3
)

func (s PostStatus) String() string {
	switch s {
	case PostStatusDraft:
		return "draft"
	case PostStatusPublished:
		return "published"
	case PostStatusArchived:
		return "archived"
	default:
		return "unknown"
	}
}

// ContentPost represents a content post
type ContentPost struct {
	ID            uuid.UUID       `json:"id"`
	ContentTypeID uuid.UUID       `json:"content_type_id"`
	AuthorID      uuid.UUID       `json:"author_id"`
	Title         string          `json:"title"`
	Slug          string          `json:"slug"`
	Excerpt       *string         `json:"excerpt,omitempty"`
	Content       *string         `json:"content,omitempty"`
	Metadata      json.RawMessage `json:"metadata,omitempty"`
	Status        PostStatus      `json:"status"`
	PublishedAt   *time.Time      `json:"published_at,omitempty"`
	ViewCount     int             `json:"view_count"`
	CreatedAt     time.Time       `json:"created_at"`
	UpdatedAt     time.Time       `json:"updated_at"`

	// Relations (populated on demand)
	ContentType *ContentType  `json:"content_type,omitempty"`
	Author      *UserResponse `json:"author,omitempty"`
	Tags        []Tag         `json:"tags,omitempty"`
	Media       []PostMedia   `json:"media,omitempty"`
}

// CreatePostRequest represents the request to create a post
type CreatePostRequest struct {
	ContentTypeID uuid.UUID       `json:"content_type_id"`
	AuthorID      uuid.UUID       `json:"author_id"`
	Title         string          `json:"title"`
	Slug          string          `json:"slug"`
	Excerpt       *string         `json:"excerpt,omitempty"`
	Content       *string         `json:"content,omitempty"`
	Metadata      json.RawMessage `json:"metadata,omitempty"`
	Status        *PostStatus     `json:"status,omitempty"`
	PublishedAt   *time.Time      `json:"published_at,omitempty"`
	TagIDs        []uuid.UUID     `json:"tag_ids,omitempty"`
}

// UpdatePostRequest represents the request to update a post
type UpdatePostRequest struct {
	ContentTypeID *uuid.UUID       `json:"content_type_id,omitempty"`
	Title         *string          `json:"title,omitempty"`
	Slug          *string          `json:"slug,omitempty"`
	Excerpt       *string          `json:"excerpt,omitempty"`
	Content       *string          `json:"content,omitempty"`
	Metadata      *json.RawMessage `json:"metadata,omitempty"`
	Status        *PostStatus      `json:"status,omitempty"`
	PublishedAt   *time.Time       `json:"published_at,omitempty"`
	TagIDs        *[]uuid.UUID     `json:"tag_ids,omitempty"`
}

// PostFilter represents filter options for posts
type PostFilter struct {
	ContentTypeID *uuid.UUID
	AuthorID      *uuid.UUID
	Status        *PostStatus
	Search        string
	PaginationParams
}
