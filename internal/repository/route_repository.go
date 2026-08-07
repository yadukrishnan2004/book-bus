package repository

import (
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"

	"book-bus/internal/domain"
)

type routeRepository struct {
	db *pgxpool.Pool
}

// NewRouteRepository creates a new routeRepository implementing domain.RouteRepository.
func NewRouteRepository(db *pgxpool.Pool) domain.RouteRepository {
	return &routeRepository{db: db}
}

func (r *routeRepository) Create(ctx context.Context, input domain.CreateRouteInput) (*domain.Route, error) {
	query := `
		INSERT INTO routes (origin, destination, distance_km)
		VALUES ($1, $2, $3)
		RETURNING id, origin, destination, distance_km, created_at, updated_at
	`
	route := &domain.Route{}
	err := r.db.QueryRow(ctx, query, input.Origin, input.Destination, input.DistanceKM).Scan(
		&route.ID, &route.Origin, &route.Destination, &route.DistanceKM, &route.CreatedAt, &route.UpdatedAt,
	)
	if err != nil {
		return nil, fmt.Errorf("repository: create route: %w", mapDBError(err, "create route"))
	}
	return route, nil
}

func (r *routeRepository) List(ctx context.Context) ([]*domain.Route, error) {
	query := `
		SELECT id, origin, destination, distance_km, created_at, updated_at
		FROM routes ORDER BY origin ASC
	`
	rows, err := r.db.Query(ctx, query)
	if err != nil {
		return nil, fmt.Errorf("repository: list routes: %w", err)
	}
	defer rows.Close()

	var routes []*domain.Route
	for rows.Next() {
		route := &domain.Route{}
		if err := rows.Scan(
			&route.ID, &route.Origin, &route.Destination, &route.DistanceKM, &route.CreatedAt, &route.UpdatedAt,
		); err != nil {
			return nil, fmt.Errorf("repository: scan route: %w", err)
		}
		routes = append(routes, route)
	}
	return routes, rows.Err()
}

func (r *routeRepository) GetByID(ctx context.Context, id string) (*domain.Route, error) {
	query := `
		SELECT id, origin, destination, distance_km, created_at, updated_at
		FROM routes WHERE id = $1
	`
	route := &domain.Route{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&route.ID, &route.Origin, &route.Destination, &route.DistanceKM, &route.CreatedAt, &route.UpdatedAt,
	)
	if err != nil {
		return nil, mapDBError(err, "get route by id")
	}
	return route, nil
}
