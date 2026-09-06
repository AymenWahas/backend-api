package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"backend-api/internal/delivery/http/dto"
	"backend-api/internal/delivery/http/response"
	"backend-api/internal/domain"
	"backend-api/internal/repository"
	"backend-api/internal/usecase"
)

type Handler struct {
	usecase        *usecase.EmployeeUsecase
	projectHandler *ProjectHandler
	taskHandler    *TaskHandler
}

func NewHandler(
	employeeUC *usecase.EmployeeUsecase,
	projectUC *usecase.ProjectUsecase,
	taskUC *usecase.TaskUsecase,
) *Handler {
	return &Handler{
		usecase:        employeeUC,
		projectHandler: NewProjectHandler(projectUC),
		taskHandler:    NewTaskHandler(taskUC),
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
	//dto to domain
	employee := domain.Employee{
		Name:       req.Name,
		Email:      req.Email,
		Department: req.Department,
	}
	//call usecase to create employee
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

		response.WriteError(w, http.StatusInternalServerError,
			"INTERNAL_ERROR",
			"internal server error",
		)
		return
	}
	// domain to dto
	res := dto.EmployeeResponse{
		ID:         created.ID,
		Name:       created.Name,
		Email:      created.Email,
		Department: req.Department,
	}
	//send response error if any other error occurs
	response.WriteJSON(w, http.StatusCreated, res)
}

func (h *Handler) GetEmployees(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query()
	filter := repository.EmployeeFilter{
		Name:  query.Get("name"),
		Email: query.Get("email"),
	}
	//sort by id, name, email
	sortParam := query.Get("sort")

	sortBy := repository.EmployeeSort{
		Field: "id",
	}

	if strings.HasPrefix(sortParam, "-") {
		sortBy.Desc = true
		sortBy.Field = strings.TrimPrefix(sortParam, "-")
	} else if sortParam != "" {
		sortBy.Field = sortParam
	}
	if sortBy.Field != "id" &&
		sortBy.Field != "name" &&
		sortBy.Field != "email" {

		response.WriteError(
			w,
			http.StatusBadRequest,
			"INVALID_SORT",
			"sort must be id, name, or email",
		)
		return
	}
	//pagination
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

	//call usecase to get employees with filter, sort and pagination
	employees, total, version, err := h.usecase.GetEmployees(
		filter,
		sortBy,
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
	//set etag header for caching
	etag := fmt.Sprintf(
		`"employees-%d-page-%d-size-%d-name-%s-email-%s-sort-%s-desc-%t"`,
		version,
		page,
		pageSize,
		filter.Name,
		filter.Email,
		sortBy.Field,
		sortBy.Desc,
	)
	w.Header().Set("ETag", etag)
	//check if the etag matches the request header, if so return 304 Not Modified
	if r.Header.Get("If-None-Match") == etag {
		w.WriteHeader(http.StatusNotModified)
		return
	}
	//convert domain employees to dto employees
	data := make([]dto.EmployeeResponse, 0, len(employees))

	for _, employee := range employees {
		data = append(data, dto.EmployeeResponse{
			ID:         employee.ID,
			Name:       employee.Name,
			Email:      employee.Email,
			Department: employee.Department,
		})
	}
	//calculate total pages for pagination
	totalPages := 0
	if total > 0 {
		totalPages = (total + pageSize - 1) / pageSize
	}
	//create response with data and pagination info
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

// GetEmployee handles the GET /employees/{id} endpoint.
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
		ID:         id,
		Name:       req.Name,
		Email:      req.Email,
		Department: req.Department,
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
		ID:         updated.ID,
		Name:       updated.Name,
		Email:      updated.Email,
		Department: updated.Department,
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
func (h *Handler) CreateProject(w http.ResponseWriter, r *http.Request) {
	h.projectHandler.Create(w, r)
}

func (h *Handler) GetProjects(w http.ResponseWriter, r *http.Request) {
	h.projectHandler.GetAll(w, r)
}

func (h *Handler) GetProject(w http.ResponseWriter, r *http.Request) {
	h.projectHandler.GetByID(w, r)
}

func (h *Handler) UpdateProject(w http.ResponseWriter, r *http.Request) {
	h.projectHandler.Update(w, r)
}

func (h *Handler) DeleteProject(w http.ResponseWriter, r *http.Request) {
	h.projectHandler.Delete(w, r)
}

func (h *Handler) CreateTask(w http.ResponseWriter, r *http.Request) {
	h.taskHandler.Create(w, r)
}
func (h *Handler) GetTasks(w http.ResponseWriter, r *http.Request) {
	h.taskHandler.GetAll(w, r)
}
func (h *Handler) GetTask(w http.ResponseWriter, r *http.Request) {
	h.taskHandler.GetByID(w, r)
}
func (h *Handler) UpdateTask(w http.ResponseWriter, r *http.Request) {
	h.taskHandler.Update(w, r)
}
func (h *Handler) DeleteTask(w http.ResponseWriter, r *http.Request) {
	h.taskHandler.Delete(w, r)
}
