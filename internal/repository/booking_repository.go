package repository

import (
	"context"
	"errors"
	"fmt"
	"strings"

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

// CreateMany atomically inserts seat rows in bulk and decrements available_seats.
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

	// Bulk insert all seats in a single SQL query round-trip
	valueStrings := make([]string, 0, len(input.SeatNumbers))
	valueArgs := make([]interface{}, 0, len(input.SeatNumbers)*7)
	for i, seatNum := range input.SeatNumbers {
		base := i * 7
		valueStrings = append(valueStrings, fmt.Sprintf("($%d, $%d, $%d, 'confirmed', $%d, $%d, $%d, $%d)", base+1, base+2, base+3, base+4, base+5, base+6, base+7))
		valueArgs = append(valueArgs, reference, input.ScheduleID, seatNum, pricePerSeat, input.PassengerName, input.PassengerPhone, input.UserID)
	}

	query := fmt.Sprintf(`
		INSERT INTO bookings
			(booking_reference, schedule_id, seat_number, status, total_price, passenger_name, passenger_phone, user_id)
		VALUES %s
		RETURNING id, booking_reference, user_id, schedule_id, seat_number, status,
		          total_price, passenger_name, passenger_phone,
		          booked_at, created_at, updated_at
	`, strings.Join(valueStrings, ", "))

	rows, err := tx.Query(ctx, query, valueArgs...)
	if err != nil {
		return nil, fmt.Errorf("repository: bulk insert bookings: %w", mapDBError(err, "bulk insert bookings"))
	}
	defer rows.Close()

	var bookings []*domain.Booking
	for rows.Next() {
		b := &domain.Booking{}
		if err := rows.Scan(
			&b.ID, &b.BookingReference, &b.UserID, &b.ScheduleID, &b.SeatNumber, &b.Status,
			&b.TotalPrice, &b.PassengerName, &b.PassengerPhone,
			&b.BookedAt, &b.CreatedAt, &b.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("repository: scan bulk insert booking: %w", err)
		}
		bookings = append(bookings, b)
	}
	if err := rows.Err(); err != nil {
		return nil, err
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
		SELECT id, booking_reference, user_id, schedule_id, seat_number, status,
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
			&b.ID, &b.BookingReference, &b.UserID, &b.ScheduleID, &b.SeatNumber, &b.Status,
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

// GetByUserID returns all bookings belonging to a specific registered user.
func (r *bookingRepository) GetByUserID(ctx context.Context, userID string) ([]*domain.Booking, error) {
	rows, err := r.db.Query(ctx, `
		SELECT id, booking_reference, user_id, schedule_id, seat_number, status,
		       total_price, passenger_name, passenger_phone,
		       booked_at, created_at, updated_at
		FROM bookings
		WHERE user_id = $1
		ORDER BY booked_at DESC, seat_number ASC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("repository: get by user id: %w", err)
	}
	defer rows.Close()

	var bookings []*domain.Booking
	for rows.Next() {
		b := &domain.Booking{}
		if err := rows.Scan(
			&b.ID, &b.BookingReference, &b.UserID, &b.ScheduleID, &b.SeatNumber, &b.Status,
			&b.TotalPrice, &b.PassengerName, &b.PassengerPhone,
			&b.BookedAt, &b.CreatedAt, &b.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("repository: scan user booking: %w", err)
		}
		bookings = append(bookings, b)
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
