package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"

	"github.com/google/uuid"
	"github.com/keeps-dev/go-cms-template/internal/models"
)

// parseUUID parses a UUID from a string
func parseUUID(s string) (uuid.UUID, error) {
	return uuid.Parse(s)
}

// parsePaginationParams extracts pagination parameters from query string
func parsePaginationParams(r *http.Request) models.PaginationParams {
	params := models.DefaultPagination()

	if page := r.URL.Query().Get("page"); page != "" {
		if p, err := strconv.Atoi(page); err == nil {
			params.Page = p
		}
	}

	if pageSize := r.URL.Query().Get("page_size"); pageSize != "" {
		if ps, err := strconv.Atoi(pageSize); err == nil {
			params.PageSize = ps
		}
	}

	if sortBy := r.URL.Query().Get("sort_by"); sortBy != "" {
		params.SortBy = sortBy
	}

	if sortDir := r.URL.Query().Get("sort_dir"); sortDir != "" {
		params.SortDir = sortDir
	}

	params.Normalize()
	return params
}

// decodeJSON decodes JSON from request body
func decodeJSON(r *http.Request, v interface{}) error {
	return json.NewDecoder(r.Body).Decode(v)
}

// getBoolParam extracts a boolean query parameter
func getBoolParam(r *http.Request, key string) *bool {
	val := r.URL.Query().Get(key)
	if val == "" {
		return nil
	}
	b := val == "true" || val == "1"
	return &b
}

// getIntParam extracts an integer query parameter
func getIntParam(r *http.Request, key string) *int {
	val := r.URL.Query().Get(key)
	if val == "" {
		return nil
	}
	if i, err := strconv.Atoi(val); err == nil {
		return &i
	}
	return nil
}
