package user

import "errors"

var (
	ErrNotFound    = errors.New("user not found")
	ErrEmailExists = errors.New("email already exists")
	ErrSelfRole    = errors.New("cannot change your own role")
)
