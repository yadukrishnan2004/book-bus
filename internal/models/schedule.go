package models

import (
	"time"

	"book-bus/internal/domain"
)

// CreateScheduleRequest is the JSON payload for creating a schedule.
type CreateScheduleRequest struct {
	BusID         string    `json:"bus_id"         binding:"required,uuid"`
	RouteID       string    `json:"route_id"       binding:"required,uuid"`
	DepartureTime time.Time `json:"departure_time" binding:"required"`
	ArrivalTime   time.Time `json:"arrival_time"   binding:"required"`
	Price         float64   `json:"price"          binding:"required,min=1"`
}

// ToDomainInput converts the request to a domain input.
func (r CreateScheduleRequest) ToDomainInput() domain.CreateScheduleInput {
	return domain.CreateScheduleInput{
		BusID:         r.BusID,
		RouteID:       r.RouteID,
		DepartureTime: r.DepartureTime,
		ArrivalTime:   r.ArrivalTime,
		Price:         r.Price,
	}
}

// BusInfo is an embedded bus summary inside schedule responses.
type BusInfo struct {
	ID          string `json:"id"`
	Name        string `json:"name"`
	NumberPlate string `json:"number_plate"`
	TotalSeats  int    `json:"total_seats"`
	BusType     string `json:"bus_type"`
}

// RouteInfo is an embedded route summary inside schedule responses.
type RouteInfo struct {
	ID          string  `json:"id"`
	Origin      string  `json:"origin"`
	Destination string  `json:"destination"`
	DistanceKM  float64 `json:"distance_km"`
}

// ScheduleResponse is the full API response for a schedule.
type ScheduleResponse struct {
	ID             string                `json:"id"`
	DepartureTime  time.Time             `json:"departure_time"`
	ArrivalTime    time.Time             `json:"arrival_time"`
	Price          float64               `json:"price"`
	AvailableSeats int                   `json:"available_seats"`
	Status         domain.ScheduleStatus `json:"status"`
	Bus            BusInfo               `json:"bus"`
	Route          RouteInfo             `json:"route"`
	CreatedAt      time.Time             `json:"created_at"`
}

// NewScheduleResponse maps a domain.Schedule to the API response.
func NewScheduleResponse(s *domain.Schedule) *ScheduleResponse {
	r := &ScheduleResponse{
		ID:             s.ID,
		DepartureTime:  s.DepartureTime,
		ArrivalTime:    s.ArrivalTime,
		Price:          s.Price,
		AvailableSeats: s.AvailableSeats,
		Status:         s.Status,
		CreatedAt:      s.CreatedAt,
	}
	if s.Bus != nil {
		r.Bus = BusInfo{
			ID:          s.Bus.ID,
			Name:        s.Bus.Name,
			NumberPlate: s.Bus.NumberPlate,
			TotalSeats:  s.TotalSeats,
			BusType:     string(s.Bus.BusType),
		}
	}
	if s.Route != nil {
		r.Route = RouteInfo{
			ID:          s.Route.ID,
			Origin:      s.Route.Origin,
			Destination: s.Route.Destination,
			DistanceKM:  s.Route.DistanceKM,
		}
	}
	return r
}

// SeatMapResponse is the API response for the seat availability map.
type SeatMapResponse struct {
	ScheduleID     string     `json:"schedule_id"`
	TotalSeats     int        `json:"total_seats"`
	AvailableSeats int        `json:"available_seats"`
	Seats          []SeatItem `json:"seats"`
}

// SeatItem represents one seat in the map.
type SeatItem struct {
	Number      int  `json:"number"`
	IsAvailable bool `json:"is_available"`
}
