package service

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/google/uuid"

	"book-bus/internal/apperrors"
	"book-bus/internal/domain"
)

type bookingService struct {
	bookingRepo  domain.BookingRepository
	scheduleRepo domain.ScheduleRepository
}

// NewBookingService creates a bookingService implementing domain.BookingService.
func NewBookingService(bookingRepo domain.BookingRepository, scheduleRepo domain.ScheduleRepository) domain.BookingService {
	return &bookingService{
		bookingRepo:  bookingRepo,
		scheduleRepo: scheduleRepo,
	}
}

// PreviewBooking validates seat availability and returns pricing — no DB writes.
func (s *bookingService) PreviewBooking(ctx context.Context, input domain.CreateBookingInput) (*domain.BookingPreview, error) {
	// Fetch schedule with bus + route info
	schedule, err := s.scheduleRepo.GetByID(ctx, input.ScheduleID)
	if err != nil {
		return nil, err
	}
	if schedule.Status != domain.ScheduleStatusScheduled {
		return nil, apperrors.ErrNoSeatsAvailable
	}

	// Get already booked seats
	bookedSeats, err := s.bookingRepo.GetBookedSeats(ctx, input.ScheduleID)
	if err != nil {
		return nil, err
	}
	booked := make(map[int]bool, len(bookedSeats))
	for _, n := range bookedSeats {
		booked[n] = true
	}

	// Determine which requested seats are unavailable
	var unavailable []int
	for _, n := range input.SeatNumbers {
		if booked[n] {
			unavailable = append(unavailable, n)
		}
	}
	if unavailable == nil {
		unavailable = []int{}
	}

	available := len(unavailable) == 0
	totalPrice := schedule.Price * float64(len(input.SeatNumbers))

	summary := domain.ScheduleSummary{}
	if schedule.Route != nil {
		summary.Origin = schedule.Route.Origin
		summary.Destination = schedule.Route.Destination
	}
	if schedule.Bus != nil {
		summary.BusName = schedule.Bus.Name
		summary.BusType = string(schedule.Bus.BusType)
	}
	summary.DepartureTime = schedule.DepartureTime
	summary.ArrivalTime = schedule.ArrivalTime

	return &domain.BookingPreview{
		Available:        available,
		UnavailableSeats: unavailable,
		ScheduleID:       input.ScheduleID,
		SeatNumbers:      input.SeatNumbers,
		PricePerSeat:     schedule.Price,
		TotalPrice:       totalPrice,
		Schedule:         summary,
	}, nil
}


// ConfirmBooking locks the seats and creates the booking inside a transaction.
func (s *bookingService) ConfirmBooking(ctx context.Context, input domain.CreateBookingInput) ([]*domain.Booking, string, error) {
	// Fetch schedule to get the price per seat
	schedule, err := s.scheduleRepo.GetByID(ctx, input.ScheduleID)
	if err != nil {
		return nil, "", err
	}
	if schedule.Status != domain.ScheduleStatusScheduled {
		return nil, "", apperrors.ErrNoSeatsAvailable
	}

	// Validate seat numbers are within the bus range
	for _, n := range input.SeatNumbers {
		if n < 1 || n > schedule.TotalSeats {
			return nil, "", fmt.Errorf("seat number %d is out of range (bus has %d seats)", n, schedule.TotalSeats)
		}
	}

	// Fast pre-check: identify any already-taken seats before entering the
	// transaction lock. This gives a richer error and avoids holding the lock
	// when we already know the booking will fail. The authoritative check is
	// still performed inside CreateMany within a FOR UPDATE transaction.
	bookedSeats, err := s.bookingRepo.GetBookedSeats(ctx, input.ScheduleID)
	if err != nil {
		return nil, "", err
	}
	booked := make(map[int]bool, len(bookedSeats))
	for _, n := range bookedSeats {
		booked[n] = true
	}
	var conflicting []int
	for _, n := range input.SeatNumbers {
		if booked[n] {
			conflicting = append(conflicting, n)
		}
	}
	if len(conflicting) > 0 {
		return nil, "", apperrors.NewSeatsConflictError(conflicting)
	}

	reference := uuid.New().String()

	bookings, err := s.bookingRepo.CreateMany(ctx, input, reference, schedule.Price)
	if err != nil {
		slog.Error("service: confirm booking failed",
			"schedule_id", input.ScheduleID,
			"seats", input.SeatNumbers,
			"error", err,
		)
		return nil, "", err
	}

	slog.Info("booking confirmed",
		"reference", reference,
		"schedule_id", input.ScheduleID,
		"seats", input.SeatNumbers,
		"passenger", input.PassengerName,
	)
	return bookings, reference, nil
}


// GetBooking retrieves all seat rows under a booking reference.
func (s *bookingService) GetBooking(ctx context.Context, reference string) ([]*domain.Booking, error) {
	bookings, err := s.bookingRepo.GetByReference(ctx, reference)
	if err != nil {
		slog.Error("service: get booking failed", "reference", reference, "error", err)
		return nil, err
	}
	return bookings, nil
}

// GetUserBookings retrieves all bookings associated with an authenticated user ID.
func (s *bookingService) GetUserBookings(ctx context.Context, userID string) ([]*domain.Booking, error) {
	bookings, err := s.bookingRepo.GetByUserID(ctx, userID)
	if err != nil {
		slog.Error("service: get user bookings failed", "user_id", userID, "error", err)
		return nil, err
	}
	return bookings, nil
}

// CancelBooking cancels all seats under a reference and restores available seats.
func (s *bookingService) CancelBooking(ctx context.Context, reference string) error {
	if err := s.bookingRepo.CancelByReference(ctx, reference); err != nil {
		slog.Error("service: cancel booking failed", "reference", reference, "error", err)
		return err
	}
	slog.Info("booking cancelled", "reference", reference)
	return nil
}
