package auth

type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=1,max=100"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=8"`
}

type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required"`
}

type RefreshRequest struct {
	RefreshToken string `json:"refreshToken" validate:"required"`
}

type TokenResponse struct {
	AccessToken  string      `json:"accessToken"`
	RefreshToken string      `json:"refreshToken"`
	User         AuthUser    `json:"user"`
	Session      AuthSession `json:"session"`
}

type SessionResponse struct {
	User    AuthUser    `json:"user"`
	Session AuthSession `json:"session"`
}
