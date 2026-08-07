package models

import (
	"time"

	"book-bus/internal/domain"
)

// BookingRequest is the JSON payload used for both preview and confirm.
type BookingRequest struct {
	ScheduleID     string `json:"schedule_id"     binding:"required,uuid"`
	SeatNumbers    []int  `json:"seat_numbers"    binding:"required,min=1,dive,min=1"`
	PassengerName  string `json:"passenger_name"  binding:"required,min=2,max=255"`
	PassengerPhone string `json:"passenger_phone" binding:"required,min=7,max=20"`
}

// ToDomainInput converts the request to a domain input.
func (r BookingRequest) ToDomainInput() domain.CreateBookingInput {
	return domain.CreateBookingInput{
		ScheduleID:     r.ScheduleID,
		SeatNumbers:    r.SeatNumbers,
		PassengerName:  r.PassengerName,
		PassengerPhone: r.PassengerPhone,
	}
}

// ScheduleSummaryResponse is an abbreviated schedule used in booking responses.
type ScheduleSummaryResponse struct {
	Origin        string    `json:"origin"`
	Destination   string    `json:"destination"`
	DepartureTime time.Time `json:"departure_time"`
	ArrivalTime   time.Time `json:"arrival_time"`
	BusName       string    `json:"bus_name"`
	BusType       string    `json:"bus_type"`
}

// BookingPreviewResponse is returned by POST /bookings/preview.
type BookingPreviewResponse struct {
	Available        bool                    `json:"available"`
	UnavailableSeats []int                   `json:"unavailable_seats"`
	ScheduleID       string                  `json:"schedule_id"`
	SeatNumbers      []int                   `json:"seat_numbers"`
	PricePerSeat     float64                 `json:"price_per_seat"`
	TotalPrice       float64                 `json:"total_price"`
	Schedule         ScheduleSummaryResponse `json:"schedule"`
}

// NewBookingPreviewResponse maps a domain.BookingPreview to the API response.
func NewBookingPreviewResponse(p *domain.BookingPreview) *BookingPreviewResponse {
	unavailable := p.UnavailableSeats
	if unavailable == nil {
		unavailable = []int{} // never return null in JSON
	}
	return &BookingPreviewResponse{
		Available:        p.Available,
		UnavailableSeats: unavailable,
		ScheduleID:       p.ScheduleID,
		SeatNumbers:      p.SeatNumbers,
		PricePerSeat:     p.PricePerSeat,
		TotalPrice:       p.TotalPrice,
		Schedule: ScheduleSummaryResponse{
			Origin:        p.Schedule.Origin,
			Destination:   p.Schedule.Destination,
			DepartureTime: p.Schedule.DepartureTime,
			ArrivalTime:   p.Schedule.ArrivalTime,
			BusName:       p.Schedule.BusName,
			BusType:       p.Schedule.BusType,
		},
	}
}

// SeatBookingItem represents a single confirmed seat in the booking response.
type SeatBookingItem struct {
	ID         string  `json:"id"`
	SeatNumber int     `json:"seat_number"`
	Status     string  `json:"status"`
	Price      float64 `json:"price"`
}

// BookingConfirmResponse is returned by POST /bookings (confirm step).
type BookingConfirmResponse struct {
	BookingReference string            `json:"booking_reference"`
	ScheduleID       string            `json:"schedule_id"`
	PassengerName    string            `json:"passenger_name"`
	PassengerPhone   string            `json:"passenger_phone"`
	Seats            []SeatBookingItem `json:"seats"`
	TotalPrice       float64           `json:"total_price"`
	BookedAt         time.Time         `json:"booked_at"`
}

// NewBookingConfirmResponse builds a BookingConfirmResponse from a slice of bookings.
func NewBookingConfirmResponse(bookings []*domain.Booking, reference string) *BookingConfirmResponse {
	if len(bookings) == 0 {
		return nil
	}
	first := bookings[0]
	var total float64
	seats := make([]SeatBookingItem, 0, len(bookings))
	for _, b := range bookings {
		seats = append(seats, SeatBookingItem{
			ID:         b.ID,
			SeatNumber: b.SeatNumber,
			Status:     string(b.Status),
			Price:      b.TotalPrice,
		})
		total += b.TotalPrice
	}
	return &BookingConfirmResponse{
		BookingReference: reference,
		ScheduleID:       first.ScheduleID,
		PassengerName:    first.PassengerName,
		PassengerPhone:   first.PassengerPhone,
		Seats:            seats,
		TotalPrice:       total,
		BookedAt:         first.BookedAt,
	}
}
