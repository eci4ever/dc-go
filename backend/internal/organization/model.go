package organization

import "time"

const (
	StatusActive    = "active"
	StatusInactive  = "inactive"
	StatusSuspended = "suspended"
	StatusArchived  = "archived"

	PermissionMembersManage           = "members.manage"
	PermissionAcademicStudentsManage  = "academic.students.manage"
	PermissionAcademicStructureManage = "academic.structure.manage"
	PermissionAcademicResultsManage   = "academic.results.manage"
	PermissionAuditView               = "audit.view"
)

type Organization struct {
	ID              string             `json:"id"`
	Name            string             `json:"name"`
	Slug            string             `json:"slug"`
	Logo            *string            `json:"logo,omitempty"`
	CreatedAt       string             `json:"created_at"`
	LogoKey         *string            `json:"-"`
	LogoContentType *string            `json:"-"`
	LogoUpdatedAt   *time.Time         `json:"-"`
	Owner           *OrganizationOwner `json:"owner,omitempty"`
	MembershipRole  *string            `json:"membership_role,omitempty"`
	Status          string             `json:"status"`
}

type OrganizationOwner struct {
	ID    string  `json:"id"`
	Name  string  `json:"name"`
	Email string  `json:"email"`
	Image *string `json:"image,omitempty"`
}

type Member struct {
	ID          string   `json:"id"`
	OrgID       string   `json:"org_id"`
	UserID      string   `json:"user_id"`
	Role        string   `json:"role"`
	CreatedAt   string   `json:"created_at"`
	Permissions []string `json:"permissions"`
	User        struct {
		Name  string  `json:"name"`
		Email string  `json:"email"`
		Image *string `json:"image,omitempty"`
	} `json:"user"`
}

type AuditLog struct {
	ID             string         `json:"id"`
	OrganizationID string         `json:"organization_id"`
	ActorUserID    *string        `json:"actor_user_id,omitempty"`
	ActorName      string         `json:"actor_name"`
	ActorEmail     string         `json:"actor_email"`
	Action         string         `json:"action"`
	TargetType     string         `json:"target_type"`
	TargetID       *string        `json:"target_id,omitempty"`
	Details        map[string]any `json:"details"`
	CreatedAt      string         `json:"created_at"`
}

type Invitation struct {
	ID        string  `json:"id"`
	OrgID     string  `json:"org_id"`
	Email     string  `json:"email"`
	Role      string  `json:"role"`
	Status    string  `json:"status"`
	InviterID string  `json:"inviter_id"`
	ExpiresAt string  `json:"expires_at"`
	CreatedAt string  `json:"created_at"`
	TeamID    *string `json:"team_id,omitempty"`
}
