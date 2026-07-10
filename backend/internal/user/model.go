package user

import "time"

type Role string

const (
	RoleUser  Role = "user"
	RoleAdmin Role = "admin"
)

type User struct {
	ID               string
	Name             string
	Email            string
	EmailVerified    bool
	Image            *string
	Role             Role
	Banned           bool
	BanReason        *string
	BanExpires       *time.Time
	TwoFactorEnabled bool
	CreatedAt        time.Time
	UpdatedAt        time.Time
}
