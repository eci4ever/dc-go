package academic

import "errors"

var (
	ErrForbidden = errors.New("forbidden")
	ErrNotFound  = errors.New("academic record not found")
	ErrConflict  = errors.New("academic record already exists")
)
