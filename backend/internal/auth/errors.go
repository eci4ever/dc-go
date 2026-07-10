package auth

import "errors"

var (
	ErrInvalidCredentials = errors.New("invalid credentials")
	ErrEmailExists        = errors.New("email already exists")
	ErrInvalidToken       = errors.New("invalid token")
	ErrExpiredToken       = errors.New("expired token")
	ErrUserBanned         = errors.New("user is banned")
	ErrForbidden          = errors.New("forbidden")
	ErrIncorrectPassword  = errors.New("current password is incorrect")
	ErrCurrentSession     = errors.New("current session cannot be revoked")
	ErrSessionNotFound    = errors.New("session not found")
)
