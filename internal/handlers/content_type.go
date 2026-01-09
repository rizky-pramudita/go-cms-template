package handlers

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/keeps-dev/go-cms-template/internal/models"
	"github.com/keeps-dev/go-cms-template/internal/repository"
	"github.com/keeps-dev/go-cms-template/internal/response"
)

type ContentTypeHandler struct {
	repo *repository.ContentTypeRepository
}

func NewContentTypeHandler(repo *repository.ContentTypeRepository) *ContentTypeHandler {
	return &ContentTypeHandler{repo: repo}
}

// List godoc
// @Summary List content types
// @Description Get all content types with optional filtering
// @Tags content-types
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(20)
// @Param is_active query bool false "Filter by active status"
// @Success 200 {object} response.APIResponse
// @Router /api/v1/content-types [get]
func (h *ContentTypeHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := models.ContentTypeFilter{
		IsActive:         getBoolParam(r, "is_active"),
		PaginationParams: parsePaginationParams(r),
	}

	contentTypes, total, err := h.repo.List(r.Context(), filter)
	if err != nil {
		response.InternalError(w, "Failed to list content types")
		return
	}

	response.JSONWithMeta(w, http.StatusOK, contentTypes, &response.Meta{
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		Total:      total,
		TotalPages: int(total)/filter.PageSize + 1,
	})
}

// Get godoc
// @Summary Get content type by ID
// @Description Get a single content type by its ID
// @Tags content-types
// @Produce json
// @Param id path string true "Content Type ID"
// @Success 200 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /api/v1/content-types/{id} [get]
func (h *ContentTypeHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, "Invalid content type ID")
		return
	}

	contentType, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.NotFound(w, "Content type not found")
			return
		}
		response.InternalError(w, "Failed to get content type")
		return
	}

	response.OK(w, contentType)
}

// GetBySlug godoc
// @Summary Get content type by slug
// @Description Get a single content type by its slug
// @Tags content-types
// @Produce json
// @Param slug path string true "Content Type Slug"
// @Success 200 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /api/v1/content-types/slug/{slug} [get]
func (h *ContentTypeHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		response.BadRequest(w, "Slug is required")
		return
	}

	contentType, err := h.repo.GetBySlug(r.Context(), slug)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.NotFound(w, "Content type not found")
			return
		}
		response.InternalError(w, "Failed to get content type")
		return
	}

	response.OK(w, contentType)
}

// Create godoc
// @Summary Create content type
// @Description Create a new content type
// @Tags content-types
// @Accept json
// @Produce json
// @Param body body models.CreateContentTypeRequest true "Content Type data"
// @Success 201 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 409 {object} response.APIResponse
// @Router /api/v1/content-types [post]
func (h *ContentTypeHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateContentTypeRequest
	if err := decodeJSON(r, &req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	// Validate required fields
	if req.Name == "" || req.Slug == "" {
		response.ValidationError(w, map[string]string{
			"name": "Name is required",
			"slug": "Slug is required",
		})
		return
	}

	contentType, err := h.repo.Create(r.Context(), &req)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			response.Conflict(w, "Content type with this name or slug already exists")
			return
		}
		response.InternalErrorWithErr(w, "Failed to create content type", err)
		return
	}

	response.Created(w, contentType)
}

// Update godoc
// @Summary Update content type
// @Description Update an existing content type
// @Tags content-types
// @Accept json
// @Produce json
// @Param id path string true "Content Type ID"
// @Param body body models.UpdateContentTypeRequest true "Content Type data"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 409 {object} response.APIResponse
// @Router /api/v1/content-types/{id} [put]
func (h *ContentTypeHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, "Invalid content type ID")
		return
	}

	var req models.UpdateContentTypeRequest
	if err := decodeJSON(r, &req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	contentType, err := h.repo.Update(r.Context(), id, &req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.NotFound(w, "Content type not found")
			return
		}
		if errors.Is(err, repository.ErrDuplicate) {
			response.Conflict(w, "Content type with this name or slug already exists")
			return
		}
		response.InternalError(w, "Failed to update content type")
		return
	}

	response.OK(w, contentType)
}

// Delete godoc
// @Summary Delete content type
// @Description Delete a content type
// @Tags content-types
// @Param id path string true "Content Type ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 409 {object} response.APIResponse
// @Router /api/v1/content-types/{id} [delete]
func (h *ContentTypeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, "Invalid content type ID")
		return
	}

	err = h.repo.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.NotFound(w, "Content type not found")
			return
		}
		if errors.Is(err, repository.ErrForeignKey) {
			response.Conflict(w, "Cannot delete content type with existing posts")
			return
		}
		response.InternalError(w, "Failed to delete content type")
		return
	}

	response.NoContent(w)
}
