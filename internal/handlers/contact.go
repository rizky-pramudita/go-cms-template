package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/keeps-dev/go-cms-template/internal/models"
	"github.com/keeps-dev/go-cms-template/internal/repository"
	"github.com/keeps-dev/go-cms-template/internal/response"
)

type ContactHandler struct {
	repo *repository.ContactRepository
}

func NewContactHandler(repo *repository.ContactRepository) *ContactHandler {
	return &ContactHandler{repo: repo}
}

// List godoc
// @Summary List contact submissions
// @Description Get all contact submissions with optional filtering
// @Tags contacts
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param status query int false "Filter by status (1=new, 2=read, 3=replied, 4=archived)"
// @Param email query string false "Filter by email"
// @Success 200 {object} response.APIResponse
// @Router /api/v1/contacts [get]
func (h *ContactHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := models.ContactFilter{
		PaginationParams: parsePaginationParams(r),
		Email:            r.URL.Query().Get("email"),
	}

	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		if s, err := strconv.Atoi(statusStr); err == nil {
			status := models.ContactStatus(s)
			filter.Status = &status
		}
	}

	contacts, total, err := h.repo.List(r.Context(), filter)
	if err != nil {
		response.InternalError(w, "Failed to list contact submissions")
		return
	}

	response.JSONWithMeta(w, http.StatusOK, contacts, &response.Meta{
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		Total:      total,
		TotalPages: int(total)/filter.PageSize + 1,
	})
}

// Get godoc
// @Summary Get contact submission by ID
// @Description Get a single contact submission by its ID
// @Tags contacts
// @Produce json
// @Param id path string true "Contact Submission ID"
// @Success 200 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /api/v1/contacts/{id} [get]
func (h *ContactHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, "Invalid contact ID")
		return
	}

	contact, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.NotFound(w, "Contact submission not found")
			return
		}
		response.InternalError(w, "Failed to get contact submission")
		return
	}

	response.OK(w, contact)
}

// Create godoc
// @Summary Create contact submission
// @Description Create a new contact submission (public endpoint)
// @Tags contacts
// @Accept json
// @Produce json
// @Param body body models.CreateContactRequest true "Contact data"
// @Success 201 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Router /api/v1/contacts [post]
func (h *ContactHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateContactRequest
	if err := decodeJSON(r, &req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	// Validate required fields
	validationErrors := make(map[string]string)
	if req.Name == "" {
		validationErrors["name"] = "Name is required"
	}
	if req.Email == "" {
		validationErrors["email"] = "Email is required"
	}
	if req.Message == "" {
		validationErrors["message"] = "Message is required"
	}

	if len(validationErrors) > 0 {
		response.ValidationError(w, validationErrors)
		return
	}

	// Capture client info
	ipAddr := r.Header.Get("X-Forwarded-For")
	if ipAddr == "" {
		ipAddr = r.Header.Get("X-Real-IP")
	}
	if ipAddr == "" {
		ipAddr = r.RemoteAddr
	}
	req.IPAddress = &ipAddr

	userAgent := r.Header.Get("User-Agent")
	req.UserAgent = &userAgent

	contact, err := h.repo.Create(r.Context(), &req)
	if err != nil {
		response.InternalError(w, "Failed to create contact submission")
		return
	}

	response.Created(w, contact)
}

// Update godoc
// @Summary Update contact submission
// @Description Update contact submission status
// @Tags contacts
// @Accept json
// @Produce json
// @Param id path string true "Contact Submission ID"
// @Param body body models.UpdateContactRequest true "Contact data"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /api/v1/contacts/{id} [put]
func (h *ContactHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, "Invalid contact ID")
		return
	}

	var req models.UpdateContactRequest
	if err := decodeJSON(r, &req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	contact, err := h.repo.Update(r.Context(), id, &req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.NotFound(w, "Contact submission not found")
			return
		}
		response.InternalError(w, "Failed to update contact submission")
		return
	}

	response.OK(w, contact)
}

// Delete godoc
// @Summary Delete contact submission
// @Description Delete a contact submission
// @Tags contacts
// @Param id path string true "Contact Submission ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /api/v1/contacts/{id} [delete]
func (h *ContactHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, "Invalid contact ID")
		return
	}

	err = h.repo.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.NotFound(w, "Contact submission not found")
			return
		}
		response.InternalError(w, "Failed to delete contact submission")
		return
	}

	response.NoContent(w)
}

// GetUnreadCount godoc
// @Summary Get unread contact count
// @Description Get the count of unread contact submissions
// @Tags contacts
// @Produce json
// @Success 200 {object} response.APIResponse
// @Router /api/v1/contacts/unread-count [get]
func (h *ContactHandler) GetUnreadCount(w http.ResponseWriter, r *http.Request) {
	count, err := h.repo.GetUnreadCount(r.Context())
	if err != nil {
		response.InternalError(w, "Failed to get unread count")
		return
	}

	response.OK(w, map[string]int64{"unread_count": count})
}
