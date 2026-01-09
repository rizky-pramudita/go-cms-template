package repository

import (
	"context"
	"errors"
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/keeps-dev/go-cms-template/internal/models"
)

type ContactRepository struct {
	db *pgxpool.Pool
}

func NewContactRepository(db *pgxpool.Pool) *ContactRepository {
	return &ContactRepository{db: db}
}

func (r *ContactRepository) Create(ctx context.Context, req *models.CreateContactRequest) (*models.ContactSubmission, error) {
	var ipAddr *net.IP
	if req.IPAddress != nil {
		ip := net.ParseIP(*req.IPAddress)
		if ip != nil {
			ipAddr = &ip
		}
	}

	contact := &models.ContactSubmission{
		ID:        uuid.New(),
		Name:      req.Name,
		Email:     req.Email,
		Phone:     req.Phone,
		Subject:   req.Subject,
		Message:   req.Message,
		Status:    models.ContactStatusNew,
		IPAddress: ipAddr,
		UserAgent: req.UserAgent,
		Metadata:  req.Metadata,
	}

	query := `
		INSERT INTO contact_submissions (id, name, email, phone, subject, message, status, ip_address, user_agent, metadata)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING created_at`

	err := r.db.QueryRow(ctx, query,
		contact.ID, contact.Name, contact.Email, contact.Phone, contact.Subject,
		contact.Message, contact.Status, contact.IPAddress, contact.UserAgent, contact.Metadata,
	).Scan(&contact.CreatedAt)
	if err != nil {
		return nil, fmt.Errorf("failed to create contact submission: %w", err)
	}

	return contact, nil
}

func (r *ContactRepository) GetByID(ctx context.Context, id uuid.UUID) (*models.ContactSubmission, error) {
	query := `
		SELECT id, name, email, phone, subject, message, status, ip_address, user_agent, metadata, read_at, created_at
		FROM contact_submissions
		WHERE id = $1`

	contact := &models.ContactSubmission{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&contact.ID, &contact.Name, &contact.Email, &contact.Phone, &contact.Subject,
		&contact.Message, &contact.Status, &contact.IPAddress, &contact.UserAgent,
		&contact.Metadata, &contact.ReadAt, &contact.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to get contact submission: %w", err)
	}

	return contact, nil
}

func (r *ContactRepository) List(ctx context.Context, filter models.ContactFilter) ([]models.ContactSubmission, int64, error) {
	filter.PaginationParams.Normalize()

	var conditions []string
	var args []interface{}
	argNum := 1

	if filter.Status != nil {
		conditions = append(conditions, fmt.Sprintf("status = $%d", argNum))
		args = append(args, *filter.Status)
		argNum++
	}

	if filter.Email != "" {
		conditions = append(conditions, fmt.Sprintf("email ILIKE $%d", argNum))
		args = append(args, "%"+filter.Email+"%")
		argNum++
	}

	whereClause := ""
	if len(conditions) > 0 {
		whereClause = "WHERE " + strings.Join(conditions, " AND ")
	}

	// Count total
	countQuery := fmt.Sprintf("SELECT COUNT(*) FROM contact_submissions %s", whereClause)
	var total int64
	if err := r.db.QueryRow(ctx, countQuery, args...).Scan(&total); err != nil {
		return nil, 0, fmt.Errorf("failed to count contact submissions: %w", err)
	}

	// Get data
	orderBy := "created_at DESC"
	if filter.SortBy != "" {
		orderBy = fmt.Sprintf("%s %s", filter.SortBy, filter.SortDir)
	}

	query := fmt.Sprintf(`
		SELECT id, name, email, phone, subject, message, status, ip_address, user_agent, metadata, read_at, created_at
		FROM contact_submissions
		%s
		ORDER BY %s
		LIMIT $%d OFFSET $%d`,
		whereClause, orderBy, argNum, argNum+1)

	args = append(args, filter.Limit(), filter.Offset())
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list contact submissions: %w", err)
	}
	defer rows.Close()

	var contacts []models.ContactSubmission
	for rows.Next() {
		var contact models.ContactSubmission
		if err := rows.Scan(
			&contact.ID, &contact.Name, &contact.Email, &contact.Phone, &contact.Subject,
			&contact.Message, &contact.Status, &contact.IPAddress, &contact.UserAgent,
			&contact.Metadata, &contact.ReadAt, &contact.CreatedAt,
		); err != nil {
			return nil, 0, fmt.Errorf("failed to scan contact submission: %w", err)
		}
		contacts = append(contacts, contact)
	}

	return contacts, total, nil
}

func (r *ContactRepository) Update(ctx context.Context, id uuid.UUID, req *models.UpdateContactRequest) (*models.ContactSubmission, error) {
	var setClauses []string
	var args []interface{}
	argNum := 1

	if req.Status != nil {
		setClauses = append(setClauses, fmt.Sprintf("status = $%d", argNum))
		args = append(args, *req.Status)
		argNum++
		// Auto-set read_at when status changes to "read"
		if *req.Status == models.ContactStatusRead {
			setClauses = append(setClauses, fmt.Sprintf("read_at = $%d", argNum))
			args = append(args, time.Now())
			argNum++
		}
	}

	if len(setClauses) == 0 {
		return r.GetByID(ctx, id)
	}

	args = append(args, id)
	query := fmt.Sprintf(`
		UPDATE contact_submissions
		SET %s
		WHERE id = $%d
		RETURNING id, name, email, phone, subject, message, status, ip_address, user_agent, metadata, read_at, created_at`,
		strings.Join(setClauses, ", "), argNum)

	contact := &models.ContactSubmission{}
	err := r.db.QueryRow(ctx, query, args...).Scan(
		&contact.ID, &contact.Name, &contact.Email, &contact.Phone, &contact.Subject,
		&contact.Message, &contact.Status, &contact.IPAddress, &contact.UserAgent,
		&contact.Metadata, &contact.ReadAt, &contact.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		return nil, fmt.Errorf("failed to update contact submission: %w", err)
	}

	return contact, nil
}

func (r *ContactRepository) Delete(ctx context.Context, id uuid.UUID) error {
	result, err := r.db.Exec(ctx, `DELETE FROM contact_submissions WHERE id = $1`, id)
	if err != nil {
		return fmt.Errorf("failed to delete contact submission: %w", err)
	}
	if result.RowsAffected() == 0 {
		return ErrNotFound
	}
	return nil
}

func (r *ContactRepository) GetUnreadCount(ctx context.Context) (int64, error) {
	var count int64
	err := r.db.QueryRow(ctx, `SELECT COUNT(*) FROM contact_submissions WHERE status = $1`, models.ContactStatusNew).Scan(&count)
	if err != nil {
		return 0, fmt.Errorf("failed to get unread count: %w", err)
	}
	return count, nil
}
