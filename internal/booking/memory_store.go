package booking

import (
	"fmt"
	"maps"
	"slices"
)

type MemoryStore struct {
	bookings map[string]Booking
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		bookings: map[string]Booking{},
	}
}
func (s *MemoryStore) Book(b Booking) error {
	if _, ok := s.bookings[b.SeatID]; ok {
		return fmt.Errorf("booking has already been created earlier")
	}
	s.bookings[b.SeatID] = b
	return nil
}

func (s *MemoryStore) ListBookings(MovieID string) []Booking {
	return slices.Collect(maps.Values(s.bookings))

}
