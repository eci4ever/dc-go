package team

type Team struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	OrgID     string `json:"org_id"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

type TeamMember struct {
	ID        string `json:"id"`
	TeamID    string `json:"team_id"`
	UserID    string `json:"user_id"`
	CreatedAt string `json:"created_at"`
	Name      string `json:"name"`
	Email     string `json:"email"`
}

const time3339 = "2006-01-02T15:04:05Z07:00"
