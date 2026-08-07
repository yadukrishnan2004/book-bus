package models

import (
	"time"
)

// BusType represents the category/class of a bus.
type BusType string

const (
	BusTypeStandard BusType = "standard"
	BusTypeSleeper  BusType = "sleeper"
	BusTypeAC       BusType = "ac"
	BusTypeLuxury   BusType = "luxury"
	BusTypeMini     BusType = "mini"
)

// Bus represents a registered bus in the system.
type Bus struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	NumberPlate string    `json:"number_plate"`
	TotalSeats  int       `json:"total_seats"`
	BusType     BusType   `json:"bus_type"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// CreateBusRequest is the payload for registering a new bus.
type CreateBusRequest struct {
	Name        string  `json:"name"         binding:"required,min=2,max=255"`
	NumberPlate string  `json:"number_plate" binding:"required,min=2,max=50"`
	TotalSeats  int     `json:"total_seats"  binding:"required,min=1,max=100"`
	BusType     BusType `json:"bus_type"     binding:"required,oneof=standard sleeper ac luxury mini"`
	Description string  `json:"description"  binding:"max=1000"`
}

// BusResponse is the API response for a single bus.
type BusResponse struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	NumberPlate string    `json:"number_plate"`
	TotalSeats  int       `json:"total_seats"`
	BusType     BusType   `json:"bus_type"`
	Description string    `json:"description"`
	IsActive    bool      `json:"is_active"`
	CreatedAt   time.Time `json:"created_at"`
}
