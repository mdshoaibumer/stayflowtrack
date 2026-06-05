package errors

import (
	"errors"

	"github.com/jackc/pgx/v5"
)

// IsNotFound checks if an error is a pgx "no rows" error.
func IsNotFound(err error) bool {
	return errors.Is(err, pgx.ErrNoRows)
}

// WrapRepository wraps a repository error with appropriate AppError.
// Use this to convert raw database errors into user-facing errors.
func WrapRepository(err error, resource, id string) error {
	if err == nil {
		return nil
	}
	if IsNotFound(err) {
		return NotFound(resource, id)
	}
	return Internal(err)
}
