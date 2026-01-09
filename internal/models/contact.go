package models

import (
	"encoding/json"
	"net"
	"time"

	"github.com/google/uuid"
)

// ContactStatus represents the status of a contact submission
type ContactStatus int16

const (
	ContactStatusNew      ContactStatus = 1
	ContactStatusRead     ContactStatus = 2
	ContactStatusReplied  ContactStatus = 3
	ContactStatusArchived ContactStatus = 4
)

func (s ContactStatus) String() string {
	switch s {
	case ContactStatusNew:
		return "new"
	case ContactStatusRead:
		return "read"
	case ContactStatusReplied:
		return "replied"
	case ContactStatusArchived:
		return "archived"
	default:
		return "unknown"
	}
}

// ContactSubmission represents a contact form submission
type ContactSubmission struct {
	ID        uuid.UUID       `json:"id"`
	Name      string          `json:"name"`
	Email     string          `json:"email"`
	Phone     *string         `json:"phone,omitempty"`
	Subject   *string         `json:"subject,omitempty"`
	Message   string          `json:"message"`
	Status    ContactStatus   `json:"status"`
	IPAddress *net.IP         `json:"ip_address,omitempty"`
	UserAgent *string         `json:"user_agent,omitempty"`
	Metadata  json.RawMessage `json:"metadata,omitempty"`
	ReadAt    *time.Time      `json:"read_at,omitempty"`
	CreatedAt time.Time       `json:"created_at"`
}

// CreateContactRequest represents the request to create a contact submission
type CreateContactRequest struct {
	Name      string          `json:"name"`
	Email     string          `json:"email"`
	Phone     *string         `json:"phone,omitempty"`
	Subject   *string         `json:"subject,omitempty"`
	Message   string          `json:"message"`
	IPAddress *string         `json:"ip_address,omitempty"`
	UserAgent *string         `json:"user_agent,omitempty"`
	Metadata  json.RawMessage `json:"metadata,omitempty"`
}

// UpdateContactRequest represents the request to update a contact submission
type UpdateContactRequest struct {
	Status *ContactStatus `json:"status,omitempty"`
}

// ContactFilter represents filter options for contact submissions
type ContactFilter struct {
	Status *ContactStatus
	Email  string
	PaginationParams
}
