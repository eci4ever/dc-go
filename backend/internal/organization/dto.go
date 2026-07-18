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

type SetOwnerRequest struct {
	UserID string `json:"user_id" validate:"required,uuid"`
}

type UpdateStatusRequest struct {
	Status string `json:"status" validate:"required,oneof=active inactive suspended archived"`
}

type UpdateMemberPermissionsRequest struct {
	Permissions []string `json:"permissions" validate:"max=5,dive,oneof=members.manage academic.students.manage academic.structure.manage academic.results.manage audit.view"`
}

type InviteRequest struct {
	Email string `json:"email" validate:"required,email"`
	Role  string `json:"role" validate:"required,oneof=admin member"`
}

type UpdateMemberRoleRequest struct {
	Role string `json:"role" validate:"required,oneof=admin member"`
}

const time3339 = "2006-01-02T15:04:05Z07:00"

func formatTime(t time.Time) string {
	return t.Format(time3339)
}
