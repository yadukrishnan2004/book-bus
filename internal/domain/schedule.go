package domain

import (
	"context"
	"time"
)

// ScheduleStatus represents the current state of a bus schedule.
type ScheduleStatus string

const (
	ScheduleStatusScheduled ScheduleStatus = "scheduled"
	ScheduleStatusOngoing   ScheduleStatus = "ongoing"
	ScheduleStatusCompleted ScheduleStatus = "completed"
	ScheduleStatusCancelled ScheduleStatus = "cancelled"
)

// Schedule represents a bus running on a specific route at a specific time.
type Schedule struct {
	ID             string
	BusID          string
	RouteID        string
	DepartureTime  time.Time
	ArrivalTime    time.Time
	Price          float64
	AvailableSeats int
	TotalSeats     int // populated from the linked Bus
	Status         ScheduleStatus
	// Joined relations — populated by List and GetByID
	Bus   *Bus
	Route *Route
	CreatedAt time.Time
	UpdatedAt time.Time
}

// SeatInfo describes a single seat on a schedule.
type SeatInfo struct {
	Number      int
	IsAvailable bool
}

// CreateScheduleInput carries the fields needed to create a schedule.
type CreateScheduleInput struct {
	BusID         string
	RouteID       string
	DepartureTime time.Time
	ArrivalTime   time.Time
	Price         float64
}

// ScheduleFilter is used to search/filter schedules.
type ScheduleFilter struct {
	Origin      string // partial match
	Destination string // partial match
	Date        string // YYYY-MM-DD (optional)
}

// ScheduleRepository defines the data-access contract for schedules.
type ScheduleRepository interface {
	Create(ctx context.Context, input CreateScheduleInput) (*Schedule, error)
	GetByID(ctx context.Context, id string) (*Schedule, error)
	List(ctx context.Context, filter ScheduleFilter, limit, offset int) ([]*Schedule, error)
	GetSeatMap(ctx context.Context, scheduleID string) ([]SeatInfo, error)
	UpdateStatus(ctx context.Context, id string, status ScheduleStatus) (*Schedule, int, error)
}

// ScheduleService defines the business-logic contract for schedules.
type ScheduleService interface {
	CreateSchedule(ctx context.Context, input CreateScheduleInput) (*Schedule, error)
	GetSchedule(ctx context.Context, id string) (*Schedule, error)
	ListSchedules(ctx context.Context, filter ScheduleFilter, limit, offset int) ([]*Schedule, error)
	GetSeatMap(ctx context.Context, scheduleID string) ([]SeatInfo, error)
	UpdateStatus(ctx context.Context, id string, status ScheduleStatus) (*Schedule, int, error)
}
