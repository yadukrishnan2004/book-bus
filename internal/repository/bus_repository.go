package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"book-bus/internal/domain"
)


// busRepository is the PostgreSQL implementation of domain.BusRepository.
type busRepository struct {
	db *pgxpool.Pool
}

// NewBusRepository creates a new busRepository.
// It returns the domain interface so callers stay decoupled from this package.
func NewBusRepository(db *pgxpool.Pool) domain.BusRepository {
	return &busRepository{db: db}
}

// Create inserts a new bus record and returns the created entity.
func (r *busRepository) Create(ctx context.Context, input domain.CreateBusInput) (*domain.Bus, error) {
	query := `
		INSERT INTO buses (name, number_plate, total_seats, bus_type, description, is_active)
		VALUES ($1, $2, $3, $4, $5, true)
		RETURNING id, name, number_plate, total_seats, bus_type, description, is_active, created_at, updated_at
	`

	bus := &domain.Bus{}
	err := r.db.QueryRow(ctx, query,
		input.Name,
		input.NumberPlate,
		input.TotalSeats,
		input.BusType,
		input.Description,
	).Scan(
		&bus.ID,
		&bus.Name,
		&bus.NumberPlate,
		&bus.TotalSeats,
		&bus.BusType,
		&bus.Description,
		&bus.IsActive,
		&bus.CreatedAt,
		&bus.UpdatedAt,
	)
	if err != nil {
		return nil, mapDBError(err, "create bus")
	}

	return bus, nil
}

// GetByID fetches a single bus by its UUID.
func (r *busRepository) GetByID(ctx context.Context, id string) (*domain.Bus, error) {
	query := `
		SELECT id, name, number_plate, total_seats, bus_type, description, is_active, created_at, updated_at
		FROM buses
		WHERE id = $1
	`

	bus := &domain.Bus{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&bus.ID,
		&bus.Name,
		&bus.NumberPlate,
		&bus.TotalSeats,
		&bus.BusType,
		&bus.Description,
		&bus.IsActive,
		&bus.CreatedAt,
		&bus.UpdatedAt,
	)
	if err != nil {
		return nil, mapDBError(err, "get bus by id")
	}

	return bus, nil
}

// List returns buses ordered by creation date with pagination.
func (r *busRepository) List(ctx context.Context, limit, offset int) ([]*domain.Bus, error) {
	query := `
		SELECT id, name, number_plate, total_seats, bus_type, description, is_active, created_at, updated_at
		FROM buses
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2
	`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, fmt.Errorf("repository: list buses: %w", err)
	}
	defer rows.Close()

	var buses []*domain.Bus
	for rows.Next() {
		bus := &domain.Bus{}
		if err := rows.Scan(
			&bus.ID,
			&bus.Name,
			&bus.NumberPlate,
			&bus.TotalSeats,
			&bus.BusType,
			&bus.Description,
			&bus.IsActive,
			&bus.CreatedAt,
			&bus.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("repository: scan bus row: %w", err)
		}
		buses = append(buses, bus)
	}

	return buses, rows.Err()
}


