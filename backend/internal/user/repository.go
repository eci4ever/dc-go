package user

import (
	"context"
	"errors"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	db "dc-express/internal/db"
)

type Repository struct {
	q *db.Queries
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{q: db.New(pool)}
}

func (r *Repository) GetByID(ctx context.Context, id string) (User, error) {
	u, err := r.q.GetUser(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, err
	}
	return toDomain(u), nil
}

func (r *Repository) GetByEmail(ctx context.Context, email string) (User, error) {
	u, err := r.q.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, err
	}
	return toDomain(u), nil
}

func (r *Repository) List(ctx context.Context) ([]User, error) {
	rows, err := r.q.ListUsers(ctx)
	if err != nil {
		return nil, err
	}
	users := make([]User, len(rows))
	for i, u := range rows {
		users[i] = toDomain(u)
	}
	return users, nil
}

func (r *Repository) Create(ctx context.Context, id, name, email string, image, role *string) (User, error) {
	u, err := r.q.CreateUser(ctx, db.CreateUserParams{
		ID:    id,
		Name:  name,
		Email: email,
		Image: pgtext(image),
		Role:  pgtext(role),
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return User{}, ErrEmailExists
		}
		return User{}, err
	}
	return toDomain(u), nil
}

func (r *Repository) Update(ctx context.Context, id, name, email string, image, role *string) (User, error) {
	u, err := r.q.UpdateUser(ctx, db.UpdateUserParams{
		ID:    id,
		Name:  name,
		Email: email,
		Image: pgtext(image),
		Role:  pgtext(role),
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, err
	}
	return toDomain(u), nil
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	_, err := r.q.GetUser(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return ErrNotFound
		}
		return err
	}
	return r.q.DeleteUser(ctx, id)
}

func toDomain(u db.User) User {
	return User{
		ID:            u.ID,
		Name:          u.Name,
		Email:         u.Email,
		EmailVerified: u.EmailVerified,
		Image:         pgtextPtr(&u.Image),
		Role:          pgtextPtr(&u.Role),
		Banned:        u.Banned.Bool,
		CreatedAt:     u.CreatedAt.Time,
		UpdatedAt:     u.UpdatedAt.Time,
	}
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
