package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"book-bus/internal/models"
)

// BusRepository handles all database operations for buses.
type BusRepository struct {
	db *pgxpool.Pool
}

// NewBusRepository creates a new BusRepository.
func NewBusRepository(db *pgxpool.Pool) *BusRepository {
	return &BusRepository{db: db}
}

// Create inserts a new bus record and returns the created bus.
func (r *BusRepository) Create(ctx context.Context, req models.CreateBusRequest) (*models.Bus, error) {
	query := `
		INSERT INTO buses (name, number_plate, total_seats, bus_type, description, is_active)
		VALUES ($1, $2, $3, $4, $5, true)
		RETURNING id, name, number_plate, total_seats, bus_type, description, is_active, created_at, updated_at
	`

	bus := &models.Bus{}
	err := r.db.QueryRow(ctx, query,
		req.Name,
		req.NumberPlate,
		req.TotalSeats,
		req.BusType,
		req.Description,
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
		return nil, fmt.Errorf("repository: create bus: %w", err)
	}

	return bus, nil
}

// GetByID fetches a single bus by its UUID.
func (r *BusRepository) GetByID(ctx context.Context, id string) (*models.Bus, error) {
	query := `
		SELECT id, name, number_plate, total_seats, bus_type, description, is_active, created_at, updated_at
		FROM buses
		WHERE id = $1
	`

	bus := &models.Bus{}
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
		return nil, fmt.Errorf("repository: get bus by id: %w", err)
	}

	return bus, nil
}

// List returns all buses with optional pagination.
func (r *BusRepository) List(ctx context.Context, limit, offset int) ([]*models.Bus, error) {
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

	var buses []*models.Bus
	for rows.Next() {
		bus := &models.Bus{}
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
