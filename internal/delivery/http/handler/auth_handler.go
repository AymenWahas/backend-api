package handler

import (
	"encoding/json"
	"net/http"

	"backend-api/internal/delivery/http/dto"
	"backend-api/internal/delivery/http/middleware"
	"backend-api/internal/delivery/http/response"
	"backend-api/internal/usecase"
)

type AuthHandler struct {
	authUC *usecase.AuthUsecase
}

func NewAuthHandler(
	authUC *usecase.AuthUsecase,
) *AuthHandler {
	return &AuthHandler{
		authUC: authUC,
	}
}

func (h *AuthHandler) Login(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req dto.LoginRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_JSON",
			"invalid JSON body",
		)
		return
	}

	result, err := h.authUC.Login(
		r.Context(),
		usecase.LoginRequest{
			Email:    req.Email,
			Password: req.Password,
		},
	)

	if err != nil {
		response.WriteError(
			w,
			http.StatusUnauthorized,
			"INVALID_CREDENTIALS",
			"invalid credentials",
		)
		return
	}

	response.WriteJSON(
		w,
		http.StatusOK,
		dto.LoginResponse{
			AccessToken:  result.AccessToken,
			RefreshToken: result.RefreshToken,
		},
	)
}
func (h *AuthHandler) Register(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req dto.RegisterRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_JSON",
			"invalid JSON body",
		)
		return
	}

	user, err := h.authUC.Register(
		r.Context(),
		usecase.RegisterRequest{
			Email:    req.Email,
			Password: req.Password,
		},
	)

	if err != nil {
		response.WriteError(
			w,
			http.StatusBadRequest,
			"REGISTRATION_FAILED",
			err.Error(),
		)
		return
	}

	response.WriteJSON(
		w,
		http.StatusCreated,
		user,
	)
}
func (h *AuthHandler) Me(w http.ResponseWriter, r *http.Request) {
	userID, ok := r.Context().Value(middleware.UserIDContextKey).(int)
	if !ok {
		response.WriteError(
			w,
			http.StatusUnauthorized,
			"UNAUTHENTICATED",
			"authentication required",
		)
		return
	}

	employeeID, ok := r.Context().Value(middleware.EmployeeIDContextKey).(int)
	if !ok {
		response.WriteError(
			w,
			http.StatusUnauthorized,
			"UNAUTHENTICATED",
			"authentication required",
		)
		return
	}

	response.WriteJSON(
		w,
		http.StatusOK,
		map[string]int{
			"user_id":     userID,
			"employee_id": employeeID,
		},
	)
}
func (h *AuthHandler) Refresh(
	w http.ResponseWriter,
	r *http.Request,
) {
	var req dto.RefreshRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_JSON",
			"invalid JSON body",
		)
		return
	}

	result, err := h.authUC.Refresh(
		r.Context(),
		usecase.RefreshRequest{
			RefreshToken: req.RefreshToken,
		},
	)

	if err != nil {
		response.WriteError(
			w,
			http.StatusUnauthorized,
			"INVALID_REFRESH_TOKEN",
			"invalid or expired refresh token",
		)
		return
	}

	response.WriteJSON(
		w,
		http.StatusOK,
		dto.LoginResponse{
			AccessToken:  result.AccessToken,
			RefreshToken: result.RefreshToken,
		},
	)
}
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	var req dto.LogoutRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_JSON",
			"invalid JSON body",
		)
		return
	}

	if err := h.authUC.Logout(
		r.Context(),
		req.RefreshToken,
	); err != nil {
		response.WriteError(
			w,
			http.StatusUnauthorized,
			"INVALID_REFRESH_TOKEN",
			"invalid or expired refresh token",
		)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
