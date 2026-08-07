package service

import (
	"context"
	"log/slog"

	"book-bus/internal/domain"
)

// busService is the concrete implementation of domain.BusService.
// It sits between the HTTP handler and the repository, and is where
// all business logic lives (validation beyond field checks, rules, etc.)
type busService struct {
	repo domain.BusRepository
}

// NewBusService creates a new busService.
// It returns the domain interface so callers stay decoupled from this package.
func NewBusService(repo domain.BusRepository) domain.BusService {
	return &busService{repo: repo}
}

// CreateBus registers a new bus.
func (s *busService) CreateBus(ctx context.Context, input domain.CreateBusInput) (*domain.Bus, error) {
	bus, err := s.repo.Create(ctx, input)
	if err != nil {
		slog.Error("service: failed to create bus",
			"number_plate", input.NumberPlate,
			"error", err,
		)
		return nil, err
	}

	slog.Info("bus registered", "id", bus.ID, "number_plate", bus.NumberPlate)
	return bus, nil
}

// GetBus retrieves a bus by its ID.
func (s *busService) GetBus(ctx context.Context, id string) (*domain.Bus, error) {
	bus, err := s.repo.GetByID(ctx, id)
	if err != nil {
		slog.Error("service: failed to get bus", "id", id, "error", err)
		return nil, err
	}
	return bus, nil
}

// ListBuses retrieves a paginated list of buses.
func (s *busService) ListBuses(ctx context.Context, limit, offset int) ([]*domain.Bus, error) {
	buses, err := s.repo.List(ctx, limit, offset)
	if err != nil {
		slog.Error("service: failed to list buses", "error", err)
		return nil, err
	}
	return buses, nil
}
