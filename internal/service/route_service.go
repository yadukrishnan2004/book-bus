package service

import (
	"context"
	"log/slog"
	"strings"

	"book-bus/internal/domain"
)

type routeService struct {
	repo domain.RouteRepository
}

// NewRouteService creates a routeService implementing domain.RouteService.
func NewRouteService(repo domain.RouteRepository) domain.RouteService {
	return &routeService{repo: repo}
}

func (s *routeService) CreateRoute(ctx context.Context, input domain.CreateRouteInput) (*domain.Route, error) {
	// Bug 4 fix: reject circular routes where origin and destination are the same.
	if strings.EqualFold(strings.TrimSpace(input.Origin), strings.TrimSpace(input.Destination)) {
		return nil, domain.ErrSameOriginDestination
	}
	route, err := s.repo.Create(ctx, input)
	if err != nil {
		slog.Error("service: create route failed", "origin", input.Origin, "destination", input.Destination, "error", err)
		return nil, err
	}
	slog.Info("route created", "id", route.ID, "origin", route.Origin, "destination", route.Destination)
	return route, nil
}


func (s *routeService) ListRoutes(ctx context.Context) ([]*domain.Route, error) {
	routes, err := s.repo.List(ctx)
	if err != nil {
		slog.Error("service: list routes failed", "error", err)
		return nil, err
	}
	return routes, nil
}
