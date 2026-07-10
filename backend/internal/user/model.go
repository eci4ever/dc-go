package user

import "time"

type User struct {
	ID            string    `json:"id"`
	Name          string    `json:"name"`
	Email         string    `json:"email"`
	EmailVerified bool      `json:"email_verified"`
	Image         *string   `json:"image,omitempty"`
	Role          *string   `json:"role,omitempty"`
	Banned        bool      `json:"banned"`
	CreatedAt     time.Time `json:"created_at"`
	UpdatedAt     time.Time `json:"updated_at"`
}
