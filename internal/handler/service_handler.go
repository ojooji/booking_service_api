package handler

import (
	"encoding/json"
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
	"github.com/ojooji/booking-service-api/internal/dto"
	"github.com/ojooji/booking-service-api/internal/service"
	"github.com/ojooji/booking-service-api/pkg/response"
)

type ServiceHandler struct {
	svc *service.ServiceService
}

func NewServiceHandler(svc *service.ServiceService) *ServiceHandler {
	return &ServiceHandler{svc: svc}
}

// ListServices returns all services
// @Summary List all services
// @Tags services
// @Produce json
// @Success 200 {object} response.APIResponse
// @Router /services [get]
func (h *ServiceHandler) List(w http.ResponseWriter, r *http.Request) {
	services, err := h.svc.List(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	response.JSON(w, http.StatusOK, services)
}

// GetService returns a service by ID
// @Summary Get service by ID
// @Tags services
// @Produce json
// @Param id path string true "Service ID"
// @Success 200 {object} response.APIResponse
// @Router /services/{id} [get]
func (h *ServiceHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid service id")
		return
	}

	svc, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}
	response.JSON(w, http.StatusOK, svc)
}

// CreateService creates a new service
// @Summary Create a service
// @Tags services
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body dto.CreateServiceRequest true "Service payload"
// @Success 201 {object} response.APIResponse
// @Router /services [post]
func (h *ServiceHandler) Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	var req dto.CreateServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	if err := dto.ValidateCreateService(req); err != nil {
		response.Error(w, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}

	svc, err := h.svc.Create(r.Context(), req.Name, req.Description, req.Duration, req.Price)
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, svc)
}

// UpdateService updates a service
// @Summary Update a service
// @Tags services
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Service ID"
// @Param body body dto.UpdateServiceRequest true "Update payload"
// @Success 200 {object} response.APIResponse
// @Router /services/{id} [put]
func (h *ServiceHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid service id")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	var req dto.UpdateServiceRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	if err := dto.ValidateUpdateService(req); err != nil {
		response.Error(w, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}

	svc, err := h.svc.Update(r.Context(), id, req.Name, req.Description, req.Duration, req.Price)
	if err != nil {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}
	response.JSON(w, http.StatusOK, svc)
}

// DeleteService deletes a service
// @Summary Delete a service
// @Tags services
// @Security BearerAuth
// @Param id path string true "Service ID"
// @Success 204
// @Router /services/{id} [delete]
func (h *ServiceHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid service id")
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
