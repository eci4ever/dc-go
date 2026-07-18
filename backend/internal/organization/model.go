package organization

import "time"

type Organization struct {
	ID              string     `json:"id"`
	Name            string     `json:"name"`
	Slug            string     `json:"slug"`
	Logo            *string    `json:"logo,omitempty"`
	CreatedAt       string     `json:"created_at"`
	LogoKey         *string    `json:"-"`
	LogoContentType *string    `json:"-"`
	LogoUpdatedAt   *time.Time `json:"-"`
}

type Member struct {
	ID        string `json:"id"`
	OrgID     string `json:"org_id"`
	UserID    string `json:"user_id"`
	Role      string `json:"role"`
	CreatedAt string `json:"created_at"`
	User      struct {
		Name  string  `json:"name"`
		Email string  `json:"email"`
		Image *string `json:"image,omitempty"`
	} `json:"user"`
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
