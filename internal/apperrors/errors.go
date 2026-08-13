package apperrors

import (
	"errors"
	"fmt"
)

// Sentinel errors used across all layers.
// Handlers check for these with errors.Is() instead of importing pgx directly.
var (
	// ErrNotFound is returned when a requested resource does not exist.
	ErrNotFound = errors.New("not found")

	// ErrDuplicateKey is returned when a unique constraint is violated.
	ErrDuplicateKey = errors.New("duplicate key")

	// ErrNoSeatsAvailable is returned when the schedule has no capacity left.
	ErrNoSeatsAvailable = errors.New("seats not available")

	// ErrBookingNotCancellable is returned when a booking cannot be cancelled
	// (e.g. it is already cancelled or the trip has completed).
	ErrBookingNotCancellable = errors.New("booking cannot be cancelled")

	// ErrInvalidReference is returned when a foreign key constraint is violated,
	// meaning a referenced entity (e.g. bus_id, route_id) does not exist.
	ErrInvalidReference = errors.New("referenced resource does not exist")

	// Auth errors
	ErrInvalidCredentials = errors.New("invalid email or password")
	ErrUserAlreadyExists  = errors.New("user with this email already exists")
	ErrUnauthorized       = errors.New("unauthorized")
	ErrForbidden          = errors.New("forbidden")
	ErrInvalidToken       = errors.New("invalid or expired token")
)

// SeatsConflictError is returned when one or more specific requested seats are
// already taken. It carries the exact conflicting seat numbers so callers can
// surface them to the client.
type SeatsConflictError struct {
	ConflictingSeats []int
}

func (e *SeatsConflictError) Error() string {
	return fmt.Sprintf("seats already taken: %v", e.ConflictingSeats)
}

// NewSeatsConflictError constructs a SeatsConflictError.
func NewSeatsConflictError(seats []int) *SeatsConflictError {
	return &SeatsConflictError{ConflictingSeats: seats}
}
