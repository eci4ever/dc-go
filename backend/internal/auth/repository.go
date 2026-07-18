package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"time"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"

	db "github.com/eci4ever/dc-go/internal/db"
	"github.com/eci4ever/dc-go/internal/user"
)

type Repository struct {
	q    *db.Queries
	pool *pgxpool.Pool
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{q: db.New(pool), pool: pool}
}

// CreateCredentialUser creates the user and credential account in one transaction.
// A failed account insert cannot leave an unusable user record behind.
func (r *Repository) CreateCredentialUser(ctx context.Context, name, email, passwordHash string) (string, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return "", err
	}
	defer tx.Rollback(ctx)

	q := db.New(tx)
	userID := uuid.New().String()
	if _, err = q.CreateUser(ctx, db.CreateUserParams{
		ID:    userID,
		Name:  name,
		Email: email,
		Image: pgtype.Text{Valid: false},
	}); err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			return "", ErrEmailExists
		}
		return "", err
	}

	if _, err = q.CreateAccount(ctx, db.CreateAccountParams{
		ID:         uuid.New().String(),
		ProviderID: "credential",
		AccountID:  email,
		UserID:     userID,
		Password:   pgtext(&passwordHash),
		Scope:      pgtype.Text{Valid: false},
	}); err != nil {
		return "", err
	}

	if err = tx.Commit(ctx); err != nil {
		return "", err
	}
	return userID, nil
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

func (r *Repository) GetAccountByProvider(ctx context.Context, providerID, accountID string) (db.Account, error) {
	return r.q.GetAccountByProvider(ctx, db.GetAccountByProviderParams{
		ProviderID: providerID,
		AccountID:  accountID,
	})
}

func (r *Repository) GetCredentialAccountByUserID(ctx context.Context, userID string) (db.Account, error) {
	return r.q.GetCredentialAccountByUserID(ctx, userID)
}

func (r *Repository) UpdateAccountPassword(ctx context.Context, accountID, password string) error {
	_, err := r.q.UpdateAccountPassword(ctx, db.UpdateAccountPasswordParams{
		ID:       accountID,
		Password: pgtext(&password),
	})
	return err
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
		ActiveTeamID:         pgtype.Text{Valid: false},
	})
	return sess, token, err
}

func (r *Repository) GetSessionContextByToken(ctx context.Context, token string) (SessionContext, error) {
	sess, err := r.q.GetSessionContextByToken(ctx, hashToken(token))
	if err != nil {
		return SessionContext{}, err
	}
	return mapSessionContext(sess), nil
}

func (r *Repository) SetActiveOrganization(ctx context.Context, token, userID, organizationID string) (SessionContext, error) {
	_, err := r.q.UpdateSessionActiveOrganization(ctx, db.UpdateSessionActiveOrganizationParams{
		Token:                hashToken(token),
		ActiveOrganizationID: pgtype.Text{String: organizationID, Valid: true},
		UserID:               userID,
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return SessionContext{}, ErrForbidden
		}
		return SessionContext{}, err
	}
	return r.GetSessionContextByToken(ctx, token)
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
	sess, err := q.CreateSession(ctx, db.CreateSessionParams{ID: uuid.New().String(), ExpiresAt: timestamptz(expiresAt), Token: hashToken(newToken), IpAddress: pgtext(&ipAddress), UserAgent: pgtext(&userAgent), UserID: userID, ActiveOrganizationID: old.ActiveOrganizationID, ActiveTeamID: old.ActiveTeamID})
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
		ID:               u.ID,
		Name:             u.Name,
		Email:            u.Email,
		EmailVerified:    u.EmailVerified,
		Image:            pgtextPtr(&u.Image),
		Role:             user.Role(u.Role),
		Banned:           u.Banned,
		BanReason:        pgtextPtr(&u.BanReason),
		BanExpires:       pgtimestamptzPtr(&u.BanExpires),
		TwoFactorEnabled: u.TwoFactorEnabled,
		CreatedAt:        u.CreatedAt.Time,
		UpdatedAt:        u.UpdatedAt.Time,
	}
}

func mapSessionContext(s db.GetSessionContextByTokenRow) SessionContext {
	return SessionContext{
		ID:                     s.ID,
		ExpiresAt:              s.ExpiresAt.Time,
		CreatedAt:              s.CreatedAt.Time,
		UpdatedAt:              s.UpdatedAt.Time,
		IPAddress:              pgtextPtr(&s.IpAddress),
		UserAgent:              pgtextPtr(&s.UserAgent),
		UserID:                 s.UserID,
		ImpersonatedBy:         pgtextPtr(&s.ImpersonatedBy),
		ActiveOrganizationID:   pgtextPtr(&s.ActiveOrganizationID),
		ActiveOrganizationRole: stringPtr(s.ActiveOrganizationRole),
		ActiveTeamID:           pgtextPtr(&s.ActiveTeamID),
	}
}

func pgtimestamptzPtr(t *pgtype.Timestamptz) *time.Time {
	if t == nil || !t.Valid {
		return nil
	}
	return &t.Time
}

func pgtextPtr(t *pgtype.Text) *string {
	if t == nil || !t.Valid {
		return nil
	}
	return &t.String
}

func stringPtr(value string) *string {
	if value == "" {
		return nil
	}
	return &value
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
