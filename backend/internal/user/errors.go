package user

import "errors"

var (
	ErrNotFound      = errors.New("user not found")
	ErrEmailExists   = errors.New("email already exists")
	ErrInvalidID     = errors.New("invalid user id")
)
