package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"time"

	"crypto/rand"
	"encoding/base64"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	db "dc-express/internal/db"
	"dc-express/internal/user"
)

type Repository struct {
	q    *db.Queries
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{q: db.New(pool), pool: pool}
}

func (r *Repository) CreateUser(ctx context.Context, name, email string, image, role *string) (string, error) {
	id := uuid.New().String()
	_, err := r.q.CreateUser(ctx, db.CreateUserParams{
		ID:    id,
		Name:  name,
		Email: email,
		Image: pgtext(image),
		Role:  pgtext(role),
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return "", ErrEmailExists
		}
		return "", err
	}
	return id, nil
}

func (r *Repository) GetUserByEmail(ctx context.Context, email string) (user.User, error) {
	return getUser(ctx, r.q.GetUserByEmail, email)
}

func (r *Repository) GetUserByID(ctx context.Context, id string) (user.User, error) {
	return getUser(ctx, r.q.GetUser, id)
}

func getUser(ctx context.Context, fn func(context.Context, string) (db.User, error), id string) (user.User, error) {
	u, err := fn(ctx, id)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return user.User{}, user.ErrNotFound
		}
		return user.User{}, err
	}
	return mapUser(u), nil
}

func (r *Repository) CreateAccount(ctx context.Context, userID, providerID, accountID string, password *string) error {
	id := uuid.New().String()
	_, err := r.q.CreateAccount(ctx, db.CreateAccountParams{
		ID:         id,
		ProviderID: providerID,
		AccountID:  accountID,
		UserID:     userID,
		Password:   pgtext(password),
		Scope:      pgtype.Text{Valid: false},
	})
	return err
}

func (r *Repository) GetAccountByProvider(ctx context.Context, providerID, accountID string) (db.Account, error) {
	return r.q.GetAccountByProvider(ctx, db.GetAccountByProviderParams{
		ProviderID: providerID,
		AccountID:  accountID,
	})
}

func (r *Repository) CreateSession(ctx context.Context, userID, ipAddress, userAgent string, expiresAt time.Time) (db.Session, string, error) {
	id := uuid.New().String()
	token, err := newRefreshToken()
	if err != nil {
		return db.Session{}, "", err
	}
	sess, err := r.q.CreateSession(ctx, db.CreateSessionParams{
		ID:                   id,
		ExpiresAt:            timestamptz(expiresAt),
		Token:                hashToken(token),
		IpAddress:            pgtext(&ipAddress),
		UserAgent:            pgtext(&userAgent),
		UserID:               userID,
		ActiveOrganizationID: pgtype.Text{Valid: false},
	})
	return sess, token, err
}

func (r *Repository) GetSessionByToken(ctx context.Context, token string) (db.Session, error) {
	return r.q.GetSessionByToken(ctx, hashToken(token))
}

func (r *Repository) DeleteSession(ctx context.Context, id string) error {
	return r.q.DeleteSession(ctx, id)
}

func (r *Repository) DeleteSessionByToken(ctx context.Context, token string) error {
	return r.q.DeleteSessionByToken(ctx, hashToken(token))
}

// RotateSession atomically consumes the old refresh session and creates its replacement.
// This prevents two concurrent refresh requests from both succeeding.
func (r *Repository) RotateSession(ctx context.Context, token, userID, ipAddress, userAgent string, expiresAt time.Time) (db.Session, string, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return db.Session{}, "", err
	}
	defer tx.Rollback(ctx)
	q := db.New(tx)
	old, err := q.GetSessionByToken(ctx, hashToken(token))
	if err != nil {
		return db.Session{}, "", err
	}
	if old.UserID != userID {
		return db.Session{}, "", errors.New("session user mismatch")
	}
	if err = q.DeleteSession(ctx, old.ID); err != nil {
		return db.Session{}, "", err
	}
	newToken, err := newRefreshToken()
	if err != nil {
		return db.Session{}, "", err
	}
	sess, err := q.CreateSession(ctx, db.CreateSessionParams{ID: uuid.New().String(), ExpiresAt: timestamptz(expiresAt), Token: hashToken(newToken), IpAddress: pgtext(&ipAddress), UserAgent: pgtext(&userAgent), UserID: userID, ActiveOrganizationID: pgtype.Text{Valid: false}})
	if err != nil {
		return db.Session{}, "", err
	}
	if err = tx.Commit(ctx); err != nil {
		return db.Session{}, "", err
	}
	return sess, newToken, nil
}

func (r *Repository) DeleteSessionsByUserID(ctx context.Context, userID string) error {
	_, err := r.pool.Exec(ctx, `DELETE FROM "session" WHERE user_id=$1`, userID)
	return err
}

func hashToken(token string) string {
	sum := sha256.Sum256([]byte(token))
	return hex.EncodeToString(sum[:])
}

func (r *Repository) ListSessionsByUserID(ctx context.Context, userID string) ([]db.Session, error) {
	return r.q.ListSessionsByUserID(ctx, userID)
}

func mapUser(u db.User) user.User {
	return user.User{
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

func timestamptz(t time.Time) pgtype.Timestamptz {
	return pgtype.Timestamptz{Time: t, Valid: true}
}

func newRefreshToken() (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(b), nil
}
