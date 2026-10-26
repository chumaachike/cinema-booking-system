package booking

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/chumaachike/cinema-booking-system/internal/utils"
)

type Handler struct {
	svc *Service
}

func NewHandler(svc *Service) *Handler {
	return &Handler{svc}
}

func (h *Handler) ListMovies(w http.ResponseWriter, r *http.Request) {
	utils.WriteJSON(w, http.StatusOK, movies)
}

func (h *Handler) ConfirmSession(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("sessionID")
	booking, err := h.svc.ConfirmSession(r.Context(), sessionID)
	if err != nil {
		panic(err)
	}

	utils.WriteJSON(w, http.StatusCreated, booking)

}

func (h *Handler) ReleaseSession(w http.ResponseWriter, r *http.Request) {
	sessionID := r.PathValue("sessionID")

	if sessionID == "" {
		http.Error(w, "session ID is required", http.StatusBadRequest)
		return
	}

	err := h.svc.ReleaseSession(r.Context(), sessionID)
	if err != nil {
		switch {
		case errors.Is(err, ErrSessionNotFound):
			http.Error(w, "session not found", http.StatusNotFound)

		case errors.Is(err, ErrInvalidSession):
			http.Error(w, "invalid session", http.StatusForbidden)

		case errors.Is(err, ErrBookingNotHeld):
			http.Error(w, "booking is not held", http.StatusConflict)

		default:
			http.Error(w, "internal server error", http.StatusInternalServerError)
		}
		return
	}

	w.WriteHeader(http.StatusNoContent)
}

func (h *Handler) HoldSeat(w http.ResponseWriter, r *http.Request) {
	seatID := r.PathValue("seatID")
	movieID := r.PathValue("movieID")

	var req holdRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		log.Println(err)
		return
	}

	data := Booking{
		UserID:  req.UserId,
		SeatID:  seatID,
		MovieID: movieID,
	}

	session, err := h.svc.Book(data)
	if err != nil {
		log.Println(err)
		return
	}

	utils.WriteJSON(w, http.StatusCreated, holdResponse{
		SeatID:    seatID,
		MovieID:   session.MovieID,
		SessionID: session.ID,
		ExpiresAt: session.ExpiresAt.Format(time.RFC3339),
	})
}

func (h *Handler) ListSeats(w http.ResponseWriter, r *http.Request) {
	movieID := r.PathValue("movieID")
	bookings := h.svc.ListBooking(movieID)

	seats := make([]seatInfo, 0, len(bookings))
	for _, b := range bookings {
		seats = append(seats, seatInfo{
			SeatID: b.SeatID,
			UserID: b.UserID,
			Booked: true,
		})

	}

	utils.WriteJSON(w, http.StatusOK, seats)
}

var movies = []movieResponse{
	{ID: "inception", Title: "Inception", Rows: 5, SeatsPerRow: 8},
	{ID: "dune", Title: "Dune: Part Two", Rows: 4, SeatsPerRow: 6},
}
