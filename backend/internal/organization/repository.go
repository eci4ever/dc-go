package organization

import (
	"context"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	db "github.com/eci4ever/dc-go/internal/db"
)

type Repository struct {
	q    *db.Queries
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{q: db.New(pool), pool: pool}
}

func (r *Repository) Create(ctx context.Context, name, slug string, logo *string, creatorID string) (Organization, error) {
	id := uuid.New().String()
	org, err := r.q.CreateOrganization(ctx, db.CreateOrganizationParams{
		ID:       id,
		Name:     name,
		Slug:     slug,
		Logo:     pgtext(logo),
		Metadata: pgtype.Text{Valid: false},
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return Organization{}, ErrSlugExists
		}
		return Organization{}, err
	}

	memberID := uuid.New().String()
	_, err = r.q.CreateMember(ctx, db.CreateMemberParams{
		ID:             memberID,
		OrganizationID: id,
		UserID:         creatorID,
		Role:           "owner",
	})
	if err != nil {
		return Organization{}, err
	}

	return mapOrg(org), nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (Organization, error) {
	org, err := r.q.GetOrganization(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Organization{}, ErrNotFound
		}
		return Organization{}, err
	}
	return mapOrg(org), nil
}

func (r *Repository) GetBySlug(ctx context.Context, slug string) (Organization, error) {
	org, err := r.q.GetOrganizationBySlug(ctx, slug)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Organization{}, ErrNotFound
		}
		return Organization{}, err
	}
	return mapOrg(org), nil
}

func (r *Repository) ListByUserID(ctx context.Context, userID string) ([]Organization, error) {
	rows, err := r.q.ListOrganizationsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	orgs := make([]Organization, len(rows))
	for i, org := range rows {
		orgs[i] = mapOrg(org)
	}
	return orgs, nil
}

func (r *Repository) ListAll(ctx context.Context) ([]Organization, error) {
	rows, err := r.q.ListOrganizations(ctx)
	if err != nil {
		return nil, err
	}
	orgs := make([]Organization, len(rows))
	for i, org := range rows {
		orgs[i] = mapOrg(org)
	}
	return orgs, nil
}

func (r *Repository) UserRole(ctx context.Context, userID string) (string, error) {
	u, err := r.q.GetUser(ctx, userID)
	if err != nil {
		return "", err
	}
	return u.Role, nil
}

func (r *Repository) Update(ctx context.Context, id, name, slug string, logo *string) (Organization, error) {
	org, err := r.q.UpdateOrganization(ctx, db.UpdateOrganizationParams{
		ID:       id,
		Name:     name,
		Slug:     slug,
		Logo:     pgtext(logo),
		Metadata: pgtype.Text{Valid: false},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Organization{}, ErrNotFound
		}
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return Organization{}, ErrSlugExists
		}
		return Organization{}, err
	}
	return mapOrg(org), nil
}

func (r *Repository) UpdateLogo(ctx context.Context, id, logoURL, key, contentType string, updatedAt time.Time) (Organization, error) {
	org, err := r.q.UpdateOrganizationLogo(ctx, db.UpdateOrganizationLogoParams{
		ID:              id,
		Logo:            pgtype.Text{String: logoURL, Valid: true},
		LogoKey:         pgtype.Text{String: key, Valid: true},
		LogoContentType: pgtype.Text{String: contentType, Valid: true},
		LogoUpdatedAt:   pgtype.Timestamptz{Time: updatedAt, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Organization{}, ErrNotFound
		}
		return Organization{}, err
	}
	return mapOrg(org), nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	_, err := r.q.GetOrganization(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	return r.q.DeleteOrganization(ctx, id)
}

func (r *Repository) GetMember(ctx context.Context, orgID, userID string) (Member, error) {
	m, err := r.q.GetMember(ctx, db.GetMemberParams{
		OrganizationID: orgID,
		UserID:         userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Member{}, ErrMemberNotFound
		}
		return Member{}, err
	}
	return Member{
		ID:        m.ID,
		OrgID:     m.OrganizationID,
		UserID:    m.UserID,
		Role:      m.Role,
		CreatedAt: m.CreatedAt.Time.Format(time3339),
	}, nil
}

func (r *Repository) GetMembers(ctx context.Context, orgID string) ([]Member, error) {
	rows, err := r.q.ListMembersByOrganizationID(ctx, orgID)
	if err != nil {
		return nil, err
	}
	members := make([]Member, len(rows))
	for i, m := range rows {
		members[i] = Member{
			ID:        m.ID,
			OrgID:     m.OrganizationID,
			UserID:    m.UserID,
			Role:      m.Role,
			CreatedAt: m.CreatedAt.Time.Format(time3339),
		}
		members[i].User.Name = m.Name
		members[i].User.Email = m.Email
		members[i].User.Image = pgtextPtr(&m.Image)
	}
	return members, nil
}

func (r *Repository) UpdateMemberRole(ctx context.Context, orgID, userID, role string) error {
	_, err := r.q.UpdateMemberRole(ctx, db.UpdateMemberRoleParams{
		OrganizationID: orgID,
		UserID:         userID,
		Role:           role,
	})
	return err
}

func (r *Repository) RemoveMember(ctx context.Context, orgID, userID string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	q := db.New(tx)
	if err := q.DeleteMember(ctx, db.DeleteMemberParams{
		OrganizationID: orgID,
		UserID:         userID,
	}); err != nil {
		return err
	}
	if err := q.ClearActiveOrganizationForMember(ctx, db.ClearActiveOrganizationForMemberParams{
		ActiveOrganizationID: pgtype.Text{String: orgID, Valid: true},
		UserID:               userID,
	}); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *Repository) IsMember(ctx context.Context, orgID, userID string) (bool, error) {
	_, err := r.q.GetMember(ctx, db.GetMemberParams{
		OrganizationID: orgID,
		UserID:         userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *Repository) CreateInvitation(ctx context.Context, orgID, email, role, inviterID string, expiresAt time.Time) (Invitation, error) {
	id := uuid.New().String()
	inv, err := r.q.CreateInvitation(ctx, db.CreateInvitationParams{
		ID:             id,
		OrganizationID: orgID,
		Email:          email,
		Role:           role,
		ExpiresAt:      timestamptz(expiresAt),
		InviterID:      inviterID,
	})
	if err != nil {
		return Invitation{}, err
	}
	return mapInvitation(inv), nil
}

func (r *Repository) GetInvitation(ctx context.Context, id string) (Invitation, error) {
	inv, err := r.q.GetInvitation(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Invitation{}, ErrInvitationNotFound
		}
		return Invitation{}, err
	}
	return mapInvitation(inv), nil
}

func (r *Repository) ListInvitationsByOrgID(ctx context.Context, orgID string) ([]Invitation, error) {
	rows, err := r.q.ListInvitationsByOrganizationID(ctx, orgID)
	if err != nil {
		return nil, err
	}
	invs := make([]Invitation, len(rows))
	for i, inv := range rows {
		invs[i] = mapInvitation(inv)
	}
	return invs, nil
}

func (r *Repository) AcceptInvitation(ctx context.Context, id, userID string) error {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	q := db.New(tx)
	inv, err := q.GetInvitationForUpdate(ctx, id)
	if err != nil {
		return ErrInvitationNotFound
	}

	if inv.Status != "pending" {
		return errors.New("invitation is not pending")
	}
	u, err := q.GetUser(ctx, userID)
	if err != nil || !strings.EqualFold(u.Email, inv.Email) {
		return ErrInvitationNotFound
	}

	if inv.ExpiresAt.Time.Before(time.Now()) {
		if _, err = q.UpdateInvitationStatus(ctx, db.UpdateInvitationStatusParams{
			ID:     id,
			Status: "expired",
		}); err != nil {
			return err
		}
		if err = tx.Commit(ctx); err != nil {
			return err
		}
		return ErrInvitationExpired
	}

	memberID := uuid.New().String()
	_, err = q.CreateMember(ctx, db.CreateMemberParams{
		ID:             memberID,
		OrganizationID: inv.OrganizationID,
		UserID:         userID,
		Role:           inv.Role,
	})
	if err != nil {
		return err
	}

	_, err = q.UpdateInvitationStatus(ctx, db.UpdateInvitationStatusParams{
		ID:     id,
		Status: "accepted",
	})
	if err != nil {
		return err
	}
	return tx.Commit(ctx)
}

func (r *Repository) DeclineInvitation(ctx context.Context, id, userID string) error {
	inv, err := r.GetInvitation(ctx, id)
	if err != nil {
		return err
	}
	u, err := r.q.GetUser(ctx, userID)
	if err != nil || !strings.EqualFold(u.Email, inv.Email) {
		return ErrInvitationNotFound
	}
	_, err = r.q.UpdateInvitationStatus(ctx, db.UpdateInvitationStatusParams{
		ID:     id,
		Status: "declined",
	})
	return err
}

func (r *Repository) CancelInvitation(ctx context.Context, id string) error {
	return r.q.DeleteInvitation(ctx, id)
}

func mapOrg(o db.Organization) Organization {
	return Organization{
		ID:              o.ID,
		Name:            o.Name,
		Slug:            o.Slug,
		Logo:            pgtextPtr(&o.Logo),
		CreatedAt:       o.CreatedAt.Time.Format(time3339),
		LogoKey:         pgtextPtr(&o.LogoKey),
		LogoContentType: pgtextPtr(&o.LogoContentType),
		LogoUpdatedAt:   pgtimestamptzPtr(&o.LogoUpdatedAt),
	}
}

func mapInvitation(i db.Invitation) Invitation {
	inv := Invitation{
		ID:        i.ID,
		OrgID:     i.OrganizationID,
		Email:     i.Email,
		Role:      i.Role,
		Status:    i.Status,
		InviterID: i.InviterID,
		ExpiresAt: i.ExpiresAt.Time.Format(time3339),
		CreatedAt: i.CreatedAt.Time.Format(time3339),
	}
	if i.TeamID.Valid {
		s := i.TeamID.String
		inv.TeamID = &s
	}
	return inv
}

func pgtextPtr(t *pgtype.Text) *string {
	if t == nil || !t.Valid {
		return nil
	}
	return &t.String
}

func pgtext(s *string) pgtype.Text {
	if s == nil {
		return pgtype.Text{Valid: false}
	}
	return pgtype.Text{String: *s, Valid: true}
}

func pgtimestamptzPtr(t *pgtype.Timestamptz) *time.Time {
	if t == nil || !t.Valid {
		return nil
	}
	return &t.Time
}

func textPtr(s string) pgtype.Text {
	return pgtype.Text{String: s, Valid: true}
}

func timestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}
