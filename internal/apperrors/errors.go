package apperrors

import "errors"

// Sentinel errors used across all layers.
// Handlers check for these with errors.Is() instead of importing pgx directly.
var (
	// ErrNotFound is returned when a requested resource does not exist.
	ErrNotFound = errors.New("not found")

	// ErrDuplicateKey is returned when a unique constraint is violated.
	ErrDuplicateKey = errors.New("duplicate key")

	// ErrNoSeatsAvailable is returned when requested seats are already booked.
	ErrNoSeatsAvailable = errors.New("seats not available")

	// ErrBookingNotCancellable is returned when a booking cannot be cancelled
	// (e.g. it is already cancelled or the trip has completed).
	ErrBookingNotCancellable = errors.New("booking cannot be cancelled")
)
