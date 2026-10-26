package booking

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log"
	"time"

	goredis "github.com/chumaachike/cinema-booking-system/internal/adapters/redis"
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

func (s *RedisStore) Book(b Booking) (Session, error) {
	session, err := s.hold(b)
	if err != nil {
		return Session{}, err
	}

	log.Printf("Session booked %v", session)
	return session, nil
}

func (s *RedisStore) ListBookings(movieID string) []Booking {
	pattern := fmt.Sprintf("seat:%s:*", movieID)
	var sessions []Booking

	ctx := context.Background()

	iter := s.rdb.Scan(ctx, 0, pattern, 0).Iterator()
	for iter.Next(ctx) {
		val, err := s.rdb.Get(ctx, iter.Val()).Result()
		if err != nil {
			continue
		}
		session, err := parseSession(val)
		if err != nil {
			continue
		}

		sessions = append(sessions, session)
	}
	return sessions
}

func (s *RedisStore) ConfirmSession(sessionID string) (Booking, error) {
	ctx := context.Background()

	result, err := goredis.ConfirmScript.Run(
		ctx,
		s.rdb,
		[]string{sessionKey(sessionID)},
		sessionID,
	).Result()

	if err != nil {
		return Booking{}, fmt.Errorf("confirm session: %w", err)
	}

	data, ok := result.(string)
	if !ok {
		return Booking{}, errors.New("unexpected Redis response")
	}

	var booking Booking

	if err := json.Unmarshal([]byte(data), &booking); err != nil {
		return Booking{}, err
	}

	return booking, nil
}

func (s *RedisStore) hold(b Booking) (Session, error) {
	b.ID = uuid.New().String()
	b.Status = "held"
	now := time.Now()
	ctx := context.Background()
	key := fmt.Sprintf("seat:%s:%s", b.MovieID, b.SeatID)

	val, err := json.Marshal(b)
	if err != nil {
		return Session{}, ErrrUnableToMarshall
	}

	res := s.rdb.SetArgs(ctx, key, val, redis.SetArgs{
		Mode: "NX",
		TTL:  defaultHoldTTL,
	})

	if res.Val() != "OK" {
		return Session{}, ErrSeatAlreadyBooked
	}

	s.rdb.Set(ctx, sessionKey(b.ID), key, defaultHoldTTL)
	return Session{
		Booking: Booking{
			ID:      b.ID,
			MovieID: b.MovieID,
			SeatID:  b.SeatID,
			UserID:  b.UserID,
			Status:  "held",
		},
		ExpiresAt: now.Add(defaultHoldTTL),
	}, nil
}

func parseSession(val string) (Booking, error) {
	var data Booking
	if err := json.Unmarshal([]byte(val), &data); err != nil {
		return Booking{}, err
	}
	return Booking{
		ID:      data.ID,
		MovieID: data.MovieID,
		SeatID:  data.SeatID,
		UserID:  data.UserID,
		Status:  data.Status,
	}, nil
}

func (s *RedisStore) ReleaseSession(sessionID string) error {
	ctx, cancel := context.WithTimeout(
		context.Background(),
		3*time.Second,
	)
	defer cancel()

	result, err := goredis.ReleaseScript.Run(
		ctx,
		s.rdb,
		[]string{sessionKey(sessionID)},
		sessionID,
	).Int()

	if err != nil {
		return fmt.Errorf("release session: %w", err)
	}

	switch result {
	case 1:
		return nil

	case 0:
		return ErrSessionNotFound

	case -1:
		return ErrInvalidSession

	case -2:
		return ErrBookingNotHeld

	default:
		return fmt.Errorf("unexpected release result: %d", result)
	}
}
