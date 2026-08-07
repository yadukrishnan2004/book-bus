package domain

import (
	"context"
	"time"
)

// Route represents a travel path between two cities.
type Route struct {
	ID          string
	Origin      string
	Destination string
	DistanceKM  float64
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// CreateRouteInput carries the fields needed to create a route.
type CreateRouteInput struct {
	Origin      string
	Destination string
	DistanceKM  float64
}

// RouteRepository defines the data-access contract for routes.
type RouteRepository interface {
	Create(ctx context.Context, input CreateRouteInput) (*Route, error)
	List(ctx context.Context) ([]*Route, error)
	GetByID(ctx context.Context, id string) (*Route, error)
}

// RouteService defines the business-logic contract for routes.
type RouteService interface {
	CreateRoute(ctx context.Context, input CreateRouteInput) (*Route, error)
	ListRoutes(ctx context.Context) ([]*Route, error)
}
