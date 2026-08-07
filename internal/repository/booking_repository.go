package repository

import (
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	"book-bus/internal/apperrors"
	"book-bus/internal/domain"
)

type bookingRepository struct {
	db *pgxpool.Pool
}

// NewBookingRepository creates a new bookingRepository implementing domain.BookingRepository.
func NewBookingRepository(db *pgxpool.Pool) domain.BookingRepository {
	return &bookingRepository{db: db}
}

// GetBookedSeats returns seat numbers already taken on a schedule.
func (r *bookingRepository) GetBookedSeats(ctx context.Context, scheduleID string) ([]int, error) {
	rows, err := r.db.Query(ctx, `
		SELECT seat_number FROM bookings
		WHERE schedule_id = $1 AND status != 'cancelled'
	`, scheduleID)
	if err != nil {
		return nil, fmt.Errorf("repository: get booked seats: %w", err)
	}
	defer rows.Close()

	var seats []int
	for rows.Next() {
		var n int
		if err := rows.Scan(&n); err != nil {
			return nil, fmt.Errorf("repository: scan booked seat: %w", err)
		}
		seats = append(seats, n)
	}
	return seats, rows.Err()
}

// CreateMany atomically inserts one booking row per seat and decrements available_seats.
// It uses SELECT FOR UPDATE on the schedule to prevent race conditions.
func (r *bookingRepository) CreateMany(ctx context.Context, input domain.CreateBookingInput, reference string, pricePerSeat float64) ([]*domain.Booking, error) {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("repository: begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Lock the schedule row to prevent concurrent overbooking
	var availableSeats int
	err = tx.QueryRow(ctx,
		`SELECT available_seats FROM schedules WHERE id = $1 FOR UPDATE`,
		input.ScheduleID,
	).Scan(&availableSeats)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apperrors.ErrNotFound
		}
		return nil, fmt.Errorf("repository: lock schedule: %w", err)
	}

	if availableSeats < len(input.SeatNumbers) {
		return nil, apperrors.ErrNoSeatsAvailable
	}

	// Verify none of the requested seats are already taken
	var takenCount int
	err = tx.QueryRow(ctx, `
		SELECT COUNT(*) FROM bookings
		WHERE schedule_id = $1
		  AND seat_number = ANY($2::int[])
		  AND status != 'cancelled'
	`, input.ScheduleID, input.SeatNumbers).Scan(&takenCount)
	if err != nil {
		return nil, fmt.Errorf("repository: check seats: %w", err)
	}
	if takenCount > 0 {
		return nil, apperrors.ErrNoSeatsAvailable
	}

	// Insert one row per seat
	var bookings []*domain.Booking
	for _, seatNum := range input.SeatNumbers {
		b := &domain.Booking{}
		err = tx.QueryRow(ctx, `
			INSERT INTO bookings
				(booking_reference, schedule_id, seat_number, status, total_price, passenger_name, passenger_phone)
			VALUES ($1, $2, $3, 'confirmed', $4, $5, $6)
			RETURNING id, booking_reference, schedule_id, seat_number, status,
			          total_price, passenger_name, passenger_phone,
			          booked_at, created_at, updated_at
		`, reference, input.ScheduleID, seatNum, pricePerSeat,
			input.PassengerName, input.PassengerPhone,
		).Scan(
			&b.ID, &b.BookingReference, &b.ScheduleID, &b.SeatNumber, &b.Status,
			&b.TotalPrice, &b.PassengerName, &b.PassengerPhone,
			&b.BookedAt, &b.CreatedAt, &b.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("repository: insert booking seat %d: %w", seatNum, mapDBError(err, "insert booking"))
		}
		bookings = append(bookings, b)
	}

	// Decrement available seats
	if _, err = tx.Exec(ctx,
		`UPDATE schedules SET available_seats = available_seats - $1 WHERE id = $2`,
		len(input.SeatNumbers), input.ScheduleID,
	); err != nil {
		return nil, fmt.Errorf("repository: decrement seats: %w", err)
	}

	if err = tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("repository: commit: %w", err)
	}
	return bookings, nil
}

// GetByReference returns all booking rows sharing a booking_reference UUID.
func (r *bookingRepository) GetByReference(ctx context.Context, reference string) ([]*domain.Booking, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, booking_reference, schedule_id, seat_number, status,
		       total_price, passenger_name, passenger_phone,
		       booked_at, created_at, updated_at
		FROM bookings
		WHERE booking_reference = $1
		ORDER BY seat_number ASC
	`, reference)
	if err != nil {
		return nil, fmt.Errorf("repository: get by reference: %w", err)
	}
	defer rows.Close()

	var bookings []*domain.Booking
	for rows.Next() {
		b := &domain.Booking{}
		if err := rows.Scan(
			&b.ID, &b.BookingReference, &b.ScheduleID, &b.SeatNumber, &b.Status,
			&b.TotalPrice, &b.PassengerName, &b.PassengerPhone,
			&b.BookedAt, &b.CreatedAt, &b.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("repository: scan booking: %w", err)
		}
		bookings = append(bookings, b)
	}
	if len(bookings) == 0 {
		return nil, apperrors.ErrNotFound
	}
	return bookings, rows.Err()
}

// CancelByReference cancels all bookings under a reference and restores available seats.
func (r *bookingRepository) CancelByReference(ctx context.Context, reference string) error {
	tx, err := r.db.Begin(ctx)
	if err != nil {
		return fmt.Errorf("repository: begin tx: %w", err)
	}
	defer tx.Rollback(ctx)

	// Lock and read the bookings
	rows, err := tx.Query(ctx, `
		SELECT id, schedule_id, status FROM bookings
		WHERE booking_reference = $1
		FOR UPDATE
	`, reference)
	if err != nil {
		return fmt.Errorf("repository: lock bookings: %w", err)
	}

	type row struct{ id, scheduleID string; status domain.BookingStatus }
	var found []row
	for rows.Next() {
		var b row
		if err := rows.Scan(&b.id, &b.scheduleID, &b.status); err != nil {
			rows.Close()
			return fmt.Errorf("repository: scan booking for cancel: %w", err)
		}
		found = append(found, b)
	}
	rows.Close()

	if len(found) == 0 {
		return apperrors.ErrNotFound
	}
	for _, b := range found {
		if b.status == domain.BookingStatusCancelled {
			return apperrors.ErrBookingNotCancellable
		}
	}

	scheduleID := found[0].scheduleID
	count := len(found)

	if _, err = tx.Exec(ctx,
		`UPDATE bookings SET status = 'cancelled', updated_at = NOW() WHERE booking_reference = $1`,
		reference,
	); err != nil {
		return fmt.Errorf("repository: cancel bookings: %w", err)
	}

	if _, err = tx.Exec(ctx,
		`UPDATE schedules SET available_seats = available_seats + $1 WHERE id = $2`,
		count, scheduleID,
	); err != nil {
		return fmt.Errorf("repository: restore seats: %w", err)
	}

	return tx.Commit(ctx)
}
