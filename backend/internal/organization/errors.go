package organization

import "errors"

var (
	ErrNotFound         = errors.New("organization not found")
	ErrMemberNotFound   = errors.New("member not found")
	ErrSlugExists       = errors.New("slug already exists")
	ErrNotMember        = errors.New("user is not a member of this organization")
	ErrInvitationExpired = errors.New("invitation has expired")
	ErrInvitationNotFound = errors.New("invitation not found")
)
