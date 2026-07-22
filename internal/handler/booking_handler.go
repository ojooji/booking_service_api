package handler

import (
	"encoding/json"
	"net/http"
	"time"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ojooji/booking-service-api/internal/dto"
	"github.com/ojooji/booking-service-api/internal/middleware"
	"github.com/ojooji/booking-service-api/internal/service"
	"github.com/ojooji/booking-service-api/pkg/response"
)

type BookingHandler struct {
	svc *service.BookingService
}

func NewBookingHandler(svc *service.BookingService) *BookingHandler {
	return &BookingHandler{svc: svc}
}

// CreateBooking creates a new booking
// @Summary Create a booking
// @Tags bookings
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body dto.CreateBookingRequest true "Booking payload"
// @Success 201 {object} response.APIResponse
// @Failure 409 {object} response.APIResponse
// @Router /bookings [post]
func (h *BookingHandler) Create(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid claims")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	var req dto.CreateBookingRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	if req.EmployeeID == uuid.Nil || req.ServiceID == uuid.Nil || req.StartTime.IsZero() {
		response.Error(w, http.StatusBadRequest, "VALIDATION", "employee_id, service_id, and start_time are required")
		return
	}

	booking, err := h.svc.Create(r.Context(), claims.UserID, req.EmployeeID, req.ServiceID, req.StartTime)
	if err != nil {
		if err == service.ErrConflict {
			response.Error(w, http.StatusConflict, "CONFLICT", err.Error())
			return
		}
		if err == service.ErrServiceNotFound || err == service.ErrEmployeeNotFound {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		if err == service.ErrPastStartTime || err == service.ErrOutsideHours || err == service.ErrUnalignedStartTime {
			response.Error(w, http.StatusBadRequest, "VALIDATION", err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, booking)
}

// ListMyBookings returns the current user's bookings
// @Summary List my bookings
// @Tags bookings
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.APIResponse
// @Router /bookings [get]
func (h *BookingHandler) ListMy(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid claims")
		return
	}
	bookings, err := h.svc.ListByUser(r.Context(), claims.UserID)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	response.JSON(w, http.StatusOK, bookings)
}

// GetBooking returns a booking by ID
// @Summary Get booking by ID
// @Tags bookings
// @Security BearerAuth
// @Produce json
// @Param id path string true "Booking ID"
// @Success 200 {object} response.APIResponse
// @Router /bookings/{id} [get]
func (h *BookingHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid claims")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid booking id")
		return
	}

	booking, err := h.svc.GetByID(r.Context(), id, claims.UserID, claims.Role)
	if err != nil {
		if err == service.ErrNotAuthorized {
			response.Error(w, http.StatusForbidden, "FORBIDDEN", err.Error())
			return
		}
		if err == service.ErrBookingNotFound {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	response.JSON(w, http.StatusOK, booking)
}

// CancelBooking cancels a booking
// @Summary Cancel a booking
// @Tags bookings
// @Security BearerAuth
// @Param id path string true "Booking ID"
// @Success 200 {object} response.APIResponse
// @Router /bookings/{id}/cancel [put]
func (h *BookingHandler) Cancel(w http.ResponseWriter, r *http.Request) {
	claims, ok := middleware.GetClaims(r.Context())
	if !ok {
		response.Error(w, http.StatusUnauthorized, "UNAUTHORIZED", "invalid claims")
		return
	}

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid booking id")
		return
	}

	if err := h.svc.Cancel(r.Context(), id, claims.UserID, claims.Role); err != nil {
		if err == service.ErrBookingNotFound {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		if err == service.ErrNotAuthorized {
			response.Error(w, http.StatusForbidden, "FORBIDDEN", err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	response.JSON(w, http.StatusOK, map[string]string{"message": "booking cancelled"})
}

// ListAllBookings returns all bookings (admin only)
// @Summary List all bookings
// @Tags bookings
// @Security BearerAuth
// @Produce json
// @Success 200 {object} response.APIResponse
// @Router /bookings/all [get]
func (h *BookingHandler) ListAll(w http.ResponseWriter, r *http.Request) {
	bookings, err := h.svc.ListAll(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	response.JSON(w, http.StatusOK, bookings)
}

// GetSlots returns available time slots for an employee on a given date
// @Summary Get available time slots
// @Tags employees
// @Produce json
// @Param id path string true "Employee ID"
// @Param date query string true "Date (YYYY-MM-DD)"
// @Success 200 {object} response.APIResponse
// @Failure 404 {object} response.APIResponse
// @Router /employees/{id}/slots [get]
func (h *BookingHandler) GetSlots(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid employee id")
		return
	}

	dateStr := r.URL.Query().Get("date")
	date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_DATE", "date must be YYYY-MM-DD")
		return
	}

	slots, err := h.svc.GetSlots(r.Context(), id, date)
	if err != nil {
		if err == service.ErrEmployeeNotFound {
			response.Error(w, http.StatusNotFound, "NOT_FOUND", err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	response.JSON(w, http.StatusOK, slots)
}
