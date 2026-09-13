package domain

import "errors"

var (
	ErrProjectNotFound           = errors.New("project not found")
	ErrTaskNotFound              = errors.New("task not found")
	ErrEmployeeNotFound          = errors.New("employee not found")
	ErrSessionNotFound           = errors.New("session not found")
	ErrEmployeeAlreadyExists     = errors.New("employee already exists")
	ErrUserNotFound              = errors.New("user not found")
	ErrUserAlreadyExists         = errors.New("user already exists")
	ErrInvalidCredentials        = errors.New("invalid credentials")
	ErrProjectConflict           = errors.New("project was modified by another transaction")
	ErrNotificationAlreadyExists = errors.New("notification already exists")
)
