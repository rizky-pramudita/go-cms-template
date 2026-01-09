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

type MediaRepository struct {
	db *pgxpool.Pool
}

func NewMediaRepository(db *pgxpool.Pool) *MediaRepository {
	return &MediaRepository{db: db}
}

func (r *MediaRepository) Create(ctx context.Context, req *models.CreateMediaRequest) (*models.Media, error) {
	media := &models.Media{
		ID:         uuid.New(),
		FileName:   req.FileName,
		ObjectKey:  req.ObjectKey,
		BucketName: req.BucketName,
		CDNUrl:     req.CDNUrl,
		FileType:   req.FileType,
		MimeType:   req.MimeType,
		FileSize:   req.FileSize,
		Dimensions: req.Dimensions,
		Variants:   req.Variants,
		AltText:    req.AltText,
		Checksum:   req.Checksum,
	}

	query := `
		INSERT INTO media (id, file_name, object_key, bucket_name, cdn_url, file_type, 
		                   mime_type, file_size, dimensions, variants, alt_text, checksum)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12)
		RETURNING created_at
	`

	err := r.db.QueryRow(ctx, query,
		media.ID, media.FileName, media.ObjectKey, media.BucketName, media.CDNUrl,
		media.FileType, media.MimeType, media.FileSize, media.Dimensions, media.Variants,
		media.AltText, media.Checksum,
	).Scan(&media.CreatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrDuplicate
		}
		return nil, fmt.Errorf("failed to create media: %w", err)
	}

	return media, nil
}

func (r *MediaRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Media, error) {
	query := `
		SELECT id, file_name, object_key, bucket_name, cdn_url, file_type, 
		       mime_type, file_size, dimensions, variants, alt_text, checksum, created_at
		FROM media
		WHERE id = $1
	`

	media := &models.Media{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&media.ID, &media.FileName, &media.ObjectKey, &media.BucketName, &media.CDNUrl,
		&media.FileType, &media.MimeType, &media.FileSize, &media.Dimensions, &media.Variants,
		&media.AltText, &media.Checksum, &media.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get media: %w", err)
	}

	return media, nil
}

func (r *MediaRepository) GetByObjectKey(ctx context.Context, objectKey string) (*models.Media, error) {
	query := `
		SELECT id, file_name, object_key, bucket_name, cdn_url, file_type, 
		       mime_type, file_size, dimensions, variants, alt_text, checksum, created_at
		FROM media
		WHERE object_key = $1
	`

	media := &models.Media{}
	err := r.db.QueryRow(ctx, query, objectKey).Scan(
		&media.ID, &media.FileName, &media.ObjectKey, &media.BucketName, &media.CDNUrl,
		&media.FileType, &media.MimeType, &media.FileSize, &media.Dimensions, &media.Variants,
		&media.AltText, &media.Checksum, &media.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get media by object key: %w", err)
	}

	return media, nil
}

func (r *MediaRepository) List(ctx context.Context, filter models.MediaFilter) ([]models.Media, int64, error) {
	filter.PaginationParams.Normalize()

	var conditions []string
	var args []interface{}
	argNum := 1

	if filter.FileType != nil {
		conditions = append(conditions, fmt.Sprintf("file_type = $%d", argNum))
		args = append(args, *filter.FileType)
		argNum++
	}
	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(file_name ILIKE $%d OR alt_text ILIKE $%d)", argNum, argNum))
		args = append(args, "%"+filter.Search+"%")
		argNum++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM media %s", whereClause)
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count media: %w", err)
	}

	// Get data
	orderBy := "created_at DESC"
	if filter.SortBy != "" {
		orderBy = fmt.Sprintf("%s %s", filter.SortBy, filter.SortDir)
	}

	query := fmt.Sprintf(`
		SELECT id, file_name, object_key, bucket_name, cdn_url, file_type, 
		       mime_type, file_size, dimensions, variants, alt_text, checksum, created_at
		FROM media
		%s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, argNum, argNum+1)

	args = append(args, filter.Limit(), filter.Offset())

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list media: %w", err)
	}
	defer rows.Close()

	var mediaList []models.Media
	for rows.Next() {
		var media models.Media
		if err := rows.Scan(
			&media.ID, &media.FileName, &media.ObjectKey, &media.BucketName, &media.CDNUrl,
			&media.FileType, &media.MimeType, &media.FileSize, &media.Dimensions, &media.Variants,
			&media.AltText, &media.Checksum, &media.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan media: %w", err)
		}
		mediaList = append(mediaList, media)
	}

	return mediaList, total, nil
}

func (r *MediaRepository) Update(ctx context.Context, id uuid.UUID, req *models.UpdateMediaRequest) (*models.Media, error) {
	var setClauses []string
	var args []interface{}
	argNum := 1

	if req.FileName != nil {
		setClauses = append(setClauses, fmt.Sprintf("file_name = $%d", argNum))
		args = append(args, *req.FileName)
		argNum++
	}
	if req.CDNUrl != nil {
		setClauses = append(setClauses, fmt.Sprintf("cdn_url = $%d", argNum))
		args = append(args, *req.CDNUrl)
		argNum++
	}
	if req.Dimensions != nil {
		setClauses = append(setClauses, fmt.Sprintf("dimensions = $%d", argNum))
		args = append(args, *req.Dimensions)
		argNum++
	}
	if req.Variants != nil {
		setClauses = append(setClauses, fmt.Sprintf("variants = $%d", argNum))
		args = append(args, *req.Variants)
		argNum++
	}
	if req.AltText != nil {
		setClauses = append(setClauses, fmt.Sprintf("alt_text = $%d", argNum))
		args = append(args, *req.AltText)
		argNum++
	}

	if len(setClauses) == 0 {
		return r.GetByID(ctx, id)
	}

	args = append(args, id)
	query := fmt.Sprintf(`
		UPDATE media
		SET %s
		WHERE id = $%d
		RETURNING id, file_name, object_key, bucket_name, cdn_url, file_type, 
		          mime_type, file_size, dimensions, variants, alt_text, checksum, created_at
	`, strings.Join(setClauses, ", "), argNum)

	media := &models.Media{}
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&media.ID, &media.FileName, &media.ObjectKey, &media.BucketName, &media.CDNUrl,
		&media.FileType, &media.MimeType, &media.FileSize, &media.Dimensions, &media.Variants,
		&media.AltText, &media.Checksum, &media.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to update media: %w", err)
	}

	return media, nil
}

func (r *MediaRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.Exec(ctx, `DELETE FROM media WHERE id = $1`, id)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23503" {
			return ErrForeignKey
		}
		return fmt.Errorf("failed to delete media: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *MediaRepository) GetByChecksum(ctx context.Context, checksum string) (*models.Media, error) {
	query := `
		SELECT id, file_name, object_key, bucket_name, cdn_url, file_type, 
		       mime_type, file_size, dimensions, variants, alt_text, checksum, created_at
		FROM media
		WHERE checksum = $1
	`

	media := &models.Media{}
	err := r.db.QueryRow(ctx, query, checksum).Scan(
		&media.ID, &media.FileName, &media.ObjectKey, &media.BucketName, &media.CDNUrl,
		&media.FileType, &media.MimeType, &media.FileSize, &media.Dimensions, &media.Variants,
		&media.AltText, &media.Checksum, &media.CreatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get media by checksum: %w", err)
	}

	return media, nil
}
