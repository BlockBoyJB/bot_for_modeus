package modeus

import "errors"

var (
	ErrIncorrectInputData = errors.New("incorrect input data")
	ErrStudentsNotFound   = errors.New("students not found")
	ErrModeusUnavailable  = errors.New("modeus unavailable")
)
