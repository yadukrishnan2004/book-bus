package models

import (
	"time"

	"book-bus/internal/domain"
)

// CreateBusRequest is the JSON payload for registering a new bus.
// It lives in the transport layer and carries Gin binding validation tags.
type CreateBusRequest struct {
	Name        string          `json:"name"         binding:"required,min=2,max=255"`
	NumberPlate string          `json:"number_plate" binding:"required,min=2,max=50"`
	TotalSeats  int             `json:"total_seats"  binding:"required,min=1,max=100"`
	BusType     domain.BusType  `json:"bus_type"     binding:"required,oneof=standard sleeper ac luxury mini"`
	Description string          `json:"description"  binding:"max=1000"`
}

// ToDomainInput converts the HTTP request to a domain input object.
func (r CreateBusRequest) ToDomainInput() domain.CreateBusInput {
	return domain.CreateBusInput{
		Name:        r.Name,
		NumberPlate: r.NumberPlate,
		TotalSeats:  r.TotalSeats,
		BusType:     r.BusType,
		Description: r.Description,
	}
}

// BusResponse is the JSON response shape for a single bus.
type BusResponse struct {
	ID          string         `json:"id"`
	Name        string         `json:"name"`
	NumberPlate string         `json:"number_plate"`
	TotalSeats  int            `json:"total_seats"`
	BusType     domain.BusType `json:"bus_type"`
	Description string         `json:"description"`
	IsActive    bool           `json:"is_active"`
	CreatedAt   time.Time      `json:"created_at"`
}

// NewBusResponse maps a domain.Bus entity to the API response shape.
func NewBusResponse(b *domain.Bus) *BusResponse {
	return &BusResponse{
		ID:          b.ID,
		Name:        b.Name,
		NumberPlate: b.NumberPlate,
		TotalSeats:  b.TotalSeats,
		BusType:     b.BusType,
		Description: b.Description,
		IsActive:    b.IsActive,
		CreatedAt:   b.CreatedAt,
	}
}
