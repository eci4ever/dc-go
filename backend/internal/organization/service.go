package organization

import (
	"context"
	"errors"
	"time"
)

const invitationTTL = 7 * 24 * time.Hour

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, req CreateOrgRequest, creatorID string) (Organization, error) {
	return s.repo.Create(ctx, req.Name, req.Slug, req.Logo, creatorID)
}

func (s *Service) GetByID(ctx context.Context, id string) (Organization, error) {
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context, userID string) ([]Organization, error) {
	return s.repo.ListByUserID(ctx, userID)
}

func (s *Service) Update(ctx context.Context, id string, req UpdateOrgRequest) (Organization, error) {
	return s.repo.Update(ctx, id, req.Name, req.Slug, req.Logo)
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) GetMembers(ctx context.Context, orgID string) ([]Member, error) {
	return s.repo.GetMembers(ctx, orgID)
}

func (s *Service) GetMember(ctx context.Context, orgID, userID string) (Member, error) {
	m, err := s.repo.GetMember(ctx, orgID, userID)
	if err != nil {
		if errors.Is(err, ErrMemberNotFound) {
			return Member{}, ErrNotMember
		}
		return Member{}, err
	}
	return m, nil
}

func (s *Service) UpdateMemberRole(ctx context.Context, orgID, userID, role string) error {
	return s.repo.UpdateMemberRole(ctx, orgID, userID, role)
}

func (s *Service) RemoveMember(ctx context.Context, orgID, userID string) error {
	return s.repo.RemoveMember(ctx, orgID, userID)
}

func (s *Service) Invite(ctx context.Context, orgID, email, role, inviterID string) (Invitation, error) {
	return s.repo.CreateInvitation(ctx, orgID, email, role, inviterID, time.Now().Add(invitationTTL))
}

func (s *Service) ListInvitations(ctx context.Context, orgID string) ([]Invitation, error) {
	return s.repo.ListInvitationsByOrgID(ctx, orgID)
}

func (s *Service) AcceptInvitation(ctx context.Context, id, userID string) error {
	return s.repo.AcceptInvitation(ctx, id, userID)
}

func (s *Service) DeclineInvitation(ctx context.Context, id string) error {
	return s.repo.DeclineInvitation(ctx, id)
}

func (s *Service) CancelInvitation(ctx context.Context, id string) error {
	return s.repo.CancelInvitation(ctx, id)
}
