package auth

import (
	"time"

	"dc-express/internal/user"
)

type AuthUser struct {
	ID               string    `json:"id"`
	Name             string    `json:"name"`
	Email            string    `json:"email"`
	EmailVerified    bool      `json:"emailVerified"`
	Image            *string   `json:"image"`
	CreatedAt        string    `json:"createdAt"`
	UpdatedAt        string    `json:"updatedAt"`
	Role             user.Role `json:"role"`
	Banned           bool      `json:"banned"`
	BanReason        *string   `json:"banReason"`
	BanExpires       *string   `json:"banExpires"`
	TwoFactorEnabled bool      `json:"twoFactorEnabled"`
}

type AuthSession struct {
	ID                     string  `json:"id"`
	ExpiresAt              string  `json:"expiresAt"`
	CreatedAt              string  `json:"createdAt"`
	UpdatedAt              string  `json:"updatedAt"`
	IPAddress              *string `json:"ipAddress"`
	UserAgent              *string `json:"userAgent"`
	UserID                 string  `json:"userId"`
	ImpersonatedBy         *string `json:"impersonatedBy"`
	ActiveOrganizationID   *string `json:"activeOrganizationId"`
	ActiveOrganizationRole *string `json:"activeOrganizationRole"`
	ActiveTeamID           *string `json:"activeTeamId"`
}

type SessionContext struct {
	ID                     string
	ExpiresAt              time.Time
	CreatedAt              time.Time
	UpdatedAt              time.Time
	IPAddress              *string
	UserAgent              *string
	UserID                 string
	ImpersonatedBy         *string
	ActiveOrganizationID   *string
	ActiveOrganizationRole *string
	ActiveTeamID           *string
}
