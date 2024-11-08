package modeus

import (
	"errors"
	"fmt"
)

var (
	ErrIncorrectInputData = errors.New("incorrect input data")
	ErrStudentsNotFound   = errors.New("students not found")
)

type ErrModeusUnavailable struct {
	StatusCode int    `json:"statusCode"`
	Message    string `json:"message"`
	Timestamp  string `json:"timestamp"`
}

func (e *ErrModeusUnavailable) Error() string {
	return fmt.Sprintf("modeus unavailable: code: %d, timestamp: %s, message: %s", e.StatusCode, e.Timestamp, e.Message)
}
