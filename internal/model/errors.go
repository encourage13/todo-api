package model

import "errors"

var (
	ErrUserNotFound = errors.New("user not found")
	ErrEmailTaken   = errors.New("not unique email")
	ErrTodoNotFound = errors.New("task not found")
)
