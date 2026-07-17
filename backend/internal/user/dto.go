package user

import "time"

type UpdateUserRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

type UpdateRoleRequest struct {
	Role Role `json:"role" validate:"required,oneof=user admin"`
}

type UserResponse struct {
	ID               string  `json:"id"`
	Name             string  `json:"name"`
	Email            string  `json:"email"`
	EmailVerified    bool    `json:"emailVerified"`
	Image            *string `json:"image"`
	CreatedAt        string  `json:"createdAt"`
	UpdatedAt        string  `json:"updatedAt"`
	Role             Role    `json:"role"`
	Banned           bool    `json:"banned"`
	BanReason        *string `json:"banReason"`
	BanExpires       *string `json:"banExpires"`
	TwoFactorEnabled bool    `json:"twoFactorEnabled"`
}

func toResponse(u User) UserResponse {
	var banExpires *string
	if u.BanExpires != nil {
		value := formatTime(*u.BanExpires)
		banExpires = &value
	}
	return UserResponse{
		ID:               u.ID,
		Name:             u.Name,
		Email:            u.Email,
		EmailVerified:    u.EmailVerified,
		Image:            u.Image,
		CreatedAt:        formatTime(u.CreatedAt),
		UpdatedAt:        formatTime(u.UpdatedAt),
		Role:             u.Role,
		Banned:           u.Banned,
		BanReason:        u.BanReason,
		BanExpires:       banExpires,
		TwoFactorEnabled: u.TwoFactorEnabled,
	}
}

func toResponses(users []User) []UserResponse {
	resp := make([]UserResponse, len(users))
	for i, u := range users {
		resp[i] = toResponse(u)
	}
	return resp
}

func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}
