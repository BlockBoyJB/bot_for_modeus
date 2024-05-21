package service

import "errors"

var (
	ErrUserAlreadyExists     = errors.New("user already exists")
	ErrUserCannotCreate      = errors.New("cannot create user")
	ErrUserCannotUpdate      = errors.New("cannot update user info")
	ErrUserNotFound          = errors.New("user not found")
	ErrUserIncorrectFullName = errors.New("incorrect user full name")
	ErrUserPermissionDenied  = errors.New("user does not have permission for action")
)
