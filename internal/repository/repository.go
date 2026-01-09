package repository

import (
	"context"
	"errors"
)

// Common errors
var (
	ErrNotFound     = errors.New("record not found")
	ErrDuplicate    = errors.New("duplicate record")
	ErrInvalidInput = errors.New("invalid input")
	ErrForeignKey   = errors.New("foreign key constraint violation")
)

// DBTX represents a database transaction or connection pool interface
type DBTX interface {
	Exec(ctx context.Context, sql string, arguments ...interface{}) (interface{}, error)
	Query(ctx context.Context, sql string, args ...interface{}) (interface{}, error)
	QueryRow(ctx context.Context, sql string, args ...interface{}) interface{}
}
