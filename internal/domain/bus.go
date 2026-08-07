package domain

import (
	"context"
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

// Bus is the core domain entity. It has no JSON or DB tags —
// those belong in the transport and persistence layers.
type Bus struct {
	ID          string
	Name        string
	NumberPlate string
	TotalSeats  int
	BusType     BusType
	Description string
	IsActive    bool
	CreatedAt   time.Time
	UpdatedAt   time.Time
}

// CreateBusInput carries the data needed to create a new bus.
// Used by both the service and repository layers.
type CreateBusInput struct {
	Name        string
	NumberPlate string
	TotalSeats  int
	BusType     BusType
	Description string
}

// BusRepository defines the data-access contract.
// The repository package provides the concrete implementation.
type BusRepository interface {
	Create(ctx context.Context, input CreateBusInput) (*Bus, error)
	GetByID(ctx context.Context, id string) (*Bus, error)
	List(ctx context.Context, limit, offset int) ([]*Bus, error)
}

// BusService defines the business-logic contract.
// The service package provides the concrete implementation.
type BusService interface {
	CreateBus(ctx context.Context, input CreateBusInput) (*Bus, error)
	GetBus(ctx context.Context, id string) (*Bus, error)
	ListBuses(ctx context.Context, limit, offset int) ([]*Bus, error)
}
