package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"

	"backend-api/internal/delivery/http/dto"
	"backend-api/internal/delivery/http/response"
	"backend-api/internal/domain"
	"backend-api/internal/repository"
	"backend-api/internal/usecase"
)

type Handler struct {
	usecase *usecase.EmployeeUsecase
}

func NewHandler(uc *usecase.EmployeeUsecase) *Handler {
	return &Handler{
		usecase: uc,
	}
}

func (h *Handler) Health(w http.ResponseWriter, r *http.Request) {
	response.WriteJSON(w, http.StatusOK, map[string]string{
		"status": "ok",
	})
}

func (h *Handler) CreateEmployee(w http.ResponseWriter, r *http.Request) {
	var req dto.CreateEmployeeRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_JSON",
			"invalid JSON body",
		)
		return
	}

	if !req.Validate() {
		response.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_EMPLOYEE",
			"name and email are required",
		)
		return
	}

	employee := domain.Employee{
		Name:  req.Name,
		Email: req.Email,
	}

	created, err := h.usecase.CreateEmployee(employee)
	if err != nil {
		if errors.Is(err, usecase.ErrInvalidEmployee) {
			response.WriteError(
				w,
				http.StatusBadRequest,
				"INVALID_EMPLOYEE",
				"employee name and email are required",
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

	res := dto.EmployeeResponse{
		ID:    created.ID,
		Name:  created.Name,
		Email: created.Email,
	}

	response.WriteJSON(w, http.StatusCreated, res)
}

func (h *Handler) GetEmployees(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()

	page := 1
	pageSize := 10

	if value := query.Get("page"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 {
			response.WriteError(
				w,
				http.StatusBadRequest,
				"INVALID_PAGE",
				"page must be greater than 0",
			)
			return
		}

		page = parsed
	}

	if value := query.Get("page_size"); value != "" {
		parsed, err := strconv.Atoi(value)
		if err != nil || parsed < 1 || parsed > 100 {
			response.WriteError(
				w,
				http.StatusBadRequest,
				"INVALID_PAGE_SIZE",
				"page_size must be between 1 and 100",
			)
			return
		}

		pageSize = parsed
	}

	offset := (page - 1) * pageSize

	employees, total, version, err := h.usecase.GetEmployees(
		offset,
		pageSize,
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

	etag := fmt.Sprintf(
		`"employees-%d-page-%d-size-%d"`,
		version,
		page,
		pageSize,
	)

	w.Header().Set("ETag", etag)

	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}

	data := make([]dto.EmployeeResponse, 0, len(employees))

	for _, employee := range employees {
		data = append(data, dto.EmployeeResponse{
			ID:    employee.ID,
			Name:  employee.Name,
			Email: employee.Email,
		})
	}

	totalPages := 0
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}

	res := dto.EmployeeListResponse{
		Data: data,
		Pagination: dto.Pagination{
			Page:       page,
			PageSize:   pageSize,
			Total:      total,
			TotalPages: totalPages,
		},
	}

	response.WriteJSON(w, http.StatusOK, res)
}

func (h *Handler) GetEmployee(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		response.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_ID",
			"invalid employee id",
		)
		return
	}
	employee, err := h.usecase.GetEmployee(id)
	if err != nil {
		if errors.Is(err, repository.ErrEmployeeNotFound) {
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

	response.WriteJSON(w, http.StatusOK, employee)
}

func (h *Handler) UpdateEmployee(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		response.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_ID",
			"invalid employee id",
		)
		return
	}

	var req dto.UpdateEmployeeRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		response.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_JSON",
			"invalid JSON body",
		)
		return
	}

	employee := domain.Employee{
		ID:    id,
		Name:  req.Name,
		Email: req.Email,
	}

	updated, err := h.usecase.UpdateEmployee(employee)
	if err != nil {
		switch {
		case errors.Is(err, usecase.ErrInvalidEmployee):
			response.WriteError(
				w,
				http.StatusBadRequest,
				"INVALID_EMPLOYEE",
				"employee name and email are required",
			)

		case errors.Is(err, repository.ErrEmployeeNotFound):
			response.WriteError(
				w,
				http.StatusNotFound,
				"EMPLOYEE_NOT_FOUND",
				"employee not found",
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

	res := dto.EmployeeResponse{
		ID:    updated.ID,
		Name:  updated.Name,
		Email: updated.Email,
	}

	response.WriteJSON(w, http.StatusOK, res)
}

func (h *Handler) DeleteEmployee(w http.ResponseWriter, r *http.Request) {
	id, err := strconv.Atoi(r.PathValue("id"))
	if err != nil || id <= 0 {
		response.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_ID",
			"invalid employee id",
		)
		return
	}

	err = h.usecase.DeleteEmployee(id)
	if err != nil {
		if errors.Is(err, repository.ErrEmployeeNotFound) {
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

	w.WriteHeader(http.StatusNoContent)
}
