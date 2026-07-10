package organization

import "time"

type CreateOrgRequest struct {
	Name string  `json:"name" validate:"required,min=1,max=100"`
	Slug string  `json:"slug" validate:"required,min=1,max=50"`
	Logo *string `json:"logo,omitempty"`
}

type UpdateOrgRequest struct {
	Name string  `json:"name" validate:"required,min=1,max=100"`
	Slug string  `json:"slug" validate:"required,min=1,max=50"`
	Logo *string `json:"logo,omitempty"`
}

type InviteRequest struct {
	Email string `json:"email" validate:"required,email"`
	Role  string `json:"role" validate:"required,oneof=owner admin member"`
}

type UpdateMemberRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=owner admin member"`
}

const time3339 = "2006-01-02T15:04:05Z07:00"

func formatTime(t time.Time) string {
	return t.Format(time3339)
}
