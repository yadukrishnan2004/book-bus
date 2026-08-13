package domain

import "errors"

// Domain-level validation errors (not DB errors — those live in apperrors).
var (
	// ErrArrivalBeforeDeparture is returned when a schedule's arrival time is
	// set to before its departure time.
	ErrArrivalBeforeDeparture = errors.New("arrival time must be after departure time")

	// ErrZeroDuration is returned when arrival and departure times are identical.
	ErrZeroDuration = errors.New("arrival time must be after departure time (zero-duration trips are not allowed)")

	// ErrDepartureInPast is returned when a schedule's departure time is in the past.
	ErrDepartureInPast = errors.New("departure time cannot be in the past")

	// ErrInvalidStatusTransition is returned when an invalid state transition is attempted
	// (e.g. attempting to change status of an already completed or cancelled trip).
	ErrInvalidStatusTransition = errors.New("invalid status transition: trip is already completed or cancelled")

	// ErrSameOriginDestination is returned when a route's origin and destination are the same.
	ErrSameOriginDestination = errors.New("origin and destination cannot be the same")

	// ErrInvalidDateFormat is returned when a date query parameter is not in YYYY-MM-DD format.
	ErrInvalidDateFormat = errors.New("invalid date format — use YYYY-MM-DD")
)
