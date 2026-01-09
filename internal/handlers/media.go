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

type MediaHandler struct {
	repo *repository.MediaRepository
}

func NewMediaHandler(repo *repository.MediaRepository) *MediaHandler {
	return &MediaHandler{repo: repo}
}

// List godoc
// @Summary List media
// @Description Get all media with optional filtering
// @Tags media
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param file_type query int false "Filter by file type (1=image, 2=video, 3=document)"
// @Param search query string false "Search in file name and alt text"
// @Success 200 {object} response.APIResponse
// @Router /api/v1/media [get]
func (h *MediaHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := models.MediaFilter{
		PaginationParams: parsePaginationParams(r),
		Search:           r.URL.Query().Get("search"),
	}

	if ftStr := r.URL.Query().Get("file_type"); ftStr != "" {
		if ft, err := strconv.Atoi(ftStr); err == nil {
			fileType := models.FileType(ft)
			filter.FileType = &fileType
		}
	}

	mediaList, total, err := h.repo.List(r.Context(), filter)
	if err != nil {
		response.InternalError(w, "Failed to list media")
		return
	}

	response.JSONWithMeta(w, http.StatusOK, mediaList, &response.Meta{
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		Total:      total,
		TotalPages: int(total)/filter.PageSize + 1,
	})
}

// Get godoc
// @Summary Get media by ID
// @Description Get a single media by its ID
// @Tags media
// @Produce json
// @Param id path string true "Media ID"
// @Success 200 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /api/v1/media/{id} [get]
func (h *MediaHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, "Invalid media ID")
		return
	}

	media, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.NotFound(w, "Media not found")
			return
		}
		response.InternalError(w, "Failed to get media")
		return
	}

	response.OK(w, media)
}

// Create godoc
// @Summary Create media record
// @Description Create a new media record (metadata only, file upload handled separately)
// @Tags media
// @Accept json
// @Produce json
// @Param body body models.CreateMediaRequest true "Media data"
// @Success 201 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Router /api/v1/media [post]
func (h *MediaHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreateMediaRequest
	if err := decodeJSON(r, &req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	// Validate required fields
	validationErrors := make(map[string]string)
	if req.FileName == "" {
		validationErrors["file_name"] = "File name is required"
	}
	if req.ObjectKey == "" {
		validationErrors["object_key"] = "Object key is required"
	}
	if req.BucketName == "" {
		validationErrors["bucket_name"] = "Bucket name is required"
	}
	if req.MimeType == "" {
		validationErrors["mime_type"] = "Mime type is required"
	}
	if req.FileSize <= 0 {
		validationErrors["file_size"] = "File size must be positive"
	}
	if req.FileType < 1 || req.FileType > 3 {
		validationErrors["file_type"] = "File type must be 1 (image), 2 (video), or 3 (document)"
	}

	if len(validationErrors) > 0 {
		response.ValidationError(w, validationErrors)
		return
	}

	media, err := h.repo.Create(r.Context(), &req)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			response.Conflict(w, "Media with this object key already exists")
			return
		}
		response.InternalError(w, "Failed to create media")
		return
	}

	response.Created(w, media)
}

// Update godoc
// @Summary Update media
// @Description Update media metadata
// @Tags media
// @Accept json
// @Produce json
// @Param id path string true "Media ID"
// @Param body body models.UpdateMediaRequest true "Media data"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /api/v1/media/{id} [put]
func (h *MediaHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, "Invalid media ID")
		return
	}

	var req models.UpdateMediaRequest
	if err := decodeJSON(r, &req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	media, err := h.repo.Update(r.Context(), id, &req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.NotFound(w, "Media not found")
			return
		}
		response.InternalError(w, "Failed to update media")
		return
	}

	response.OK(w, media)
}

// Delete godoc
// @Summary Delete media
// @Description Delete a media record
// @Tags media
// @Param id path string true "Media ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Failure 409 {object} response.APIResponse
// @Router /api/v1/media/{id} [delete]
func (h *MediaHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, "Invalid media ID")
		return
	}

	err = h.repo.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.NotFound(w, "Media not found")
			return
		}
		if errors.Is(err, repository.ErrForeignKey) {
			response.Conflict(w, "Cannot delete media that is attached to posts")
			return
		}
		response.InternalError(w, "Failed to delete media")
		return
	}

	response.NoContent(w)
}
