package models

import (
	"time"

	"book-bus/internal/domain"
)

// CreateRouteRequest is the JSON payload for creating a route.
type CreateRouteRequest struct {
	Origin      string  `json:"origin"       binding:"required,min=2,max=255"`
	Destination string  `json:"destination"  binding:"required,min=2,max=255"`
	DistanceKM  float64 `json:"distance_km"  binding:"min=0"`
}

// ToDomainInput converts the request to a domain input.
func (r CreateRouteRequest) ToDomainInput() domain.CreateRouteInput {
	return domain.CreateRouteInput{
		Origin:      r.Origin,
		Destination: r.Destination,
		DistanceKM:  r.DistanceKM,
	}
}

// RouteResponse is the API response shape for a route.
type RouteResponse struct {
	ID          string    `json:"id"`
	Origin      string    `json:"origin"`
	Destination string    `json:"destination"`
	DistanceKM  float64   `json:"distance_km"`
	CreatedAt   time.Time `json:"created_at"`
}

// NewRouteResponse maps a domain.Route to the API response.
func NewRouteResponse(r *domain.Route) *RouteResponse {
	return &RouteResponse{
		ID:          r.ID,
		Origin:      r.Origin,
		Destination: r.Destination,
		DistanceKM:  r.DistanceKM,
		CreatedAt:   r.CreatedAt,
	}
}
