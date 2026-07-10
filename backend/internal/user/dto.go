package user

type UpdateUserRequest struct {
	Name  string `json:"name" validate:"required,min=1,max=100"`
	Email string `json:"email" validate:"required,email,max=100"`
	Image *string `json:"image,omitempty"`
}

type UserResponse struct {
	ID            string  `json:"id"`
	Name          string  `json:"name"`
	Email         string  `json:"email"`
	EmailVerified bool    `json:"email_verified"`
	Image         *string `json:"image,omitempty"`
	Role          *string `json:"role,omitempty"`
	CreatedAt     string  `json:"created_at"`
	UpdatedAt     string  `json:"updated_at"`
}

func toResponse(u User) UserResponse {
	return UserResponse{
		ID:            u.ID,
		Name:          u.Name,
		Email:         u.Email,
		EmailVerified: u.EmailVerified,
		Image:         u.Image,
		Role:          u.Role,
		CreatedAt:     u.CreatedAt.Format(time3339),
		UpdatedAt:     u.UpdatedAt.Format(time3339),
	}
}

func toResponses(users []User) []UserResponse {
	resp := make([]UserResponse, len(users))
	for i, u := range users {
		resp[i] = toResponse(u)
	}
	return resp
}

const time3339 = "2006-01-02T15:04:05Z07:00"
