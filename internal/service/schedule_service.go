package service

import (
	"context"
	"log/slog"

	"book-bus/internal/domain"
)

type scheduleService struct {
	repo domain.ScheduleRepository
}

// NewScheduleService creates a scheduleService implementing domain.ScheduleService.
func NewScheduleService(repo domain.ScheduleRepository) domain.ScheduleService {
	return &scheduleService{repo: repo}
}

func (s *scheduleService) CreateSchedule(ctx context.Context, input domain.CreateScheduleInput) (*domain.Schedule, error) {
	if input.ArrivalTime.Before(input.DepartureTime) {
		return nil, domain.ErrArrivalBeforeDeparture
	}
	schedule, err := s.repo.Create(ctx, input)
	if err != nil {
		slog.Error("service: create schedule failed", "bus_id", input.BusID, "error", err)
		return nil, err
	}
	slog.Info("schedule created", "id", schedule.ID, "bus_id", input.BusID, "route_id", input.RouteID)
	return schedule, nil
}

func (s *scheduleService) GetSchedule(ctx context.Context, id string) (*domain.Schedule, error) {
	schedule, err := s.repo.GetByID(ctx, id)
	if err != nil {
		slog.Error("service: get schedule failed", "id", id, "error", err)
		return nil, err
	}
	return schedule, nil
}

func (s *scheduleService) ListSchedules(ctx context.Context, filter domain.ScheduleFilter, limit, offset int) ([]*domain.Schedule, error) {
	schedules, err := s.repo.List(ctx, filter, limit, offset)
	if err != nil {
		slog.Error("service: list schedules failed", "error", err)
		return nil, err
	}
	return schedules, nil
}

func (s *scheduleService) GetSeatMap(ctx context.Context, scheduleID string) ([]domain.SeatInfo, error) {
	seats, err := s.repo.GetSeatMap(ctx, scheduleID)
	if err != nil {
		slog.Error("service: get seat map failed", "schedule_id", scheduleID, "error", err)
		return nil, err
	}
	return seats, nil
}

func (s *scheduleService) UpdateStatus(ctx context.Context, id string, status domain.ScheduleStatus) (*domain.Schedule, int, error) {
	schedule, affected, err := s.repo.UpdateStatus(ctx, id, status)
	if err != nil {
		slog.Error("service: update schedule status failed", "id", id, "status", status, "error", err)
		return nil, 0, err
	}
	slog.Info("schedule status updated", "id", id, "new_status", status, "affected_bookings", affected)
	return schedule, affected, nil
}

