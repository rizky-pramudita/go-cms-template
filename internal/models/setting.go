package models

import (
	"time"

	"github.com/google/uuid"
)

// Setting represents a key-value setting
type Setting struct {
	ID          uuid.UUID `json:"id"`
	Key         string    `json:"key"`
	Value       *string   `json:"value,omitempty"`
	Description *string   `json:"description,omitempty"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateSettingRequest represents the request to create a setting
type CreateSettingRequest struct {
	Key         string  `json:"key"`
	Value       *string `json:"value,omitempty"`
	Description *string `json:"description,omitempty"`
}

// UpdateSettingRequest represents the request to update a setting
type UpdateSettingRequest struct {
	Value       *string `json:"value,omitempty"`
	Description *string `json:"description,omitempty"`
}

// SettingFilter represents filter options for settings
type SettingFilter struct {
	Search string
	PaginationParams
}
