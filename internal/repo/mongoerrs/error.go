package mongoerrs

import "errors"

var (
	ErrUserCannotCreate = errors.New("cannot create user")
	ErrNotFound         = errors.New("not found")
)
