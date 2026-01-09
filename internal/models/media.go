package models

import (
	"encoding/json"
	"time"

	"github.com/google/uuid"
)

// FileType represents the type of media file
type FileType int16

const (
	FileTypeImage    FileType = 1
	FileTypeVideo    FileType = 2
	FileTypeDocument FileType = 3
)

func (t FileType) String() string {
	switch t {
	case FileTypeImage:
		return "image"
	case FileTypeVideo:
		return "video"
	case FileTypeDocument:
		return "document"
	default:
		return "unknown"
	}
}

// MediaRole represents the role of media in a post
type MediaRole int16

const (
	MediaRoleFeatured MediaRole = 1
	MediaRoleGallery  MediaRole = 2
	MediaRoleContent  MediaRole = 3
)

func (r MediaRole) String() string {
	switch r {
	case MediaRoleFeatured:
		return "featured"
	case MediaRoleGallery:
		return "gallery"
	case MediaRoleContent:
		return "content"
	default:
		return "unknown"
	}
}

// Media represents a media file
type Media struct {
	ID         uuid.UUID       `json:"id"`
	FileName   string          `json:"file_name"`
	ObjectKey  string          `json:"object_key"`
	BucketName string          `json:"bucket_name"`
	CDNUrl     *string         `json:"cdn_url,omitempty"`
	FileType   FileType        `json:"file_type"`
	MimeType   string          `json:"mime_type"`
	FileSize   int             `json:"file_size"`
	Dimensions json.RawMessage `json:"dimensions,omitempty"`
	Variants   json.RawMessage `json:"variants,omitempty"`
	AltText    *string         `json:"alt_text,omitempty"`
	Checksum   *string         `json:"checksum,omitempty"`
	CreatedAt  time.Time       `json:"created_at"`
}

// PostMedia represents the relationship between a post and media
type PostMedia struct {
	ID           uuid.UUID `json:"id"`
	PostID       uuid.UUID `json:"post_id"`
	MediaID      uuid.UUID `json:"media_id"`
	MediaRole    MediaRole `json:"media_role"`
	DisplayOrder int       `json:"display_order"`
	CreatedAt    time.Time `json:"created_at"`
	Media        *Media    `json:"media,omitempty"`
}

// CreateMediaRequest represents the request to create a media record
type CreateMediaRequest struct {
	FileName   string          `json:"file_name"`
	ObjectKey  string          `json:"object_key"`
	BucketName string          `json:"bucket_name"`
	CDNUrl     *string         `json:"cdn_url,omitempty"`
	FileType   FileType        `json:"file_type"`
	MimeType   string          `json:"mime_type"`
	FileSize   int             `json:"file_size"`
	Dimensions json.RawMessage `json:"dimensions,omitempty"`
	Variants   json.RawMessage `json:"variants,omitempty"`
	AltText    *string         `json:"alt_text,omitempty"`
	Checksum   *string         `json:"checksum,omitempty"`
}

// UpdateMediaRequest represents the request to update a media record
type UpdateMediaRequest struct {
	FileName   *string          `json:"file_name,omitempty"`
	CDNUrl     *string          `json:"cdn_url,omitempty"`
	Dimensions *json.RawMessage `json:"dimensions,omitempty"`
	Variants   *json.RawMessage `json:"variants,omitempty"`
	AltText    *string          `json:"alt_text,omitempty"`
}

// AttachMediaRequest represents the request to attach media to a post
type AttachMediaRequest struct {
	MediaID      uuid.UUID `json:"media_id"`
	MediaRole    MediaRole `json:"media_role"`
	DisplayOrder *int      `json:"display_order,omitempty"`
}

// MediaFilter represents filter options for media
type MediaFilter struct {
	FileType *FileType
	Search   string
	PaginationParams
}
