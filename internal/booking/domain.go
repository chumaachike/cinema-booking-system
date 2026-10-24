package booking

import (
	"errors"
	"time"
)

var (
	ErrSeatAlreadyBooked = errors.New("Serat is already taken")
	ErrrUnableToMarshall = errors.New("Unable to Marshall struct")
)

// Booking represents a confirmed seat reservation.
type Booking struct {
	ID        string
	MovieID   string
	SeatID    string
	UserID    string
	Status    string
	ExpiresAt time.Time
}

type BookingStore interface {
	Book(b Booking) error
	ListBookings(movieID string) []Booking
}
