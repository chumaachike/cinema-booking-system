package booking

import (
	"fmt"
	"maps"
	"slices"
	"sync"
)

type MemoryStore struct {
	bookings map[string]Booking
	sync.RWMutex
}

func NewMemoryStore() *MemoryStore {
	return &MemoryStore{
		bookings: map[string]Booking{},
	}
}
func (s *MemoryStore) Book(b Booking) error {
	s.Lock()
	defer s.Unlock()
	if _, ok := s.bookings[b.SeatID]; ok {
		return fmt.Errorf("booking has already been created earlier")
	}
	s.bookings[b.SeatID] = b
	return nil
}

func (s *MemoryStore) ListBookings(MovieID string) []Booking {
	s.RLock()
	defer s.RUnlock()
	return slices.Collect(maps.Values(s.bookings))
}
