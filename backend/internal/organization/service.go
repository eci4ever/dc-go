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

func (s *Service) GetByID(ctx context.Context, id, actorID string) (Organization, error) {
	if ok, err := s.repo.IsMember(ctx, id, actorID); err != nil {
		return Organization{}, err
	} else if !ok {
		return Organization{}, ErrNotFound
	}
	return s.repo.GetByID(ctx, id)
}

func (s *Service) List(ctx context.Context, userID string) ([]Organization, error) {
	return s.repo.ListByUserID(ctx, userID)
}

func (s *Service) Update(ctx context.Context, id, actorID string, req UpdateOrgRequest) (Organization, error) {
	if err := s.requireRole(ctx, id, actorID, "owner"); err != nil {
		return Organization{}, err
	}
	return s.repo.Update(ctx, id, req.Name, req.Slug, req.Logo)
}

func (s *Service) Delete(ctx context.Context, id, actorID string) error {
	if err := s.requireRole(ctx, id, actorID, "owner"); err != nil {
		return err
	}
	return s.repo.Delete(ctx, id)
}

func (s *Service) GetMembers(ctx context.Context, orgID, actorID string) ([]Member, error) {
	if err := s.requireMember(ctx, orgID, actorID); err != nil {
		return nil, err
	}
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

func (s *Service) UpdateMemberRole(ctx context.Context, orgID, userID, actorID, role string) error {
	if err := s.requireRole(ctx, orgID, actorID, "owner"); err != nil {
		return err
	}
	return s.repo.UpdateMemberRole(ctx, orgID, userID, role)
}

func (s *Service) RemoveMember(ctx context.Context, orgID, userID, actorID string) error {
	if actorID != userID {
		if err := s.requireRole(ctx, orgID, actorID, "owner"); err != nil {
			return err
		}
	}
	return s.repo.RemoveMember(ctx, orgID, userID)
}

func (s *Service) Invite(ctx context.Context, orgID, email, role, inviterID string) (Invitation, error) {
	if err := s.requireManager(ctx, orgID, inviterID); err != nil {
		return Invitation{}, err
	}
	return s.repo.CreateInvitation(ctx, orgID, email, role, inviterID, time.Now().Add(invitationTTL))
}

func (s *Service) ListInvitations(ctx context.Context, orgID, actorID string) ([]Invitation, error) {
	if err := s.requireMember(ctx, orgID, actorID); err != nil {
		return nil, err
	}
	return s.repo.ListInvitationsByOrgID(ctx, orgID)
}

func (s *Service) AcceptInvitation(ctx context.Context, id, userID string) error {
	return s.repo.AcceptInvitation(ctx, id, userID)
}

func (s *Service) DeclineInvitation(ctx context.Context, id, userID string) error {
	return s.repo.DeclineInvitation(ctx, id, userID)
}

func (s *Service) CancelInvitation(ctx context.Context, id, actorID string) error {
	inv, err := s.repo.GetInvitation(ctx, id)
	if err != nil {
		return err
	}
	if err := s.requireManager(ctx, inv.OrgID, actorID); err != nil {
		return err
	}
	return s.repo.CancelInvitation(ctx, id)
}

func (s *Service) requireMember(ctx context.Context, orgID, actorID string) error {
	ok, err := s.repo.IsMember(ctx, orgID, actorID)
	if err != nil {
		return err
	}
	if !ok {
		return ErrForbidden
	}
	return nil
}
func (s *Service) requireRole(ctx context.Context, orgID, actorID, role string) error {
	m, err := s.repo.GetMember(ctx, orgID, actorID)
	if err != nil {
		return ErrForbidden
	}
	if m.Role != role {
		return ErrForbidden
	}
	return nil
}
func (s *Service) requireManager(ctx context.Context, orgID, actorID string) error {
	m, err := s.repo.GetMember(ctx, orgID, actorID)
	if err != nil {
		return ErrForbidden
	}
	if m.Role != "owner" && m.Role != "admin" {
		return ErrForbidden
	}
	return nil
}
