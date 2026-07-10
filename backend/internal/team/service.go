package team

import "context"

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, orgID, actorID string, req CreateTeamRequest) (Team, error) {
	if err := s.manager(ctx, orgID, actorID); err != nil {
		return Team{}, err
	}
	return s.repo.Create(ctx, orgID, req.Name)
}

func (s *Service) GetByID(ctx context.Context, id, actorID string) (Team, error) {
	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return Team{}, err
	}
	if err := s.member(ctx, t.OrgID, actorID); err != nil {
		return Team{}, ErrNotFound
	}
	return t, nil
}

func (s *Service) List(ctx context.Context, orgID, actorID string) ([]Team, error) {
	if err := s.member(ctx, orgID, actorID); err != nil {
		return nil, err
	}
	return s.repo.ListByOrgID(ctx, orgID)
}

func (s *Service) Update(ctx context.Context, id, actorID string, req UpdateTeamRequest) (Team, error) {
	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return Team{}, err
	}
	if err := s.manager(ctx, t.OrgID, actorID); err != nil {
		return Team{}, err
	}
	return s.repo.Update(ctx, id, req.Name)
}

func (s *Service) Delete(ctx context.Context, id, actorID string) error {
	t, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := s.manager(ctx, t.OrgID, actorID); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) AddMember(ctx context.Context, teamID, actorID string, req AddMemberRequest) (TeamMember, error) {
	t, err := s.repo.GetByID(ctx, teamID)
	if err != nil {
		return TeamMember{}, err
	}
	if err := s.manager(ctx, t.OrgID, actorID); err != nil {
		return TeamMember{}, err
	}
	if ok, err := s.repo.CheckOrgMembership(ctx, t.OrgID, req.UserID); err != nil || !ok {
		return TeamMember{}, ErrForbidden
	}
	return s.repo.AddMember(ctx, teamID, req.UserID)
}

func (s *Service) GetMembers(ctx context.Context, teamID, actorID string) ([]TeamMember, error) {
	t, err := s.repo.GetByID(ctx, teamID)
	if err != nil {
		return nil, err
	}
	if err := s.member(ctx, t.OrgID, actorID); err != nil {
		return nil, err
	}
	return s.repo.GetMembers(ctx, teamID)
}

func (s *Service) RemoveMember(ctx context.Context, teamID, userID, actorID string) error {
	t, err := s.repo.GetByID(ctx, teamID)
	if err != nil {
		return err
	}
	if err := s.manager(ctx, t.OrgID, actorID); err != nil {
		return err
	}
	return s.repo.RemoveMember(ctx, teamID, userID)
}

func (s *Service) member(ctx context.Context, orgID, actorID string) error {
	role, err := s.repo.MemberRole(ctx, orgID, actorID)
	if err != nil {
		return err
	}
	if role == "" {
		return ErrForbidden
	}
	return nil
}
func (s *Service) manager(ctx context.Context, orgID, actorID string) error {
	role, err := s.repo.MemberRole(ctx, orgID, actorID)
	if err != nil {
		return err
	}
	if role != "owner" && role != "admin" {
		return ErrForbidden
	}
	return nil
}
