package service

import "errors"

var (
	ErrUserAlreadyExists  = errors.New("user already exists")
	ErrUserNotFound       = errors.New("user not found")
	ErrUserIncorrectLogin = errors.New("user incorrect login input")
)
