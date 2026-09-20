package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"backend-api/internal/delivery/http/middleware"
	"backend-api/internal/delivery/http/response"
	"backend-api/internal/domain"
	"backend-api/internal/usecase"
)

type ProjectMembershipHandler struct {
	usecase *usecase.ProjectMembershipUsecase
}

func NewProjectMembershipHandler(
	uc *usecase.ProjectMembershipUsecase,
) *ProjectMembershipHandler {
	return &ProjectMembershipHandler{
		usecase: uc,
	}
}

type AddProjectMemberRequest struct {
	EmployeeID int    `json:"employee_id"`
	Role       string `json:"role"`
}

type UpdateProjectMemberRoleRequest struct {
	Role string `json:"role"`
}

func (h *ProjectMembershipHandler) GetMembers(
	w http.ResponseWriter,
	r *http.Request,
) {
	projectID, err := strconv.ParseUint(
		r.PathValue("id"),
		10,
		64,
	)
	if err != nil || projectID == 0 {
		response.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_ID",
			"invalid project id",
		)
		return
	}

	userID, ok := r.Context().Value(
		middleware.UserIDContextKey,
	).(int)

	if !ok {
		response.WriteError(
			w,
			http.StatusUnauthorized,
			"UNAUTHORIZED",
			"authentication required",
		)
		return
	}

	members, err := h.usecase.GetMembers(
		r.Context(),
		userID,
		uint(projectID),
	)
	if err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			response.WriteError(
				w,
				http.StatusForbidden,
				"FORBIDDEN",
				"project access denied",
			)
			return
		}

		response.WriteError(
			w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"internal server error",
		)
		return
	}

	response.WriteJSON(
		w,
		http.StatusOK,
		members,
	)
}

func (h *ProjectMembershipHandler) AddMember(
	w http.ResponseWriter,
	r *http.Request,
) {
	projectID, err := strconv.ParseUint(
		r.PathValue("id"),
		10,
		64,
	)
	if err != nil || projectID == 0 {
		response.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_ID",
			"invalid project id",
		)
		return
	}

	userID, ok := r.Context().Value(
		middleware.UserIDContextKey,
	).(int)

	if !ok {
		response.WriteError(
			w,
			http.StatusUnauthorized,
			"UNAUTHORIZED",
			"authentication required",
		)
		return
	}

	var req AddProjectMemberRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_JSON",
			"invalid JSON body",
		)
		return
	}

	if req.EmployeeID <= 0 || req.Role == "" {
		response.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_MEMBER",
			"employee_id and role are required",
		)
		return
	}

	if err := h.usecase.AddMember(
		r.Context(),
		userID,
		uint(projectID),
		req.EmployeeID,
		req.Role,
	); err != nil {

		if errors.Is(err, domain.ErrForbidden) {
			response.WriteError(
				w,
				http.StatusForbidden,
				"FORBIDDEN",
				"member management denied",
			)
			return
		}

		if errors.Is(err, domain.ErrProjectNotFound) {
			response.WriteError(
				w,
				http.StatusNotFound,
				"PROJECT_NOT_FOUND",
				"project not found",
			)
			return
		}

		if errors.Is(err, domain.ErrInvalidProjectRole) {
			response.WriteError(
				w,
				http.StatusBadRequest,
				"INVALID_ROLE",
				"role must be one of member, manager, admin",
			)
			return
		}

		if errors.Is(err, domain.ErrEmployeeNotFound) {
			response.WriteError(
				w,
				http.StatusNotFound,
				"EMPLOYEE_NOT_FOUND",
				"employee not found",
			)
			return
		}

		response.WriteError(
			w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"internal server error",
		)
		return
	}

	response.WriteJSON(
		w,
		http.StatusCreated,
		map[string]string{
			"status": "member added",
		},
	)
}

func (h *ProjectMembershipHandler) UpdateRole(
	w http.ResponseWriter,
	r *http.Request,
) {
	projectID, err := strconv.ParseUint(
		r.PathValue("id"),
		10,
		64,
	)
	if err != nil || projectID == 0 {
		response.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_ID",
			"invalid project id",
		)
		return
	}

	employeeID, err := strconv.Atoi(
		r.PathValue("employeeId"),
	)
	if err != nil || employeeID <= 0 {
		response.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_EMPLOYEE_ID",
			"invalid employee id",
		)
		return
	}

	userID, ok := r.Context().Value(
		middleware.UserIDContextKey,
	).(int)

	if !ok {
		response.WriteError(
			w,
			http.StatusUnauthorized,
			"UNAUTHORIZED",
			"authentication required",
		)
		return
	}

	var req UpdateProjectMemberRoleRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_JSON",
			"invalid JSON body",
		)
		return
	}

	if req.Role == "" {
		response.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_ROLE",
			"role is required",
		)
		return
	}

	if err := h.usecase.UpdateRole(
		r.Context(),
		userID,
		uint(projectID),
		employeeID,
		req.Role,
	); err != nil {

		if errors.Is(err, domain.ErrForbidden) {
			response.WriteError(
				w,
				http.StatusForbidden,
				"FORBIDDEN",
				"member management denied",
			)
			return
		}

		if errors.Is(err, domain.ErrProjectNotFound) {
			response.WriteError(
				w,
				http.StatusNotFound,
				"PROJECT_NOT_FOUND",
				"project not found",
			)
			return
		}

		if errors.Is(err, domain.ErrInvalidProjectRole) {
			response.WriteError(
				w,
				http.StatusBadRequest,
				"INVALID_ROLE",
				"role must be one of member, manager, admin",
			)
			return
		}

		if errors.Is(err, domain.ErrMembershipNotFound) {
			response.WriteError(
				w,
				http.StatusNotFound,
				"MEMBERSHIP_NOT_FOUND",
				"membership not found",
			)
			return
		}

		response.WriteError(
			w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"internal server error",
		)
		return
	}

	response.WriteJSON(
		w,
		http.StatusOK,
		map[string]string{
			"status": "role updated",
		},
	)
}

func (h *ProjectMembershipHandler) RemoveMember(
	w http.ResponseWriter,
	r *http.Request,
) {
	projectID, err := strconv.ParseUint(
		r.PathValue("id"),
		10,
		64,
	)
	if err != nil || projectID == 0 {
		response.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_ID",
			"invalid project id",
		)
		return
	}

	employeeID, err := strconv.Atoi(
		r.PathValue("employeeId"),
	)
	if err != nil || employeeID <= 0 {
		response.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_EMPLOYEE_ID",
			"invalid employee id",
		)
		return
	}

	userID, ok := r.Context().Value(
		middleware.UserIDContextKey,
	).(int)

	if !ok {
		response.WriteError(
			w,
			http.StatusUnauthorized,
			"UNAUTHORIZED",
			"authentication required",
		)
		return
	}

	if err := h.usecase.RemoveMember(
		r.Context(),
		userID,
		uint(projectID),
		employeeID,
	); err != nil {

		if errors.Is(err, domain.ErrForbidden) {
			response.WriteError(
				w,
				http.StatusForbidden,
				"FORBIDDEN",
				"member management denied",
			)
			return
		}

		if errors.Is(err, domain.ErrProjectNotFound) {
			response.WriteError(
				w,
				http.StatusNotFound,
				"PROJECT_NOT_FOUND",
				"project not found",
			)
			return
		}

		if errors.Is(err, domain.ErrMembershipNotFound) {
			response.WriteError(
				w,
				http.StatusNotFound,
				"MEMBERSHIP_NOT_FOUND",
				"membership not found",
			)
			return
		}

		response.WriteError(
			w,
			http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"internal server error",
		)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
