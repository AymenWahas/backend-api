package domain

import "errors"

var (
	ErrProjectNotFound  = errors.New("project not found")
	ErrTaskNotFound     = errors.New("task not found")
	ErrEmployeeNotFound = errors.New("employee not found")
)
