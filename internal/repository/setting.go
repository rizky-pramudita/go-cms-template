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

type SettingRepository struct {
	db *pgxpool.Pool
}

func NewSettingRepository(db *pgxpool.Pool) *SettingRepository {
	return &SettingRepository{db: db}
}

func (r *SettingRepository) Create(ctx context.Context, req *models.CreateSettingRequest) (*models.Setting, error) {
	setting := &models.Setting{
		ID:          uuid.New(),
		Key:         req.Key,
		Value:       req.Value,
		Description: req.Description,
	}

	query := `
		INSERT INTO settings (id, key, value, description)
		VALUES ($1, $2, $3, $4)
		RETURNING updated_at
	`

	err := r.db.QueryRow(ctx, query, setting.ID, setting.Key, setting.Value, setting.Description).Scan(&setting.UpdatedAt)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return nil, ErrDuplicate
		}
		return nil, fmt.Errorf("failed to create setting: %w", err)
	}

	return setting, nil
}

func (r *SettingRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.Setting, error) {
	query := `SELECT id, key, value, description, updated_at FROM settings WHERE id = $1`

	setting := &models.Setting{}
	err := r.db.QueryRow(ctx, query, id).Scan(&setting.ID, &setting.Key, &setting.Value, &setting.Description, &setting.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get setting: %w", err)
	}

	return setting, nil
}

func (r *SettingRepository) GetByKey(ctx context.Context, key string) (*models.Setting, error) {
	query := `SELECT id, key, value, description, updated_at FROM settings WHERE key = $1`

	setting := &models.Setting{}
	err := r.db.QueryRow(ctx, query, key).Scan(&setting.ID, &setting.Key, &setting.Value, &setting.Description, &setting.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get setting by key: %w", err)
	}

	return setting, nil
}

func (r *SettingRepository) List(ctx context.Context, filter models.SettingFilter) ([]models.Setting, int64, error) {
	filter.PaginationParams.Normalize()

	var conditions []string
	var args []interface{}
	argNum := 1

	if filter.Search != "" {
		conditions = append(conditions, fmt.Sprintf("(key ILIKE $%d OR description ILIKE $%d)", argNum, argNum))
		args = append(args, "%"+filter.Search+"%")
		argNum++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM settings %s", whereClause)
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count settings: %w", err)
	}

	// Get data
	orderBy := "key ASC"
	if filter.SortBy != "" {
		orderBy = fmt.Sprintf("%s %s", filter.SortBy, filter.SortDir)
	}

	query := fmt.Sprintf(`
		SELECT id, key, value, description, updated_at
		FROM settings
		%s
		ORDER BY %s
		LIMIT $%d OFFSET $%d
	`, whereClause, orderBy, argNum, argNum+1)

	args = append(args, filter.Limit(), filter.Offset())

	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list settings: %w", err)
	}
	defer rows.Close()

	var settings []models.Setting
	for rows.Next() {
		var setting models.Setting
		if err := rows.Scan(&setting.ID, &setting.Key, &setting.Value, &setting.Description, &setting.UpdatedAt); err != nil {
			return nil, 0, fmt.Errorf("failed to scan setting: %w", err)
		}
		settings = append(settings, setting)
	}

	return settings, total, nil
}

func (r *SettingRepository) Update(ctx context.Context, key string, req *models.UpdateSettingRequest) (*models.Setting, error) {
	var setClauses []string
	var args []interface{}
	argNum := 1

	if req.Value != nil {
		setClauses = append(setClauses, fmt.Sprintf("value = $%d", argNum))
		args = append(args, *req.Value)
		argNum++
	}
	if req.Description != nil {
		setClauses = append(setClauses, fmt.Sprintf("description = $%d", argNum))
		args = append(args, *req.Description)
		argNum++
	}

	if len(setClauses) == 0 {
		return r.GetByKey(ctx, key)
	}

	args = append(args, key)
	query := fmt.Sprintf(`
		UPDATE settings
		SET %s
		WHERE key = $%d
		RETURNING id, key, value, description, updated_at
	`, strings.Join(setClauses, ", "), argNum)

	setting := &models.Setting{}
	err := r.db.QueryRow(ctx, query, args...).Scan(&setting.ID, &setting.Key, &setting.Value, &setting.Description, &setting.UpdatedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to update setting: %w", err)
	}

	return setting, nil
}

func (r *SettingRepository) Upsert(ctx context.Context, req *models.CreateSettingRequest) (*models.Setting, error) {
	setting := &models.Setting{
		ID:          uuid.New(),
		Key:         req.Key,
		Value:       req.Value,
		Description: req.Description,
	}

	query := `
		INSERT INTO settings (id, key, value, description)
		VALUES ($1, $2, $3, $4)
		ON CONFLICT (key) DO UPDATE SET value = EXCLUDED.value, description = EXCLUDED.description
		RETURNING id, updated_at
	`

	err := r.db.QueryRow(ctx, query, setting.ID, setting.Key, setting.Value, setting.Description).Scan(&setting.ID, &setting.UpdatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to upsert setting: %w", err)
	}

	return setting, nil
}

func (r *SettingRepository) Delete(ctx context.Context, key string) error {
	result, err := r.db.Exec(ctx, `DELETE FROM settings WHERE key = $1`, key)
	if err != nil {
		return fmt.Errorf("failed to delete setting: %w", err)
	}

	if result.RowsAffected() == 0 {
		return ErrNotFound
	}

	return nil
}

// GetMultiple returns multiple settings by keys
func (r *SettingRepository) GetMultiple(ctx context.Context, keys []string) (map[string]string, error) {
	query := `SELECT key, value FROM settings WHERE key = ANY($1)`

	rows, err := r.db.Query(ctx, query, keys)
	if err != nil {
		return nil, fmt.Errorf("failed to get multiple settings: %w", err)
	}
	defer rows.Close()

	result := make(map[string]string)
	for rows.Next() {
		var key string
		var value *string
		if err := rows.Scan(&key, &value); err != nil {
			return nil, fmt.Errorf("failed to scan setting: %w", err)
		}
		if value != nil {
			result[key] = *value
		}
	}

	return result, nil
}
