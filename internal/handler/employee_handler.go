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

type EmployeeHandler struct {
	svc *service.EmployeeService
}

func NewEmployeeHandler(svc *service.EmployeeService) *EmployeeHandler {
	return &EmployeeHandler{svc: svc}
}

// ListEmployees returns all employees
// @Summary List all employees
// @Tags employees
// @Produce json
// @Success 200 {object} response.APIResponse
// @Router /employees [get]
func (h *EmployeeHandler) List(w http.ResponseWriter, r *http.Request) {
	emps, err := h.svc.List(r.Context())
	if err != nil {
		response.Error(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	response.JSON(w, http.StatusOK, emps)
}

// GetEmployee returns an employee by ID
// @Summary Get employee by ID
// @Tags employees
// @Produce json
// @Param id path string true "Employee ID"
// @Success 200 {object} response.APIResponse
// @Router /employees/{id} [get]
func (h *EmployeeHandler) GetByID(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid employee id")
		return
	}

	emp, err := h.svc.GetByID(r.Context(), id)
	if err != nil {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}
	response.JSON(w, http.StatusOK, emp)
}

// CreateEmployee creates a new employee
// @Summary Create an employee
// @Tags employees
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param body body dto.CreateEmployeeRequest true "Employee payload"
// @Success 201 {object} response.APIResponse
// @Router /employees [post]
func (h *EmployeeHandler) Create(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	var req dto.CreateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	userID, err := uuid.Parse(req.UserID)
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_USER_ID", "invalid user id")
		return
	}

	emp, err := h.svc.Create(r.Context(), userID, req.Bio, req.ServiceIDs)
	if err != nil {
		if err == service.ErrEmployeeExists {
			response.Error(w, http.StatusConflict, "EMPLOYEE_EXISTS", err.Error())
			return
		}
		if err == service.ErrUserNotFound {
			response.Error(w, http.StatusNotFound, "USER_NOT_FOUND", err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}
	response.JSON(w, http.StatusCreated, emp)
}

// UpdateEmployee updates an employee
// @Summary Update an employee
// @Tags employees
// @Security BearerAuth
// @Accept json
// @Produce json
// @Param id path string true "Employee ID"
// @Param body body dto.UpdateEmployeeRequest true "Update payload"
// @Success 200 {object} response.APIResponse
// @Router /employees/{id} [put]
func (h *EmployeeHandler) Update(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid employee id")
		return
	}

	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	var req dto.UpdateEmployeeRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	emp, err := h.svc.Update(r.Context(), id, req.Bio, req.ServiceIDs)
	if err != nil {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}
	response.JSON(w, http.StatusOK, emp)
}

// DeleteEmployee deletes an employee
// @Summary Delete an employee
// @Tags employees
// @Security BearerAuth
// @Param id path string true "Employee ID"
// @Success 204
// @Router /employees/{id} [delete]
func (h *EmployeeHandler) Delete(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_ID", "invalid employee id")
		return
	}

	if err := h.svc.Delete(r.Context(), id); err != nil {
		response.Error(w, http.StatusNotFound, "NOT_FOUND", err.Error())
		return
	}
	w.WriteHeader(http.StatusNoContent)
}
