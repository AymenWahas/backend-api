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

type TaskHandler struct {
	usecase *usecase.TaskUsecase
}

func NewTaskHandler(
	uc *usecase.TaskUsecase,
) *TaskHandler {
	return &TaskHandler{
		usecase: uc,
	}
}

func (h *TaskHandler) Create(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, ok := r.Context().
		Value(middleware.UserIDContextKey).(int)

	if !ok {
		response.WriteError(
			w,
			http.StatusUnauthorized,
			"UNAUTHORIZED",
			"authentication required",
		)
		return
	}

	var req dto.CreateTaskRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_JSON",
			"invalid JSON body",
		)
		return
	}

	if req.Title == "" {
		response.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_TASK",
			"task title is required",
		)
		return
	}

	if req.ProjectID == 0 {
		response.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_PROJECT_ID",
			"project id is required",
		)
		return
	}

	task := &domain.Task{
		ProjectID: req.ProjectID,
		Title:     req.Title,
		Status:    req.Status,
	}

	if err := h.usecase.Create(
		r.Context(),
		userID,
		task,
	); err != nil {
		if errors.Is(err, domain.ErrForbidden) {
			response.WriteError(
				w,
				http.StatusForbidden,
				"FORBIDDEN",
				"task creation denied",
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
		task,
	)
}

func (h *TaskHandler) GetAll(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, ok := r.Context().
		Value(middleware.UserIDContextKey).(int)

	if !ok {
		response.WriteError(
			w,
			http.StatusUnauthorized,
			"UNAUTHORIZED",
			"authentication required",
		)
		return
	}

	tasks, err := h.usecase.GetAll(
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
		tasks,
	)
}

func (h *TaskHandler) GetByID(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, ok := r.Context().
		Value(middleware.UserIDContextKey).(int)

	if !ok {
		response.WriteError(
			w,
			http.StatusUnauthorized,
			"UNAUTHORIZED",
			"authentication required",
		)
		return
	}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		response.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_ID",
			"invalid task id",
		)
		return
	}

	task, err := h.usecase.GetByID(
		r.Context(),
		userID,
		uint(id),
	)
	if err != nil {
		if errors.Is(err, domain.ErrTaskNotFound) {
			response.WriteError(
				w,
				http.StatusNotFound,
				"TASK_NOT_FOUND",
				"task not found",
			)
			return
		}

		if errors.Is(err, domain.ErrForbidden) {
			response.WriteError(
				w,
				http.StatusForbidden,
				"FORBIDDEN",
				"task access denied",
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
		task,
	)
}

func (h *TaskHandler) Update(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, ok := r.Context().
		Value(middleware.UserIDContextKey).(int)

	if !ok {
		response.WriteError(
			w,
			http.StatusUnauthorized,
			"UNAUTHORIZED",
			"authentication required",
		)
		return
	}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		response.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_ID",
			"invalid task id",
		)
		return
	}

	task, err := h.usecase.GetByID(
		r.Context(),
		userID,
		uint(id),
	)
	if err != nil {
		if errors.Is(err, domain.ErrTaskNotFound) {
			response.WriteError(
				w,
				http.StatusNotFound,
				"TASK_NOT_FOUND",
				"task not found",
			)
			return
		}

		if errors.Is(err, domain.ErrForbidden) {
			response.WriteError(
				w,
				http.StatusForbidden,
				"FORBIDDEN",
				"task access denied",
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

	var req dto.UpdateTaskRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_JSON",
			"invalid JSON body",
		)
		return
	}

	if req.Title == "" {
		response.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_TASK",
			"task title is required",
		)
		return
	}

	task.Title = req.Title
	task.Status = req.Status

	if err := h.usecase.Update(
		r.Context(),
		userID,
		task,
	); err != nil {
		if errors.Is(err, domain.ErrTaskNotFound) {
			response.WriteError(
				w,
				http.StatusNotFound,
				"TASK_NOT_FOUND",
				"task not found",
			)
			return
		}

		if errors.Is(err, domain.ErrForbidden) {
			response.WriteError(
				w,
				http.StatusForbidden,
				"FORBIDDEN",
				"task modification denied",
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
		task,
	)
}

func (h *TaskHandler) Delete(
	w http.ResponseWriter,
	r *http.Request,
) {
	userID, ok := r.Context().
		Value(middleware.UserIDContextKey).(int)

	if !ok {
		response.WriteError(
			w,
			http.StatusUnauthorized,
			"UNAUTHORIZED",
			"authentication required",
		)
		return
	}

	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		response.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_ID",
			"invalid task id",
		)
		return
	}

	if err := h.usecase.Delete(
		r.Context(),
		userID,
		uint(id),
	); err != nil {
		if errors.Is(err, domain.ErrTaskNotFound) {
			response.WriteError(
				w,
				http.StatusNotFound,
				"TASK_NOT_FOUND",
				"task not found",
			)
			return
		}

		if errors.Is(err, domain.ErrForbidden) {
			response.WriteError(
				w,
				http.StatusForbidden,
				"FORBIDDEN",
				"task deletion denied",
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
