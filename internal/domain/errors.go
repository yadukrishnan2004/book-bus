package domain

import "errors"

// Domain-level validation errors (not DB errors — those live in apperrors).
var (
	// ErrArrivalBeforeDeparture is returned when a schedule's arrival time is
	// set to before its departure time.
	ErrArrivalBeforeDeparture = errors.New("arrival time must be after departure time")

	// ErrInvalidStatusTransition is returned when an invalid state transition is attempted
	// (e.g. attempting to change status of an already completed or cancelled trip).
	ErrInvalidStatusTransition = errors.New("invalid status transition: trip is already completed or cancelled")
)
