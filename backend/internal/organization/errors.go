package organization

import "errors"

var (
	ErrNotFound           = errors.New("organization not found")
	ErrMemberNotFound     = errors.New("member not found")
	ErrSlugExists         = errors.New("slug already exists")
	ErrNotMember          = errors.New("user is not a member of this organization")
	ErrInvitationExpired  = errors.New("invitation has expired")
	ErrInvitationNotFound = errors.New("invitation not found")
	ErrForbidden          = errors.New("forbidden")
	ErrOwnerProtected     = errors.New("organization owner cannot be changed or removed")
	ErrLogoTooLarge       = errors.New("logo must be 2 MB or smaller")
	ErrInvalidLogo        = errors.New("logo must be a valid JPEG or PNG up to 2048 by 2048 pixels")
	ErrLogoUnavailable    = errors.New("logo storage is unavailable")
	ErrNoLogo             = errors.New("logo not found")
)
