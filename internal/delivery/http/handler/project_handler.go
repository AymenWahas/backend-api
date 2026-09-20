package handler

import (
	"encoding/json"
	"errors"
	"net/http"
	"strconv"

	"backend-api/internal/delivery/http/dto"
	"backend-api/internal/delivery/http/middleware"
	"backend-api/internal/delivery/http/response"
	"backend-api/internal/domain"
	"backend-api/internal/usecase"
)

type ProjectHandler struct {
	usecase *usecase.ProjectUsecase
}

func NewProjectHandler(
	uc *usecase.ProjectUsecase,
) *ProjectHandler {
	return &ProjectHandler{
		usecase: uc,
	}
}

func getUserID(r *http.Request) (int, bool) {
	userID, ok := r.Context().
		Value(middleware.UserIDContextKey).(int)

	return userID, ok
}

func (h *ProjectHandler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
	employeeID, ok := r.Context().
		Value(middleware.EmployeeIDContextKey).(int)

	if !ok {
		response.WriteError(
			w,
			http.StatusUnauthorized,
			"UNAUTHORIZED",
			"authentication required",
		)
		return
	}

	var req dto.CreateProjectRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_JSON",
			"invalid JSON body",
		)
		return
	}

	if req.Name == "" {
		response.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_PROJECT",
			"project name is required",
		)
		return
	}

	project := &domain.Project{
		OwnerID:     employeeID,
		Name:        req.Name,
		Description: req.Description,
	}

	if err := h.usecase.Create(
		r.Context(),
		project,
	); err != nil {
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
		project,
	)
}

func (h *ProjectHandler) GetAll(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, ok := getUserID(r)

	if !ok {
		response.WriteError(
			w,
			http.StatusUnauthorized,
			"UNAUTHORIZED",
			"authentication required",
		)
		return
	}

	projects, err := h.usecase.GetAllAuthorized(
		r.Context(),
		userID,
	)

	if err != nil {
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
		projects,
	)
}

func (h *ProjectHandler) GetByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, ok := getUserID(r)

	if !ok {
		response.WriteError(
			w,
			http.StatusUnauthorized,
			"UNAUTHORIZED",
			"authentication required",
		)
		return
	}

	id, err := strconv.Atoi(
		r.PathValue("id"),
	)

	if err != nil || id <= 0 {
		response.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_ID",
			"invalid project id",
		)
		return
	}

	project, err := h.usecase.GetAuthorized(
		r.Context(),
		userID,
		uint(id),
	)

	if err != nil {
		switch {
		case errors.Is(err, domain.ErrProjectNotFound):
			response.WriteError(
				w,
				http.StatusNotFound,
				"PROJECT_NOT_FOUND",
				"project not found",
			)

		case errors.Is(err, domain.ErrForbidden):
			response.WriteError(
				w,
				http.StatusForbidden,
				"FORBIDDEN",
				"project access denied",
			)

		default:
			response.WriteError(
				w,
				http.StatusInternalServerError,
				"INTERNAL_ERROR",
				"internal server error",
			)
		}

		return
	}

	response.WriteJSON(
		w,
		http.StatusOK,
		project,
	)
}

func (h *ProjectHandler) Update(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, ok := getUserID(r)

	if !ok {
		response.WriteError(
			w,
			http.StatusUnauthorized,
			"UNAUTHORIZED",
			"authentication required",
		)
		return
	}

	id, err := strconv.Atoi(
		r.PathValue("id"),
	)

	if err != nil || id <= 0 {
		response.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_ID",
			"invalid project id",
		)
		return
	}

	var req dto.UpdateProjectRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_JSON",
			"invalid JSON body",
		)
		return
	}

	project := &domain.Project{
		ID:          uint(id),
		Name:        req.Name,
		Description: req.Description,
		Version:     req.Version,
	}

	if err := h.usecase.UpdateAuthorized(
		r.Context(),
		userID,
		project,
	); err != nil {
		switch {
		case errors.Is(err, domain.ErrProjectNotFound):
			response.WriteError(
				w,
				http.StatusNotFound,
				"PROJECT_NOT_FOUND",
				"project not found",
			)

		case errors.Is(err, domain.ErrForbidden):
			response.WriteError(
				w,
				http.StatusForbidden,
				"FORBIDDEN",
				"project modification denied",
			)

		default:
			response.WriteError(
				w,
				http.StatusInternalServerError,
				"INTERNAL_ERROR",
				"internal server error",
			)
		}

		return
	}

	response.WriteJSON(
		w,
		http.StatusOK,
		project,
	)
}

func (h *ProjectHandler) Delete(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, ok := getUserID(r)

	if !ok {
		response.WriteError(
			w,
			http.StatusUnauthorized,
			"UNAUTHORIZED",
			"authentication required",
		)
		return
	}

	id, err := strconv.Atoi(
		r.PathValue("id"),
	)

	if err != nil || id <= 0 {
		response.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_ID",
			"invalid project id",
		)
		return
	}

	if err := h.usecase.DeleteAuthorized(
		r.Context(),
		userID,
		uint(id),
	); err != nil {
		switch {
		case errors.Is(err, domain.ErrProjectNotFound):
			response.WriteError(
				w,
				http.StatusNotFound,
				"PROJECT_NOT_FOUND",
				"project not found",
			)

		case errors.Is(err, domain.ErrForbidden):
			response.WriteError(
				w,
				http.StatusForbidden,
				"FORBIDDEN",
				"project deletion denied",
			)

		default:
			response.WriteError(
				w,
				http.StatusInternalServerError,
				"INTERNAL_ERROR",
				"internal server error",
			)
		}

		return
	}

	w.WriteHeader(http.StatusNoContent)
}
