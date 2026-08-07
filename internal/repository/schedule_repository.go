package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"book-bus/internal/domain"
)

type scheduleRepository struct {
	db *pgxpool.Pool
}

// NewScheduleRepository creates a new scheduleRepository implementing domain.ScheduleRepository.
func NewScheduleRepository(db *pgxpool.Pool) domain.ScheduleRepository {
	return &scheduleRepository{db: db}
}

// Create inserts a schedule. available_seats is auto-set from the linked bus's total_seats.
func (r *scheduleRepository) Create(ctx context.Context, input domain.CreateScheduleInput) (*domain.Schedule, error) {
	query := `
		INSERT INTO schedules (bus_id, route_id, departure_time, arrival_time, price, available_seats, status)
		SELECT $1, $2, $3, $4, $5, b.total_seats, 'scheduled'
		FROM buses b WHERE b.id = $1
		RETURNING id, bus_id, route_id, departure_time, arrival_time, price, available_seats, status, created_at, updated_at
	`
	s := &domain.Schedule{}
	err := r.db.QueryRow(ctx, query,
		input.BusID, input.RouteID, input.DepartureTime, input.ArrivalTime, input.Price,
	).Scan(
		&s.ID, &s.BusID, &s.RouteID, &s.DepartureTime, &s.ArrivalTime,
		&s.Price, &s.AvailableSeats, &s.Status, &s.CreatedAt, &s.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("repository: create schedule: %w", mapDBError(err, "create schedule"))
	}
	return s, nil
}

// GetByID returns the full schedule with embedded bus and route.
func (r *scheduleRepository) GetByID(ctx context.Context, id string) (*domain.Schedule, error) {
	query := `
		SELECT
			s.id, s.bus_id, s.route_id, s.departure_time, s.arrival_time,
			s.price, s.available_seats, s.status, s.created_at, s.updated_at,
			b.id, b.name, b.number_plate, b.total_seats, b.bus_type,
			r.id, r.origin, r.destination, COALESCE(r.distance_km, 0)
		FROM schedules s
		JOIN buses  b ON b.id = s.bus_id
		JOIN routes r ON r.id = s.route_id
		WHERE s.id = $1
	`
	s := &domain.Schedule{Bus: &domain.Bus{}, Route: &domain.Route{}}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&s.ID, &s.BusID, &s.RouteID, &s.DepartureTime, &s.ArrivalTime,
		&s.Price, &s.AvailableSeats, &s.Status, &s.CreatedAt, &s.UpdatedAt,
		&s.Bus.ID, &s.Bus.Name, &s.Bus.NumberPlate, &s.TotalSeats, &s.Bus.BusType,
		&s.Route.ID, &s.Route.Origin, &s.Route.Destination, &s.Route.DistanceKM,
	)
	if err != nil {
		return nil, mapDBError(err, "get schedule by id")
	}
	return s, nil
}

// List returns schedules with optional filtering by origin, destination, and date.
func (r *scheduleRepository) List(ctx context.Context, filter domain.ScheduleFilter, limit, offset int) ([]*domain.Schedule, error) {
	query := `
		SELECT
			s.id, s.bus_id, s.route_id, s.departure_time, s.arrival_time,
			s.price, s.available_seats, s.status, s.created_at, s.updated_at,
			b.id, b.name, b.number_plate, b.total_seats, b.bus_type,
			r.id, r.origin, r.destination, COALESCE(r.distance_km, 0)
		FROM schedules s
		JOIN buses  b ON b.id = s.bus_id
		JOIN routes r ON r.id = s.route_id
		WHERE s.status = 'scheduled'
		  AND ($1 = '' OR LOWER(r.origin)      LIKE '%' || LOWER($1) || '%')
		  AND ($2 = '' OR LOWER(r.destination) LIKE '%' || LOWER($2) || '%')
		  AND ($3 = '' OR DATE(s.departure_time AT TIME ZONE 'UTC') = $3::DATE)
		ORDER BY s.departure_time ASC
		LIMIT $4 OFFSET $5
	`
	rows, err := r.db.Query(ctx, query, filter.Origin, filter.Destination, filter.Date, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("repository: list schedules: %w", err)
	}
	defer rows.Close()

	var schedules []*domain.Schedule
	for rows.Next() {
		s := &domain.Schedule{Bus: &domain.Bus{}, Route: &domain.Route{}}
		if err := rows.Scan(
			&s.ID, &s.BusID, &s.RouteID, &s.DepartureTime, &s.ArrivalTime,
			&s.Price, &s.AvailableSeats, &s.Status, &s.CreatedAt, &s.UpdatedAt,
			&s.Bus.ID, &s.Bus.Name, &s.Bus.NumberPlate, &s.TotalSeats, &s.Bus.BusType,
			&s.Route.ID, &s.Route.Origin, &s.Route.Destination, &s.Route.DistanceKM,
		); err != nil {
			return nil, fmt.Errorf("repository: scan schedule: %w", err)
		}
		schedules = append(schedules, s)
	}
	return schedules, rows.Err()
}

// GetSeatMap builds a seat availability map for a schedule.
func (r *scheduleRepository) GetSeatMap(ctx context.Context, scheduleID string) ([]domain.SeatInfo, error) {
	// Step 1: get total seats from the linked bus
	var totalSeats int
	err := r.db.QueryRow(ctx, `
		SELECT b.total_seats FROM buses b
		JOIN schedules s ON s.bus_id = b.id
		WHERE s.id = $1
	`, scheduleID).Scan(&totalSeats)
	if err != nil {
		return nil, mapDBError(err, "get seat map total")
	}

	// Step 2: get booked seat numbers (non-cancelled)
	rows, err := r.db.Query(ctx, `
		SELECT seat_number FROM bookings
		WHERE schedule_id = $1 AND status != 'cancelled'
	`, scheduleID)
	if err != nil {
		return nil, fmt.Errorf("repository: get booked seats: %w", err)
	}
	defer rows.Close()

	booked := make(map[int]bool)
	for rows.Next() {
		var n int
		if err := rows.Scan(&n); err != nil {
			return nil, fmt.Errorf("repository: scan booked seat: %w", err)
		}
		booked[n] = true
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// Step 3: build seat map (seats numbered 1 → totalSeats)
	seats := make([]domain.SeatInfo, totalSeats)
	for i := 1; i <= totalSeats; i++ {
		seats[i-1] = domain.SeatInfo{
			Number:      i,
			IsAvailable: !booked[i],
		}
	}
	return seats, nil
}
