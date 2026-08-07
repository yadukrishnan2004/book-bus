package domain

import (
	"context"
	"time"
)

// BookingStatus represents the state of a booking row.
type BookingStatus string

const (
	BookingStatusConfirmed BookingStatus = "confirmed"
	BookingStatusCancelled BookingStatus = "cancelled"
)

// Booking represents a single seat reservation.
// Multi-seat bookings share the same BookingReference.
type Booking struct {
	ID               string
	BookingReference string
	UserID           *string // nullable — populated when auth is added
	ScheduleID       string
	SeatNumber       int
	Status           BookingStatus
	TotalPrice       float64
	PassengerName    string
	PassengerPhone   string
	BookedAt         time.Time
	CreatedAt        time.Time
	UpdatedAt        time.Time
}

// BookingPreview is the result of the "check before confirm" step.
// No DB rows are written for a preview.
type BookingPreview struct {
	Available        bool
	UnavailableSeats []int
	ScheduleID       string
	SeatNumbers      []int
	PricePerSeat     float64
	TotalPrice       float64
	Schedule         ScheduleSummary
}

// ScheduleSummary is an abbreviated view of schedule + route + bus
// used inside BookingPreview and booking responses.
type ScheduleSummary struct {
	Origin        string
	Destination   string
	DepartureTime time.Time
	ArrivalTime   time.Time
	BusName       string
	BusType       string
}

// CreateBookingInput is the input for both preview and confirm actions.
type CreateBookingInput struct {
	ScheduleID     string
	SeatNumbers    []int
	PassengerName  string
	PassengerPhone string
}

// BookingRepository defines the data-access contract for bookings.
type BookingRepository interface {
	// GetBookedSeats returns seat numbers that are already taken on a schedule.
	GetBookedSeats(ctx context.Context, scheduleID string) ([]int, error)
	// CreateMany inserts one booking row per seat inside a single transaction.
	CreateMany(ctx context.Context, input CreateBookingInput, reference string, pricePerSeat float64) ([]*Booking, error)
	// GetByReference returns all seat bookings sharing a reference UUID.
	GetByReference(ctx context.Context, reference string) ([]*Booking, error)
	// CancelByReference cancels all bookings under a reference and restores seats.
	CancelByReference(ctx context.Context, reference string) error
}

// BookingService defines the business-logic contract for bookings.
type BookingService interface {
	// PreviewBooking validates seats and calculates price — no DB write.
	PreviewBooking(ctx context.Context, input CreateBookingInput) (*BookingPreview, error)
	// ConfirmBooking locks the seats and creates the booking.
	ConfirmBooking(ctx context.Context, input CreateBookingInput) ([]*Booking, string, error)
	// GetBooking retrieves all seat rows under a booking reference.
	GetBooking(ctx context.Context, reference string) ([]*Booking, error)
	// CancelBooking cancels all seats under a booking reference.
	CancelBooking(ctx context.Context, reference string) error
}
