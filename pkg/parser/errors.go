package parser

import "errors"

var (
	ErrFindElementTimeout = errors.New("finding element timeout")
	ErrIncorrectUserData  = errors.New("incorrect user login or password")
	ErrIncorrectFullName  = errors.New("user not found")
)
