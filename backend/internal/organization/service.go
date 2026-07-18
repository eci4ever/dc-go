package organization

import (
	"context"
	"errors"
	"io"
	"log/slog"
	"time"

	"github.com/eci4ever/dc-go/internal/storage"
)

const invitationTTL = 7 * 24 * time.Hour

type Service struct {
	repo      *Repository
	logoStore storage.ObjectStore
}

func NewService(repo *Repository, logoStore storage.ObjectStore) *Service {
	return &Service{repo: repo, logoStore: logoStore}
}

func (s *Service) Create(ctx context.Context, req CreateOrgRequest, creatorID string) (Organization, error) {
	org, err := s.repo.Create(ctx, req.Name, req.Slug, req.Logo, creatorID)
	if err == nil {
		s.recordAudit(ctx, org.ID, creatorID, "organization.created", "organization", &org.ID, map[string]any{"name": org.Name, "slug": org.Slug})
	}
	return org, err
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
	role, err := s.repo.UserRole(ctx, userID)
	if err != nil {
		return nil, err
	}
	if role == "admin" {
		return s.repo.ListAll(ctx)
	}
	return s.repo.ListByUserID(ctx, userID)
}

func (s *Service) ListOwned(ctx context.Context, userID string) ([]Organization, error) {
	return s.repo.ListOwnedByUserID(ctx, userID)
}

func (s *Service) ListMemberships(ctx context.Context, userID string) ([]Organization, error) {
	return s.repo.ListMembershipsByUserID(ctx, userID)
}

func (s *Service) AdminList(ctx context.Context) ([]Organization, error) {
	return s.repo.ListAll(ctx)
}

func (s *Service) AdminListAudit(ctx context.Context, orgID string) ([]AuditLog, error) {
	return s.repo.ListAudit(ctx, orgID, 100, 0)
}

func (s *Service) AdminCreate(ctx context.Context, req CreateOrgRequest, actorID string) (Organization, error) {
	return s.Create(ctx, req, actorID)
}

func (s *Service) AdminUpdate(ctx context.Context, id, actorID string, req UpdateOrgRequest) (Organization, error) {
	org, err := s.repo.Update(ctx, id, req.Name, req.Slug, req.Logo)
	if err == nil {
		s.recordAudit(ctx, id, actorID, "organization.updated", "organization", &id, map[string]any{"name": org.Name, "slug": org.Slug})
	}
	return org, err
}

func (s *Service) AdminSetOwner(ctx context.Context, id, actorID string, req SetOwnerRequest) (OrganizationOwner, error) {
	owner, err := s.repo.SetOwner(ctx, id, req.UserID)
	if err == nil {
		s.recordAudit(ctx, id, actorID, "organization.owner_changed", "member", &owner.ID, map[string]any{"owner_name": owner.Name, "owner_email": owner.Email})
	}
	return owner, err
}

func (s *Service) AdminUpdateStatus(ctx context.Context, id, actorID string, req UpdateStatusRequest) (Organization, error) {
	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return Organization{}, err
	}
	org, err := s.repo.UpdateStatus(ctx, id, req.Status)
	if err == nil {
		s.recordAudit(ctx, id, actorID, "organization.status_changed", "organization", &id, map[string]any{"from": current.Status, "to": req.Status})
	}
	return org, err
}

func (s *Service) AdminUploadLogo(ctx context.Context, id, actorID string, reader io.Reader, size int64) (Organization, error) {
	org, err := s.UploadLogo(ctx, id, reader, size)
	if err == nil {
		s.recordAudit(ctx, id, actorID, "organization.logo_updated", "organization", &id, nil)
	}
	return org, err
}

func (s *Service) AdminDelete(ctx context.Context, id string) error {
	return s.delete(ctx, id)
}

func (s *Service) Update(ctx context.Context, id, actorID string, req UpdateOrgRequest) (Organization, error) {
	if err := s.requireRole(ctx, id, actorID, "owner"); err != nil {
		return Organization{}, err
	}
	if err := s.requireActive(ctx, id); err != nil {
		return Organization{}, err
	}
	org, err := s.repo.Update(ctx, id, req.Name, req.Slug, req.Logo)
	if err == nil {
		s.recordAudit(ctx, id, actorID, "organization.updated", "organization", &id, map[string]any{"name": org.Name, "slug": org.Slug})
	}
	return org, err
}

func (s *Service) OwnerUploadLogo(ctx context.Context, id, actorID string, reader io.Reader, size int64) (Organization, error) {
	if err := s.requireRole(ctx, id, actorID, "owner"); err != nil {
		return Organization{}, err
	}
	if err := s.requireActive(ctx, id); err != nil {
		return Organization{}, err
	}
	org, err := s.UploadLogo(ctx, id, reader, size)
	if err == nil {
		s.recordAudit(ctx, id, actorID, "organization.logo_updated", "organization", &id, nil)
	}
	return org, err
}

func (s *Service) Delete(ctx context.Context, id, actorID string) error {
	if err := s.requireRole(ctx, id, actorID, "owner"); err != nil {
		return err
	}
	return s.delete(ctx, id)
}

func (s *Service) delete(ctx context.Context, id string) error {
	org, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	if org.LogoKey != nil {
		if err := s.logoStore.Delete(ctx, *org.LogoKey); err != nil {
			slog.Warn("failed to delete logo for removed organization", "key", *org.LogoKey, "error", err)
		}
	}
	return nil
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
	target, err := s.repo.GetMember(ctx, orgID, userID)
	if err != nil {
		return err
	}
	if target.Role == "owner" {
		return ErrOwnerProtected
	}
	if err := s.requireActive(ctx, orgID); err != nil {
		return err
	}
	if err := s.repo.UpdateMemberRole(ctx, orgID, userID, role); err != nil {
		return err
	}
	s.recordAudit(ctx, orgID, actorID, "member.role_changed", "member", &userID, map[string]any{"from": target.Role, "to": role})
	return nil
}

func (s *Service) UpdateMemberPermissions(ctx context.Context, orgID, userID, actorID string, permissions []string) (Member, error) {
	if err := s.requireRole(ctx, orgID, actorID, "owner"); err != nil {
		return Member{}, err
	}
	if err := s.requireActive(ctx, orgID); err != nil {
		return Member{}, err
	}
	target, err := s.repo.GetMember(ctx, orgID, userID)
	if err != nil {
		return Member{}, err
	}
	if target.Role == "owner" {
		return Member{}, ErrOwnerProtected
	}
	updated, err := s.repo.UpdateMemberPermissions(ctx, orgID, userID, permissions)
	if err == nil {
		s.recordAudit(ctx, orgID, actorID, "member.permissions_changed", "member", &userID, map[string]any{"from": target.Permissions, "to": permissions})
	}
	return updated, err
}

func (s *Service) RemoveMember(ctx context.Context, orgID, userID, actorID string) error {
	if actorID != userID {
		if err := s.requireRole(ctx, orgID, actorID, "owner"); err != nil {
			return err
		}
	}
	target, err := s.repo.GetMember(ctx, orgID, userID)
	if err != nil {
		return err
	}
	if target.Role == "owner" {
		return ErrOwnerProtected
	}
	if err := s.requireActive(ctx, orgID); err != nil {
		return err
	}
	if err := s.repo.RemoveMember(ctx, orgID, userID); err != nil {
		return err
	}
	s.recordAudit(ctx, orgID, actorID, "member.removed", "member", &userID, map[string]any{"role": target.Role})
	return nil
}

func (s *Service) Invite(ctx context.Context, orgID, email, role, inviterID string) (Invitation, error) {
	if err := s.requirePermission(ctx, orgID, inviterID, PermissionMembersManage); err != nil {
		return Invitation{}, err
	}
	if err := s.requireActive(ctx, orgID); err != nil {
		return Invitation{}, err
	}
	invitation, err := s.repo.CreateInvitation(ctx, orgID, email, role, inviterID, time.Now().Add(invitationTTL))
	if err == nil {
		s.recordAudit(ctx, orgID, inviterID, "invitation.created", "invitation", &invitation.ID, map[string]any{"email": email, "role": role})
	}
	return invitation, err
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
	if err := s.requirePermission(ctx, inv.OrgID, actorID, PermissionMembersManage); err != nil {
		return err
	}
	if err := s.requireActive(ctx, inv.OrgID); err != nil {
		return err
	}
	if err := s.repo.CancelInvitation(ctx, id); err != nil {
		return err
	}
	s.recordAudit(ctx, inv.OrgID, actorID, "invitation.cancelled", "invitation", &id, map[string]any{"email": inv.Email})
	return nil
}

func (s *Service) ListAudit(ctx context.Context, orgID, actorID string) ([]AuditLog, error) {
	if err := s.requirePermission(ctx, orgID, actorID, PermissionAuditView); err != nil {
		return nil, err
	}
	return s.repo.ListAudit(ctx, orgID, 100, 0)
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

func (s *Service) requirePermission(ctx context.Context, orgID, actorID, permission string) error {
	m, err := s.repo.GetMember(ctx, orgID, actorID)
	if err != nil {
		return ErrForbidden
	}
	if m.Role == "owner" || m.Role == "admin" {
		return nil
	}
	for _, granted := range m.Permissions {
		if granted == permission {
			return nil
		}
	}
	return ErrForbidden
}

func (s *Service) requireActive(ctx context.Context, orgID string) error {
	status, err := s.repo.Status(ctx, orgID)
	if err != nil {
		return err
	}
	if status != StatusActive {
		return ErrOrganizationLocked
	}
	return nil
}

func (s *Service) recordAudit(ctx context.Context, orgID, actorID, action, targetType string, targetID *string, details map[string]any) {
	if details == nil {
		details = map[string]any{}
	}
	if err := s.repo.RecordAudit(ctx, orgID, actorID, action, targetType, targetID, details); err != nil {
		slog.Warn("failed to record organization audit log", "organization_id", orgID, "action", action, "error", err)
	}
}
