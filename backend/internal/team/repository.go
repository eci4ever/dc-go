package team

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"

	db "dc-express/internal/db"
)

type Repository struct {
	q *db.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{q: db.New(pool)}
}

func (r *Repository) Create(ctx context.Context, orgID, name string) (Team, error) {
	id := uuid.New().String()
	team, err := r.q.CreateTeam(ctx, db.CreateTeamParams{
		ID:             id,
		Name:           name,
		OrganizationID: orgID,
	})
	if err != nil {
		return Team{}, err
	}
	return mapTeam(team), nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (Team, error) {
	team, err := r.q.GetTeam(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Team{}, ErrNotFound
		}
		return Team{}, err
	}
	return mapTeam(team), nil
}

func (r *Repository) ListByOrgID(ctx context.Context, orgID string) ([]Team, error) {
	rows, err := r.q.ListTeamsByOrganizationID(ctx, orgID)
	if err != nil {
		return nil, err
	}
	teams := make([]Team, len(rows))
	for i, t := range rows {
		teams[i] = mapTeam(t)
	}
	return teams, nil
}

func (r *Repository) Update(ctx context.Context, id, name string) (Team, error) {
	team, err := r.q.UpdateTeam(ctx, db.UpdateTeamParams{
		ID:   id,
		Name: name,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return Team{}, ErrNotFound
		}
		return Team{}, err
	}
	return mapTeam(team), nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	_, err := r.q.GetTeam(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	return r.q.DeleteTeam(ctx, id)
}

func (r *Repository) AddMember(ctx context.Context, teamID, userID string) (TeamMember, error) {
	id := uuid.New().String()
	_, err := r.q.CreateTeamMember(ctx, db.CreateTeamMemberParams{
		ID:     id,
		TeamID: teamID,
		UserID: userID,
	})
	if err != nil {
		return TeamMember{}, err
	}

	members, err := r.q.ListTeamMembers(ctx, teamID)
	if err != nil {
		return TeamMember{}, err
	}
	for _, m := range members {
		if m.UserID == userID {
			return mapTeamMemberRow(m), nil
		}
	}
	return TeamMember{}, ErrMemberNotFound
}

func (r *Repository) GetMembers(ctx context.Context, teamID string) ([]TeamMember, error) {
	rows, err := r.q.ListTeamMembers(ctx, teamID)
	if err != nil {
		return nil, err
	}
	members := make([]TeamMember, len(rows))
	for i, m := range rows {
		members[i] = mapTeamMemberRow(m)
	}
	return members, nil
}

func (r *Repository) RemoveMember(ctx context.Context, teamID, userID string) error {
	return r.q.DeleteTeamMember(ctx, db.DeleteTeamMemberParams{
		TeamID: teamID,
		UserID: userID,
	})
}

func (r *Repository) IsMember(ctx context.Context, teamID, userID string) (bool, error) {
	_, err := r.q.GetTeamMember(ctx, db.GetTeamMemberParams{
		TeamID: teamID,
		UserID: userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

func (r *Repository) CheckOrgMembership(ctx context.Context, orgID, userID string) (bool, error) {
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

func (r *Repository) MemberRole(ctx context.Context, orgID, userID string) (string, error) {
	m, err := r.q.GetMember(ctx, db.GetMemberParams{OrganizationID: orgID, UserID: userID})
	if errors.Is(err, pgx.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", err
	}
	return m.Role, nil
}

func mapTeam(t db.Team) Team {
	return Team{
		ID:        t.ID,
		Name:      t.Name,
		OrgID:     t.OrganizationID,
		CreatedAt: t.CreatedAt.Time.Format(time3339),
		UpdatedAt: t.UpdatedAt.Time.Format(time3339),
	}
}

func mapTeamMemberRow(m db.ListTeamMembersRow) TeamMember {
	return TeamMember{
		ID:        m.ID,
		TeamID:    m.TeamID,
		UserID:    m.UserID,
		CreatedAt: m.CreatedAt.Time.Format(time3339),
		Name:      m.Name,
		Email:     m.Email,
	}
}
