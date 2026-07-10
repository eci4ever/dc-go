package auth

type AuthUser struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Email         string  `json:"email"`
	EmailVerified bool    `json:"email_verified"`
	Image         *string `json:"image,omitempty"`
	Role          *string `json:"role,omitempty"`
	Banned        bool    `json:"banned"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

type AuthSession struct {
	ID        string `json:"id"`
	ExpiresAt string `json:"expires_at"`
	CreatedAt string `json:"created_at"`
	UserID    string `json:"user_id"`
}
