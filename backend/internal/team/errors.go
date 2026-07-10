package team

import "errors"

var (
	ErrNotFound      = errors.New("team not found")
	ErrMemberNotFound = errors.New("team member not found")
	ErrAlreadyMember = errors.New("user is already a member of this team")
)
