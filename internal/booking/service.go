package booking

import (
	"context"
	"fmt"
)

type Service struct {
	store BookingStore
}

func NewService(store BookingStore) *Service {
	return &Service{store}
}

func (s *Service) Book(b Booking) (Session, error) {
	return s.store.Book(b)
}

func (s *Service) ListBooking(movieID string) []Booking {
	return s.store.ListBookings(movieID)
}

func (s *Service) ConfirmSession(ctx context.Context, sessionID string) (Booking, error) {
	if sessionID == "" {
		return Booking{}, ErrInvalidSession
	}

	booking, err := s.store.ConfirmSession(sessionID)
	if err != nil {
		return Booking{}, fmt.Errorf("confirm session: %w", err)
	}

	return booking, nil
}

func (s *Service) ReleaseSession(ctx context.Context, sessionID string) error {
	if sessionID == "" {
		return ErrInvalidSession
	}

	if err := s.store.ReleaseSession(sessionID); err != nil {
		return fmt.Errorf("release session: %w", err)
	}

	return nil
}
