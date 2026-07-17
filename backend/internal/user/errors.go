package user

import "errors"

var (
	ErrNotFound          = errors.New("user not found")
	ErrEmailExists       = errors.New("email already exists")
	ErrSelfRole          = errors.New("cannot change your own role")
	ErrAvatarTooLarge    = errors.New("avatar must be 2 MB or smaller")
	ErrInvalidAvatar     = errors.New("avatar must be a valid JPEG or PNG up to 2048 by 2048 pixels")
	ErrAvatarUnavailable = errors.New("avatar storage is unavailable")
	ErrNoAvatar          = errors.New("avatar not found")
)
