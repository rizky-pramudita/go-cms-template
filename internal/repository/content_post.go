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

type ContentPostRepository struct {
	db *pgxpool.Pool
}

func NewContentPostRepository(db *pgxpool.Pool) *ContentPostRepository {
	return &ContentPostRepository{db: db}
}

func (r *ContentPostRepository) Create(ctx context.Context, req *models.CreatePostRequest) (*models.ContentPost, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	post := &models.ContentPost{
		ID:            uuid.New(),
		ContentTypeID: req.ContentTypeID,
		AuthorID:      req.AuthorID,
		Title:         req.Title,
		Slug:          req.Slug,
		Excerpt:       req.Excerpt,
		Content:       req.Content,
		Metadata:      req.Metadata,
		Status:        models.PostStatusDraft,
		PublishedAt:   req.PublishedAt,
	}

	if req.Status != nil {
		post.Status = *req.Status
	}

	query := `
		INSERT INTO content_posts (id, content_type_id, author_id, title, slug, excerpt, content, metadata, status, published_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING view_count, created_at, updated_at
	`

	err = tx.QueryRow(ctx, query,
		post.ID, post.ContentTypeID, post.AuthorID, post.Title, post.Slug,
		post.Excerpt, post.Content, post.Metadata, post.Status, post.PublishedAt,
	).Scan(&post.ViewCount, &post.CreatedAt, &post.UpdatedAt)

	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return nil, ErrDuplicate
			case "23503":
				return nil, ErrForeignKey
			}
		}
		return nil, fmt.Errorf("failed to create post: %w", err)
	}

	// Attach tags if provided
	if len(req.TagIDs) > 0 {
		if err := r.attachTagsTx(ctx, tx, post.ID, req.TagIDs); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return post, nil
}

func (r *ContentPostRepository) attachTagsTx(ctx context.Context, tx pgx.Tx, postID uuid.UUID, tagIDs []uuid.UUID) error {
	for _, tagID := range tagIDs {
		_, err := tx.Exec(ctx,
			`INSERT INTO post_tags (post_id, tag_id) VALUES ($1, $2) ON CONFLICT DO NOTHING`,
			postID, tagID,
		)
		if err != nil {
			return fmt.Errorf("failed to attach tag: %w", err)
		}
	}
	return nil
}

func (r *ContentPostRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.ContentPost, error) {
	query := `
		SELECT cp.id, cp.content_type_id, cp.author_id, cp.title, cp.slug, cp.excerpt, 
		       cp.content, cp.metadata, cp.status, cp.published_at, cp.view_count, 
		       cp.created_at, cp.updated_at,
		       ct.id, ct.name, ct.slug, ct.schema_fields, ct.is_active, ct.display_order, ct.created_at, ct.updated_at,
		       u.id, u.email, u.full_name, u.role, u.is_active, u.last_login, u.created_at, u.updated_at
		FROM content_posts cp
		JOIN content_types ct ON cp.content_type_id = ct.id
		JOIN users u ON cp.author_id = u.id
		WHERE cp.id = $1
	`

	post := &models.ContentPost{
		ContentType: &models.ContentType{},
		Author:      &models.UserResponse{},
	}

	err := r.db.QueryRow(ctx, query, id).Scan(
		&post.ID, &post.ContentTypeID, &post.AuthorID, &post.Title, &post.Slug,
		&post.Excerpt, &post.Content, &post.Metadata, &post.Status, &post.PublishedAt,
		&post.ViewCount, &post.CreatedAt, &post.UpdatedAt,
		&post.ContentType.ID, &post.ContentType.Name, &post.ContentType.Slug,
		&post.ContentType.SchemaFields, &post.ContentType.IsActive, &post.ContentType.DisplayOrder,
		&post.ContentType.CreatedAt, &post.ContentType.UpdatedAt,
		&post.Author.ID, &post.Author.Email, &post.Author.FullName, &post.Author.Role,
		&post.Author.IsActive, &post.Author.LastLogin, &post.Author.CreatedAt, &post.Author.UpdatedAt,
	)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get post: %w", err)
	}

	// Load tags
	tags, err := r.getPostTags(ctx, post.ID)
	if err != nil {
		return nil, err
	}
	post.Tags = tags

	// Load media
	media, err := r.getPostMedia(ctx, post.ID)
	if err != nil {
		return nil, err
	}
	post.Media = media

	return post, nil
}

func (r *ContentPostRepository) GetBySlug(ctx context.Context, slug string) (*models.ContentPost, error) {
	// First get the post ID
	var postID uuid.UUID
	err := r.db.QueryRow(ctx, `SELECT id FROM content_posts WHERE slug = $1`, slug).Scan(&postID)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get post by slug: %w", err)
	}

	return r.GetByID(ctx, postID)
}

func (r *ContentPostRepository) getPostTags(ctx context.Context, postID uuid.UUID) ([]models.Tag, error) {
	query := `
		SELECT t.id, t.name, t.slug, t.created_at
		FROM tags t
		JOIN post_tags pt ON t.id = pt.tag_id
		WHERE pt.post_id = $1
		ORDER BY t.name
	`

	rows, err := r.db.Query(ctx, query, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to get post tags: %w", err)
	}
	defer rows.Close()

	var tags []models.Tag
	for rows.Next() {
		var tag models.Tag
		if err := rows.Scan(&tag.ID, &tag.Name, &tag.Slug, &tag.CreatedAt); err != nil {
			return nil, fmt.Errorf("failed to scan tag: %w", err)
		}
		tags = append(tags, tag)
	}

	return tags, nil
}

func (r *ContentPostRepository) getPostMedia(ctx context.Context, postID uuid.UUID) ([]models.PostMedia, error) {
	query := `
		SELECT pm.id, pm.post_id, pm.media_id, pm.media_role, pm.display_order, pm.created_at,
		       m.id, m.file_name, m.object_key, m.bucket_name, m.cdn_url, m.file_type,
		       m.mime_type, m.file_size, m.dimensions, m.variants, m.alt_text, m.checksum, m.created_at
		FROM post_media pm
		JOIN media m ON pm.media_id = m.id
		WHERE pm.post_id = $1
		ORDER BY pm.display_order
	`

	rows, err := r.db.Query(ctx, query, postID)
	if err != nil {
		return nil, fmt.Errorf("failed to get post media: %w", err)
	}
	defer rows.Close()

	var mediaList []models.PostMedia
	for rows.Next() {
		pm := models.PostMedia{Media: &models.Media{}}
		if err := rows.Scan(
			&pm.ID, &pm.PostID, &pm.MediaID, &pm.MediaRole, &pm.DisplayOrder, &pm.CreatedAt,
			&pm.Media.ID, &pm.Media.FileName, &pm.Media.ObjectKey, &pm.Media.BucketName,
			&pm.Media.CDNUrl, &pm.Media.FileType, &pm.Media.MimeType, &pm.Media.FileSize,
			&pm.Media.Dimensions, &pm.Media.Variants, &pm.Media.AltText, &pm.Media.Checksum,
			&pm.Media.CreatedAt,
		); err != nil {
			return nil, fmt.Errorf("failed to scan post media: %w", err)
		}
		mediaList = append(mediaList, pm)
	}

	return mediaList, nil
}

func (r *ContentPostRepository) List(ctx context.Context, filter models.PostFilter) ([]models.ContentPost, int64, error) {
	filter.PaginationParams.Normalize()

	var conditions []string
	var args []interface{}
	argNum := 1

	if filter.ContentTypeID != nil {
		conditions = append(conditions, fmt.Sprintf("cp.content_type_id = $%d", argNum))
		args = append(args, *filter.ContentTypeID)
		argNum++
	}
	if filter.AuthorID != nil {
		conditions = append(conditions, fmt.Sprintf("cp.author_id = $%d", argNum))
		args = append(args, *filter.AuthorID)
		argNum++
	}
	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("cp.status = $%d", argNum))
		args = append(args, *filter.Status)
		argNum++
	}
	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(cp.title ILIKE $%d OR cp.excerpt ILIKE $%d)", argNum, argNum))
		args = append(args, "%"+filter.Search+"%")
		argNum++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM content_posts cp %s", whereClause)
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count posts: %w", err)
	}

	// Get data
	orderBy := "cp.created_at DESC"
	if filter.SortBy != "" {
		orderBy = fmt.Sprintf("cp.%s %s", filter.SortBy, filter.SortDir)
	}

	query := fmt.Sprintf(`
		SELECT cp.id, cp.content_type_id, cp.author_id, cp.title, cp.slug, cp.excerpt,
		       cp.content, cp.metadata, cp.status, cp.published_at, cp.view_count,
		       cp.created_at, cp.updated_at,
		       ct.name as content_type_name, ct.slug as content_type_slug,
		       u.full_name as author_name
		FROM content_posts cp
		JOIN content_types ct ON cp.content_type_id = ct.id
		JOIN users u ON cp.author_id = u.id
		%s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, argNum, argNum+1)

	args = append(args, filter.Limit(), filter.Offset())

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list posts: %w", err)
	}
	defer rows.Close()

	var posts []models.ContentPost
	for rows.Next() {
		var post models.ContentPost
		var ctName, ctSlug, authorName string

		if err := rows.Scan(
			&post.ID, &post.ContentTypeID, &post.AuthorID, &post.Title, &post.Slug,
			&post.Excerpt, &post.Content, &post.Metadata, &post.Status, &post.PublishedAt,
			&post.ViewCount, &post.CreatedAt, &post.UpdatedAt,
			&ctName, &ctSlug, &authorName,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan post: %w", err)
		}

		// Minimal relations for list view
		post.ContentType = &models.ContentType{ID: post.ContentTypeID, Name: ctName, Slug: ctSlug}
		post.Author = &models.UserResponse{ID: post.AuthorID, FullName: authorName}

		posts = append(posts, post)
	}

	return posts, total, nil
}

func (r *ContentPostRepository) Update(ctx context.Context, id uuid.UUID, req *models.UpdatePostRequest) (*models.ContentPost, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback(ctx)

	var setClauses []string
	var args []interface{}
	argNum := 1

	if req.ContentTypeID != nil {
		setClauses = append(setClauses, fmt.Sprintf("content_type_id = $%d", argNum))
		args = append(args, *req.ContentTypeID)
		argNum++
	}
	if req.Title != nil {
		setClauses = append(setClauses, fmt.Sprintf("title = $%d", argNum))
		args = append(args, *req.Title)
		argNum++
	}
	if req.Slug != nil {
		setClauses = append(setClauses, fmt.Sprintf("slug = $%d", argNum))
		args = append(args, *req.Slug)
		argNum++
	}
	if req.Excerpt != nil {
		setClauses = append(setClauses, fmt.Sprintf("excerpt = $%d", argNum))
		args = append(args, *req.Excerpt)
		argNum++
	}
	if req.Content != nil {
		setClauses = append(setClauses, fmt.Sprintf("content = $%d", argNum))
		args = append(args, *req.Content)
		argNum++
	}
	if req.Metadata != nil {
		setClauses = append(setClauses, fmt.Sprintf("metadata = $%d", argNum))
		args = append(args, *req.Metadata)
		argNum++
	}
	if req.Status != nil {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", argNum))
		args = append(args, *req.Status)
		argNum++
	}
	if req.PublishedAt != nil {
		setClauses = append(setClauses, fmt.Sprintf("published_at = $%d", argNum))
		args = append(args, *req.PublishedAt)
		argNum++
	}

	if len(setClauses) > 0 {
		args = append(args, id)
		query := fmt.Sprintf(`UPDATE content_posts SET %s WHERE id = $%d`, strings.Join(setClauses, ", "), argNum)

		result, err := tx.Exec(ctx, query, args...)
		if err != nil {
			var pgErr *pgconn.PgError
			if errors.As(err, &pgErr) {
				switch pgErr.Code {
				case "23505":
					return nil, ErrDuplicate
				case "23503":
					return nil, ErrForeignKey
				}
			}
			return nil, fmt.Errorf("failed to update post: %w", err)
		}

		if result.RowsAffected() == 0 {
			return nil, ErrNotFound
		}
	}

	// Update tags if provided
	if req.TagIDs != nil {
		// Remove existing tags
		_, err := tx.Exec(ctx, `DELETE FROM post_tags WHERE post_id = $1`, id)
		if err != nil {
			return nil, fmt.Errorf("failed to remove existing tags: %w", err)
		}

		// Add new tags
		if len(*req.TagIDs) > 0 {
			if err := r.attachTagsTx(ctx, tx, id, *req.TagIDs); err != nil {
				return nil, err
			}
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("failed to commit transaction: %w", err)
	}

	return r.GetByID(ctx, id)
}

func (r *ContentPostRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.Exec(ctx, `DELETE FROM content_posts WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete post: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *ContentPostRepository) AttachMedia(ctx context.Context, postID uuid.UUID, req *models.AttachMediaRequest) (*models.PostMedia, error) {
	pm := &models.PostMedia{
		ID:           uuid.New(),
		PostID:       postID,
		MediaID:      req.MediaID,
		MediaRole:    req.MediaRole,
		DisplayOrder: 0,
	}

	if req.DisplayOrder != nil {
		pm.DisplayOrder = *req.DisplayOrder
	}

	query := `
		INSERT INTO post_media (id, post_id, media_id, media_role, display_order)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING created_at
	`

	err := r.db.QueryRow(ctx, query, pm.ID, pm.PostID, pm.MediaID, pm.MediaRole, pm.DisplayOrder).Scan(&pm.CreatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) {
			switch pgErr.Code {
			case "23505":
				return nil, ErrDuplicate
			case "23503":
				return nil, ErrForeignKey
			}
		}
		return nil, fmt.Errorf("failed to attach media: %w", err)
	}

	return pm, nil
}

func (r *ContentPostRepository) DetachMedia(ctx context.Context, postID, mediaID uuid.UUID) error {
	result, err := r.db.Exec(ctx, `DELETE FROM post_media WHERE post_id = $1 AND media_id = $2`, postID, mediaID)
	if err != nil {
		return fmt.Errorf("failed to detach media: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

func (r *ContentPostRepository) IncrementViewCount(ctx context.Context, id uuid.UUID) error {
	_, err := r.db.Exec(ctx, `UPDATE content_posts SET view_count = view_count + 1 WHERE id = $1`, id)
	return err
}
