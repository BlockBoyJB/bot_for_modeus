package parser

import "errors"

var (
	ErrStudentsNotFound       = errors.New("students not found")
	ErrIncorrectLoginPassword = errors.New("incorrect login or password")
	ErrModeusUnavailable      = errors.New("modeus unavailable")
)
