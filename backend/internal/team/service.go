package team

import "context"

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, orgID string, req CreateTeamRequest) (Team, error) {
	return s.repo.Create(ctx, orgID, req.Name)
}

func (s *Service) GetByID(ctx context.Context, id string) (Team, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context, orgID string) ([]Team, error) {
	return s.repo.ListByOrgID(ctx, orgID)
}

func (s *Service) Update(ctx context.Context, id string, req UpdateTeamRequest) (Team, error) {
	return s.repo.Update(ctx, id, req.Name)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) AddMember(ctx context.Context, teamID string, req AddMemberRequest) (TeamMember, error) {
	return s.repo.AddMember(ctx, teamID, req.UserID)
}

func (s *Service) GetMembers(ctx context.Context, teamID string) ([]TeamMember, error) {
	return s.repo.GetMembers(ctx, teamID)
}

func (s *Service) RemoveMember(ctx context.Context, teamID, userID string) error {
	return s.repo.RemoveMember(ctx, teamID, userID)
}
