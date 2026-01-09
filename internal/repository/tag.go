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

type TagRepository struct {
	db *pgxpool.Pool
}

func NewTagRepository(db *pgxpool.Pool) *TagRepository {
	return &TagRepository{db: db}
}

func (r *TagRepository) Create(ctx context.Context, req *models.CreateTagRequest) (*models.Tag, error) {
	tag := &models.Tag{
		ID:   uuid.New(),
		Name: req.Name,
		Slug: req.Slug,
	}

	query := `
		INSERT INTO tags (id, name, slug)
		VALUES ($1, $2, $3)
		RETURNING created_at`

	err := r.db.QueryRow(ctx, query, tag.ID, tag.Name, tag.Slug).Scan(&tag.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrDuplicate
		}
		return nil, fmt.Errorf("failed to create tag: %w", err)
	}

	return tag, nil
}

func (r *TagRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Tag, error) {
	query := `SELECT id, name, slug, created_at FROM tags WHERE id = $1`
	tag := &models.Tag{}
	err := r.db.QueryRow(ctx, query, id).Scan(&tag.ID, &tag.Name, &tag.Slug, &tag.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get tag: %w", err)
	}
	return tag, nil
}

func (r *TagRepository) GetBySlug(ctx context.Context, slug string) (*models.Tag, error) {
	query := `SELECT id, name, slug, created_at FROM tags WHERE slug = $1`
	tag := &models.Tag{}
	err := r.db.QueryRow(ctx, query, slug).Scan(&tag.ID, &tag.Name, &tag.Slug, &tag.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get tag by slug: %w", err)
	}
	return tag, nil
}

func (r *TagRepository) List(ctx context.Context, filter models.TagFilter) ([]models.Tag, int64, error) {
	filter.PaginationParams.Normalize()

	var conditions []string
	var args []interface{}
	argNum := 1

	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(name ILIKE $%d OR slug ILIKE $%d)", argNum, argNum))
		args = append(args, "%"+filter.Search+"%")
		argNum++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM tags %s", whereClause)
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count tags: %w", err)
	}

	// Get data
	orderBy := "name ASC"
	if filter.SortBy != "" {
		orderBy = fmt.Sprintf("%s %s", filter.SortBy, filter.SortDir)
	}

	query := fmt.Sprintf(`
		SELECT id, name, slug, created_at
		FROM tags
		%s
		ORDER BY %s
		LIMIT $%d OFFSET $%d`,
		whereClause, orderBy, argNum, argNum+1)

	args = append(args, filter.Limit(), filter.Offset())
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list tags: %w", err)
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var tag models.Tag
		if err := rows.Scan(&tag.ID, &tag.Name, &tag.Slug, &tag.CreatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}

	return tags, total, nil
}

func (r *TagRepository) Update(ctx context.Context, id uuid.UUID, req *models.UpdateTagRequest) (*models.Tag, error) {
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

	if len(setClauses) == 0 {
		return r.GetByID(ctx, id)
	}

	args = append(args, id)
	query := fmt.Sprintf(`
		UPDATE tags
		SET %s
		WHERE id = $%d
		RETURNING id, name, slug, created_at`,
		strings.Join(setClauses, ", "), argNum)

	tag := &models.Tag{}
	err := r.db.QueryRow(ctx, query, args...).Scan(&tag.ID, &tag.Name, &tag.Slug, &tag.CreatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrDuplicate
		}
		return nil, fmt.Errorf("failed to update tag: %w", err)
	}

	return tag, nil
}

func (r *TagRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.Exec(ctx, `DELETE FROM tags WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete tag: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *TagRepository) GetPostCountByTag(ctx context.Context, tagID uuid.UUID) (int64, error) {
	var count int64
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM post_tags WHERE tag_id = $1`, tagID).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get post count: %w", err)
	}
	return count, nil
}
