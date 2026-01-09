package handlers

import (
	"errors"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/keeps-dev/go-cms-template/internal/models"
	"github.com/keeps-dev/go-cms-template/internal/repository"
	"github.com/keeps-dev/go-cms-template/internal/response"
)

type SettingHandler struct {
	repo *repository.SettingRepository
}

func NewSettingHandler(repo *repository.SettingRepository) *SettingHandler {
	return &SettingHandler{repo: repo}
}

// List godoc
// @Summary List settings
// @Description Get all settings with optional search
// @Tags settings
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param search query string false "Search in key and description"
// @Success 200 {object} response.APIResponse
// @Router /api/v1/settings [get]
func (h *SettingHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := models.SettingFilter{
		PaginationParams: parsePaginationParams(r),
		Search:           r.URL.Query().Get("search"),
	}

	settings, total, err := h.repo.List(r.Context(), filter)
	if err != nil {
		response.InternalError(w, "Failed to list settings")
		return
	}

	response.JSONWithMeta(w, http.StatusOK, settings, &response.Meta{
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		Total:      total,
		TotalPages: int(total)/filter.PageSize + 1,
	})
}

// Get godoc
// @Summary Get setting by key
// @Description Get a single setting by its key
// @Tags settings
// @Produce json
// @Param key path string true "Setting Key"
// @Success 200 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /api/v1/settings/{key} [get]
func (h *SettingHandler) Get(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	if key == "" {
		response.BadRequest(w, "Key is required")
		return
	}

	setting, err := h.repo.GetByKey(r.Context(), key)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.NotFound(w, "Setting not found")
			return
		}
		response.InternalError(w, "Failed to get setting")
		return
	}

	response.OK(w, setting)
}

// Create godoc
// @Summary Create setting
// @Description Create a new setting
// @Tags settings
// @Accept json
// @Produce json
// @Param body body models.CreateSettingRequest true "Setting data"
// @Success 201 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 409 {object} response.APIResponse
// @Router /api/v1/settings [post]
func (h *SettingHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateSettingRequest
	if err := decodeJSON(r, &req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	// Validate required fields
	if req.Key == "" {
		response.ValidationError(w, map[string]string{
			"key": "Key is required",
		})
		return
	}

	setting, err := h.repo.Create(r.Context(), &req)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			response.Conflict(w, "Setting with this key already exists")
			return
		}
		response.InternalError(w, "Failed to create setting")
		return
	}

	response.Created(w, setting)
}

// Update godoc
// @Summary Update setting
// @Description Update an existing setting by key
// @Tags settings
// @Accept json
// @Produce json
// @Param key path string true "Setting Key"
// @Param body body models.UpdateSettingRequest true "Setting data"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /api/v1/settings/{key} [put]
func (h *SettingHandler) Update(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	if key == "" {
		response.BadRequest(w, "Key is required")
		return
	}

	var req models.UpdateSettingRequest
	if err := decodeJSON(r, &req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	setting, err := h.repo.Update(r.Context(), key, &req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.NotFound(w, "Setting not found")
			return
		}
		response.InternalError(w, "Failed to update setting")
		return
	}

	response.OK(w, setting)
}

// Upsert godoc
// @Summary Upsert setting
// @Description Create or update a setting
// @Tags settings
// @Accept json
// @Produce json
// @Param body body models.CreateSettingRequest true "Setting data"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Router /api/v1/settings/upsert [post]
func (h *SettingHandler) Upsert(w http.ResponseWriter, r *http.Request) {
	var req models.CreateSettingRequest
	if err := decodeJSON(r, &req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if req.Key == "" {
		response.ValidationError(w, map[string]string{
			"key": "Key is required",
		})
		return
	}

	setting, err := h.repo.Upsert(r.Context(), &req)
	if err != nil {
		response.InternalError(w, "Failed to upsert setting")
		return
	}

	response.OK(w, setting)
}

// Delete godoc
// @Summary Delete setting
// @Description Delete a setting by key
// @Tags settings
// @Param key path string true "Setting Key"
// @Success 204 "No Content"
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /api/v1/settings/{key} [delete]
func (h *SettingHandler) Delete(w http.ResponseWriter, r *http.Request) {
	key := chi.URLParam(r, "key")
	if key == "" {
		response.BadRequest(w, "Key is required")
		return
	}

	err := h.repo.Delete(r.Context(), key)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.NotFound(w, "Setting not found")
			return
		}
		response.InternalError(w, "Failed to delete setting")
		return
	}

	response.NoContent(w)
}

// GetMultiple godoc
// @Summary Get multiple settings
// @Description Get multiple settings by their keys
// @Tags settings
// @Accept json
// @Produce json
// @Param body body []string true "Array of setting keys"
// @Success 200 {object} response.APIResponse
// @Router /api/v1/settings/bulk [post]
func (h *SettingHandler) GetMultiple(w http.ResponseWriter, r *http.Request) {
	var keys []string
	if err := decodeJSON(r, &keys); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if len(keys) == 0 {
		response.BadRequest(w, "At least one key is required")
		return
	}

	settings, err := h.repo.GetMultiple(r.Context(), keys)
	if err != nil {
		response.InternalError(w, "Failed to get settings")
		return
	}

	response.OK(w, settings)
}
