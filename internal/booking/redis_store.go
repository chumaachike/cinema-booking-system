package booking

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/google/uuid"
	"github.com/redis/go-redis/v9"
)

const defaultHoldTTL = 2 * time.Minute

//RedisStore implements session based-seat booking by Redis.
//
// Key design:
//
// Seat: {movieID}:{seatID} -> session JSON (TTL = held, no TTL = confirmed)
// session:{sessionID} -> seat key (reverse lookup)

type RedisStore struct {
	rdb *redis.Client
}

func NewRedisStore(rdb *redis.Client) *RedisStore {
	return &RedisStore{rdb: rdb}
}

// Sessionkey builds the reverse-lookup for a session
func sessionKey(id string) string {
	return fmt.Sprintf("sesssion:s%s", id)
}

func (s *RedisStore) Book(b Booking) error {
	session, err := s.hold(b)
	if err != nil {
		return err
	}

	log.Printf("Session booked %v", session)
	return nil
}

func (s *RedisStore) ListBookings(movieID string) []Booking {
	return []Booking{}
}
func (s *RedisStore) hold(b Booking) (Booking, error) {
	b.ID = uuid.New().String()
	now := time.Now()
	ctx := context.Background()
	key := fmt.Sprintf("seat:%s:%s", b.MovieID, b.SeatID)

	val, err := json.Marshal(b)
	if err != nil {
		return Booking{}, ErrrUnableToMarshall
	}

	res := s.rdb.SetArgs(ctx, key, val, redis.SetArgs{
		Mode: "NX",
		TTL:  defaultHoldTTL,
	})

	if res.Val() != "OK" {
		return Booking{}, ErrSeatAlreadyBooked
	}

	s.rdb.Set(ctx, sessionKey(b.ID), key, defaultHoldTTL)
	return Booking{
		ID:        b.ID,
		MovieID:   b.MovieID,
		SeatID:    b.SeatID,
		UserID:    b.UserID,
		Status:    "held",
		ExpiresAt: now.Add(defaultHoldTTL),
	}, nil
}
