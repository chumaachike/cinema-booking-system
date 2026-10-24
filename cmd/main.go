package main

import (
	"log"
	"net/http"

	"github.com/chumaachike/cinema-booking-system/internal/adapters/redis"
	"github.com/chumaachike/cinema-booking-system/internal/booking"
)

func main() {
	mux := http.NewServeMux()

	store := booking.NewRedisStore(redis.NewClient("localhost:6379"))
	svc := booking.NewService(store)
	bookingHandler := booking.NewHandler(svc)

	mux.HandleFunc("GET /movies", bookingHandler.ListMovies)
	mux.HandleFunc("GET /movies/{movieID}/seats", bookingHandler.ListSeats)
	mux.HandleFunc("POST /movies/{movieID}/seats/{seatID}/hold", bookingHandler.HoldSeat)

	mux.Handle("GET /", http.FileServer(http.Dir("static")))
	if err := http.ListenAndServe(":8080", mux); err != nil {
		log.Fatal(err)
	}
}
