package handlers

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/keeps-dev/go-cms-template/internal/models"
	"github.com/keeps-dev/go-cms-template/internal/repository"
	"github.com/keeps-dev/go-cms-template/internal/response"
)

type TagHandler struct {
	repo *repository.TagRepository
}

func NewTagHandler(repo *repository.TagRepository) *TagHandler {
	return &TagHandler{repo: repo}
}

// List godoc
// @Summary List tags
// @Description Get all tags with optional search
// @Tags tags
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param search query string false "Search in name and slug"
// @Success 200 {object} response.APIResponse
// @Router /api/v1/tags [get]
func (h *TagHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := models.TagFilter{
		PaginationParams: parsePaginationParams(r),
		Search:           r.URL.Query().Get("search"),
	}

	tags, total, err := h.repo.List(r.Context(), filter)
	if err != nil {
		response.InternalError(w, "Failed to list tags")
		return
	}

	response.JSONWithMeta(w, http.StatusOK, tags, &response.Meta{
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		Total:      total,
		TotalPages: int(total)/filter.PageSize + 1,
	})
}

// Get godoc
// @Summary Get tag by ID
// @Description Get a single tag by its ID
// @Tags tags
// @Produce json
// @Param id path string true "Tag ID"
// @Success 200 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /api/v1/tags/{id} [get]
func (h *TagHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, "Invalid tag ID")
		return
	}

	tag, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.NotFound(w, "Tag not found")
			return
		}
		response.InternalError(w, "Failed to get tag")
		return
	}

	response.OK(w, tag)
}

// GetBySlug godoc
// @Summary Get tag by slug
// @Description Get a single tag by its slug
// @Tags tags
// @Produce json
// @Param slug path string true "Tag Slug"
// @Success 200 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /api/v1/tags/slug/{slug} [get]
func (h *TagHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		response.BadRequest(w, "Slug is required")
		return
	}

	tag, err := h.repo.GetBySlug(r.Context(), slug)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.NotFound(w, "Tag not found")
			return
		}
		response.InternalError(w, "Failed to get tag")
		return
	}

	response.OK(w, tag)
}

// Create godoc
// @Summary Create tag
// @Description Create a new tag
// @Tags tags
// @Accept json
// @Produce json
// @Param body body models.CreateTagRequest true "Tag data"
// @Success 201 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 409 {object} response.APIResponse
// @Router /api/v1/tags [post]
func (h *TagHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateTagRequest
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

	tag, err := h.repo.Create(r.Context(), &req)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			response.Conflict(w, "Tag with this name or slug already exists")
			return
		}
		response.InternalError(w, "Failed to create tag")
		return
	}

	response.Created(w, tag)
}

// Update godoc
// @Summary Update tag
// @Description Update an existing tag
// @Tags tags
// @Accept json
// @Produce json
// @Param id path string true "Tag ID"
// @Param body body models.UpdateTagRequest true "Tag data"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 409 {object} response.APIResponse
// @Router /api/v1/tags/{id} [put]
func (h *TagHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, "Invalid tag ID")
		return
	}

	var req models.UpdateTagRequest
	if err := decodeJSON(r, &req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	tag, err := h.repo.Update(r.Context(), id, &req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.NotFound(w, "Tag not found")
			return
		}
		if errors.Is(err, repository.ErrDuplicate) {
			response.Conflict(w, "Tag with this name or slug already exists")
			return
		}
		response.InternalError(w, "Failed to update tag")
		return
	}

	response.OK(w, tag)
}

// Delete godoc
// @Summary Delete tag
// @Description Delete a tag
// @Tags tags
// @Param id path string true "Tag ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /api/v1/tags/{id} [delete]
func (h *TagHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, "Invalid tag ID")
		return
	}

	err = h.repo.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.NotFound(w, "Tag not found")
			return
		}
		response.InternalError(w, "Failed to delete tag")
		return
	}

	response.NoContent(w)
}
