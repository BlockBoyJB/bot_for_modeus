package modeus

import "errors"

var (
	ErrFindElementTimeout = errors.New("find element timeout")
	ErrIncorrectInputData = errors.New("incorrect input data")

	ErrStudentsNotFound = errors.New("students not found")
)
