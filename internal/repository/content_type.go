package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/keeps-dev/go-cms-template/internal/models"
)

type ContentTypeRepository struct {
	db *pgxpool.Pool
}

func NewContentTypeRepository(db *pgxpool.Pool) *ContentTypeRepository {
	return &ContentTypeRepository{db: db}
}

func (r *ContentTypeRepository) Create(ctx context.Context, req *models.CreateContentTypeRequest) (*models.ContentType, error) {
	ct := &models.ContentType{
		ID:           uuid.New(),
		Name:         req.Name,
		Slug:         req.Slug,
		SchemaFields: req.SchemaFields,
		IsActive:     true,
		DisplayOrder: 0,
	}

	if req.IsActive != nil {
		ct.IsActive = *req.IsActive
	}
	if req.DisplayOrder != nil {
		ct.DisplayOrder = *req.DisplayOrder
	}

	query := `
		INSERT INTO content_types (id, name, slug, schema_fields, is_active, display_order)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING created_at, updated_at`

	err := r.db.QueryRow(ctx, query,
		ct.ID, ct.Name, ct.Slug, ct.SchemaFields, ct.IsActive, ct.DisplayOrder,
	).Scan(&ct.CreatedAt, &ct.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrDuplicate
		}
		return nil, fmt.Errorf("failed to create content type: %w", err)
	}

	return ct, nil
}

func (r *ContentTypeRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.ContentType, error) {
	query := `
		SELECT id, name, slug, schema_fields, is_active, display_order, created_at, updated_at
		FROM content_types
		WHERE id = $1`

	ct := &models.ContentType{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&ct.ID, &ct.Name, &ct.Slug, &ct.SchemaFields,
		&ct.IsActive, &ct.DisplayOrder, &ct.CreatedAt, &ct.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get content type: %w", err)
	}

	return ct, nil
}

func (r *ContentTypeRepository) GetBySlug(ctx context.Context, slug string) (*models.ContentType, error) {
	query := `
		SELECT id, name, slug, schema_fields, is_active, display_order, created_at, updated_at
		FROM content_types
		WHERE slug = $1`

	ct := &models.ContentType{}
	err := r.db.QueryRow(ctx, query, slug).Scan(
		&ct.ID, &ct.Name, &ct.Slug, &ct.SchemaFields,
		&ct.IsActive, &ct.DisplayOrder, &ct.CreatedAt, &ct.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get content type by slug: %w", err)
	}

	return ct, nil
}

func (r *ContentTypeRepository) List(ctx context.Context, filter models.ContentTypeFilter) ([]models.ContentType, int64, error) {
	filter.PaginationParams.Normalize()

	var conditions []string
	var args []interface{}
	argNum := 1

	if filter.IsActive != nil {
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", argNum))
		args = append(args, *filter.IsActive)
		argNum++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM content_types %s", whereClause)
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count content types: %w", err)
	}

	// Get data
	orderBy := "display_order ASC, created_at DESC"
	if filter.SortBy != "" {
		orderBy = fmt.Sprintf("%s %s", filter.SortBy, filter.SortDir)
	}

	query := fmt.Sprintf(`
		SELECT id, name, slug, schema_fields, is_active, display_order, created_at, updated_at
		FROM content_types
		%s
		ORDER BY %s
		LIMIT $%d OFFSET $%d`,
		whereClause, orderBy, argNum, argNum+1)

	args = append(args, filter.Limit(), filter.Offset())
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list content types: %w", err)
	}
	defer rows.Close()

	var contentTypes []models.ContentType
	for rows.Next() {
		var ct models.ContentType
		if err := rows.Scan(
			&ct.ID, &ct.Name, &ct.Slug, &ct.SchemaFields,
			&ct.IsActive, &ct.DisplayOrder, &ct.CreatedAt, &ct.UpdatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan content type: %w", err)
		}
		contentTypes = append(contentTypes, ct)
	}

	return contentTypes, total, nil
}

func (r *ContentTypeRepository) Update(ctx context.Context, id uuid.UUID, req *models.UpdateContentTypeRequest) (*models.ContentType, error) {
	var setClauses []string
	var args []interface{}
	argNum := 1

	if req.Name != nil {
		setClauses = append(setClauses, fmt.Sprintf("name = $%d", argNum))
		args = append(args, *req.Name)
		argNum++
	}

	if req.Slug != nil {
		setClauses = append(setClauses, fmt.Sprintf("slug = $%d", argNum))
		args = append(args, *req.Slug)
		argNum++
	}

	if req.SchemaFields != nil {
		setClauses = append(setClauses, fmt.Sprintf("schema_fields = $%d", argNum))
		args = append(args, *req.SchemaFields)
		argNum++
	}

	if req.IsActive != nil {
		setClauses = append(setClauses, fmt.Sprintf("is_active = $%d", argNum))
		args = append(args, *req.IsActive)
		argNum++
	}

	if req.DisplayOrder != nil {
		setClauses = append(setClauses, fmt.Sprintf("display_order = $%d", argNum))
		args = append(args, *req.DisplayOrder)
		argNum++
	}

	if len(setClauses) == 0 {
		return r.GetByID(ctx, id)
	}

	args = append(args, id)
	query := fmt.Sprintf(`
		UPDATE content_types
		SET %s
		WHERE id = $%d
		RETURNING id, name, slug, schema_fields, is_active, display_order, created_at, updated_at`,
		strings.Join(setClauses, ", "), argNum)

	ct := &models.ContentType{}
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&ct.ID, &ct.Name, &ct.Slug, &ct.SchemaFields,
		&ct.IsActive, &ct.DisplayOrder, &ct.CreatedAt, &ct.UpdatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrDuplicate
		}
		return nil, fmt.Errorf("failed to update content type: %w", err)
	}

	return ct, nil
}

func (r *ContentTypeRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM content_types WHERE id = $1`
	result, err := r.db.Exec(ctx, query, id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return ErrForeignKey
		}
		return fmt.Errorf("failed to delete content type: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}
