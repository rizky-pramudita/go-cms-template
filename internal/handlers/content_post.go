package handlers

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/keeps-dev/go-cms-template/internal/models"
	"github.com/keeps-dev/go-cms-template/internal/repository"
	"github.com/keeps-dev/go-cms-template/internal/response"
)

type ContentPostHandler struct {
	repo *repository.ContentPostRepository
}

func NewContentPostHandler(repo *repository.ContentPostRepository) *ContentPostHandler {
	return &ContentPostHandler{repo: repo}
}

// List godoc
// @Summary List posts
// @Description Get all posts with optional filtering
// @Tags posts
// @Produce json
// @Param page query int false "Page number"
// @Param page_size query int false "Page size"
// @Param content_type_id query string false "Filter by content type ID"
// @Param author_id query string false "Filter by author ID"
// @Param status query int false "Filter by status (1=draft, 2=published, 3=archived)"
// @Param search query string false "Search in title and excerpt"
// @Success 200 {object} response.APIResponse
// @Router /api/v1/posts [get]
func (h *ContentPostHandler) List(w http.ResponseWriter, r *http.Request) {
	filter := models.PostFilter{
		PaginationParams: parsePaginationParams(r),
		Search:           r.URL.Query().Get("search"),
	}

	if ctID := r.URL.Query().Get("content_type_id"); ctID != "" {
		if id, err := uuid.Parse(ctID); err == nil {
			filter.ContentTypeID = &id
		}
	}

	if authorID := r.URL.Query().Get("author_id"); authorID != "" {
		if id, err := uuid.Parse(authorID); err == nil {
			filter.AuthorID = &id
		}
	}

	if statusStr := r.URL.Query().Get("status"); statusStr != "" {
		if s, err := strconv.Atoi(statusStr); err == nil {
			status := models.PostStatus(s)
			filter.Status = &status
		}
	}

	posts, total, err := h.repo.List(r.Context(), filter)
	if err != nil {
		response.InternalError(w, "Failed to list posts")
		return
	}

	response.JSONWithMeta(w, http.StatusOK, posts, &response.Meta{
		Page:       filter.Page,
		PageSize:   filter.PageSize,
		Total:      total,
		TotalPages: int(total)/filter.PageSize + 1,
	})
}

// Get godoc
// @Summary Get post by ID
// @Description Get a single post by its ID with all relations
// @Tags posts
// @Produce json
// @Param id path string true "Post ID"
// @Success 200 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /api/v1/posts/{id} [get]
func (h *ContentPostHandler) Get(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, "Invalid post ID")
		return
	}

	post, err := h.repo.GetByID(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.NotFound(w, "Post not found")
			return
		}
		response.InternalError(w, "Failed to get post")
		return
	}

	response.OK(w, post)
}

// GetBySlug godoc
// @Summary Get post by slug
// @Description Get a single post by its slug (for public access)
// @Tags posts
// @Produce json
// @Param slug path string true "Post Slug"
// @Success 200 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /api/v1/posts/slug/{slug} [get]
func (h *ContentPostHandler) GetBySlug(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if slug == "" {
		response.BadRequest(w, "Slug is required")
		return
	}

	post, err := h.repo.GetBySlug(r.Context(), slug)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.NotFound(w, "Post not found")
			return
		}
		response.InternalError(w, "Failed to get post")
		return
	}

	// Increment view count asynchronously
	go func() {
		_ = h.repo.IncrementViewCount(r.Context(), post.ID)
	}()

	response.OK(w, post)
}

// Create godoc
// @Summary Create post
// @Description Create a new post
// @Tags posts
// @Accept json
// @Produce json
// @Param body body models.CreatePostRequest true "Post data"
// @Success 201 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Router /api/v1/posts [post]
func (h *ContentPostHandler) Create(w http.ResponseWriter, r *http.Request) {
	var req models.CreatePostRequest
	if err := decodeJSON(r, &req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	// Validate required fields
	validationErrors := make(map[string]string)
	if req.Title == "" {
		validationErrors["title"] = "Title is required"
	}
	if req.Slug == "" {
		validationErrors["slug"] = "Slug is required"
	}
	if req.ContentTypeID == uuid.Nil {
		validationErrors["content_type_id"] = "Content type ID is required"
	}
	if req.AuthorID == uuid.Nil {
		validationErrors["author_id"] = "Author ID is required"
	}

	if len(validationErrors) > 0 {
		response.ValidationError(w, validationErrors)
		return
	}

	post, err := h.repo.Create(r.Context(), &req)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			response.Conflict(w, "Post with this slug already exists")
			return
		}
		if errors.Is(err, repository.ErrForeignKey) {
			response.BadRequest(w, "Invalid content type ID or author ID")
			return
		}
		response.InternalErrorWithErr(w, "Failed to create post", err)
		return
	}

	response.Created(w, post)
}

// Update godoc
// @Summary Update post
// @Description Update an existing post
// @Tags posts
// @Accept json
// @Produce json
// @Param id path string true "Post ID"
// @Param body body models.UpdatePostRequest true "Post data"
// @Success 200 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /api/v1/posts/{id} [put]
func (h *ContentPostHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, "Invalid post ID")
		return
	}

	var req models.UpdatePostRequest
	if err := decodeJSON(r, &req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	post, err := h.repo.Update(r.Context(), id, &req)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.NotFound(w, "Post not found")
			return
		}
		if errors.Is(err, repository.ErrDuplicate) {
			response.Conflict(w, "Post with this slug already exists")
			return
		}
		if errors.Is(err, repository.ErrForeignKey) {
			response.BadRequest(w, "Invalid content type ID")
			return
		}
		response.InternalError(w, "Failed to update post")
		return
	}

	response.OK(w, post)
}

// Delete godoc
// @Summary Delete post
// @Description Delete a post
// @Tags posts
// @Param id path string true "Post ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /api/v1/posts/{id} [delete]
func (h *ContentPostHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, "Invalid post ID")
		return
	}

	err = h.repo.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.NotFound(w, "Post not found")
			return
		}
		response.InternalError(w, "Failed to delete post")
		return
	}

	response.NoContent(w)
}

// AttachMedia godoc
// @Summary Attach media to post
// @Description Attach a media file to a post
// @Tags posts
// @Accept json
// @Produce json
// @Param id path string true "Post ID"
// @Param body body models.AttachMediaRequest true "Media attachment data"
// @Success 201 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Router /api/v1/posts/{id}/media [post]
func (h *ContentPostHandler) AttachMedia(w http.ResponseWriter, r *http.Request) {
	postID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, "Invalid post ID")
		return
	}

	var req models.AttachMediaRequest
	if err := decodeJSON(r, &req); err != nil {
		response.BadRequest(w, "Invalid request body")
		return
	}

	if req.MediaID == uuid.Nil {
		response.ValidationError(w, map[string]string{"media_id": "Media ID is required"})
		return
	}

	postMedia, err := h.repo.AttachMedia(r.Context(), postID, &req)
	if err != nil {
		if errors.Is(err, repository.ErrDuplicate) {
			response.Conflict(w, "Media already attached to this post")
			return
		}
		if errors.Is(err, repository.ErrForeignKey) {
			response.BadRequest(w, "Invalid post ID or media ID")
			return
		}
		response.InternalError(w, "Failed to attach media")
		return
	}

	response.Created(w, postMedia)
}

// DetachMedia godoc
// @Summary Detach media from post
// @Description Remove a media attachment from a post
// @Tags posts
// @Param id path string true "Post ID"
// @Param mediaId path string true "Media ID"
// @Success 204 "No Content"
// @Failure 400 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /api/v1/posts/{id}/media/{mediaId} [delete]
func (h *ContentPostHandler) DetachMedia(w http.ResponseWriter, r *http.Request) {
	postID, err := parseUUID(chi.URLParam(r, "id"))
	if err != nil {
		response.BadRequest(w, "Invalid post ID")
		return
	}

	mediaID, err := parseUUID(chi.URLParam(r, "mediaId"))
	if err != nil {
		response.BadRequest(w, "Invalid media ID")
		return
	}

	err = h.repo.DetachMedia(r.Context(), postID, mediaID)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			response.NotFound(w, "Media attachment not found")
			return
		}
		response.InternalError(w, "Failed to detach media")
		return
	}

	response.NoContent(w)
}
