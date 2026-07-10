package team

type CreateTeamRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

type UpdateTeamRequest struct {
	Name string `json:"name" validate:"required,min=1,max=100"`
}

type AddMemberRequest struct {
	UserID string `json:"userID" validate:"required"`
}
