package domain

import "errors"

var (
	ErrProjectNotFound           = errors.New("project not found")
	ErrTaskNotFound              = errors.New("task not found")
	ErrEmployeeNotFound          = errors.New("employee not found")
	ErrEmployeeAlreadyExists     = errors.New("employee already exists")
	ErrProjectConflict           = errors.New("project was modified by another transaction")
	ErrNotificationAlreadyExists = errors.New("notification already exists")
)
