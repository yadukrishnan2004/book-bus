package repository

import (
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"

	"book-bus/internal/apperrors"
)

// mapDBError translates raw pgx errors into domain sentinel errors.
// Call this in every repository method before returning an error.
func mapDBError(err error, op string) error {
	if err == nil {
		return nil
	}
	if errors.Is(err, pgx.ErrNoRows) {
		return apperrors.ErrNotFound
	}
	var pgErr *pgconn.PgError
	if errors.As(err, &pgErr) {
		switch pgErr.Code {
		case "23505": // unique_violation
			return apperrors.ErrDuplicateKey
		case "23503": // foreign_key_violation
			return apperrors.ErrInvalidReference
		}
	}
	return err
}

