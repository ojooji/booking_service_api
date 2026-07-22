package handler

import (
	"encoding/json"
	"net/http"

	"github.com/ojooji/booking-service-api/internal/dto"
	"github.com/ojooji/booking-service-api/internal/service"
	"github.com/ojooji/booking-service-api/pkg/response"
)

const maxBodyBytes = 1 << 20

type AuthHandler struct {
	authSvc *service.AuthService
}

func NewAuthHandler(authSvc *service.AuthService) *AuthHandler {
	return &AuthHandler{authSvc: authSvc}
}

// Register creates a new user account
// @Summary Register a new user
// @Tags auth
// @Accept json
// @Produce json
// @Param body body dto.RegisterRequest true "Registration payload"
// @Success 201 {object} response.APIResponse
// @Failure 400 {object} response.APIResponse
// @Failure 409 {object} response.APIResponse
// @Router /auth/register [post]
func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	var req dto.RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	if err := dto.ValidateRegister(req); err != nil {
		response.Error(w, http.StatusBadRequest, "VALIDATION", err.Error())
		return
	}

	user, err := h.authSvc.Register(r.Context(), req.Email, req.Password, req.FirstName, req.LastName)
	if err != nil {
		if err == service.ErrEmailTaken {
			response.Error(w, http.StatusConflict, "EMAIL_TAKEN", err.Error())
			return
		}
		response.Error(w, http.StatusInternalServerError, "INTERNAL", err.Error())
		return
	}

	response.JSON(w, http.StatusCreated, user)
}

// Login authenticates a user and returns a JWT
// @Summary Login and get JWT token
// @Tags auth
// @Accept json
// @Produce json
// @Param body body dto.LoginRequest true "Login payload"
// @Success 200 {object} response.APIResponse
// @Failure 401 {object} response.APIResponse
// @Router /auth/login [post]
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	r.Body = http.MaxBytesReader(w, r.Body, maxBodyBytes)

	var req dto.LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.Error(w, http.StatusBadRequest, "INVALID_BODY", "invalid request body")
		return
	}

	if req.Email == "" || req.Password == "" {
		response.Error(w, http.StatusBadRequest, "VALIDATION", "email and password are required")
		return
	}

	token, user, err := h.authSvc.Login(r.Context(), req.Email, req.Password)
	if err != nil {
		response.Error(w, http.StatusUnauthorized, "INVALID_CREDS", err.Error())
		return
	}

	resp := dto.AuthResponse{
		AccessToken: token,
		UserID:      user.ID.String(),
		Role:        string(user.Role),
	}
	response.JSON(w, http.StatusOK, resp)
}
