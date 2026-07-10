package auth

import (
	"context"
	"errors"
	"time"

	"golang.org/x/crypto/bcrypt"

	"dc-express/internal/user"
)

const (
	accessTokenTTL  = 15 * time.Minute
	refreshTokenTTL = 7 * 24 * time.Hour
)

type Service struct {
	repo     *Repository
	jwt      *JWTService
	userRepo *user.Repository
}

func NewService(repo *Repository, jwt *JWTService, userRepo *user.Repository) *Service {
	return &Service{repo: repo, jwt: jwt, userRepo: userRepo}
}

func (s *Service) Register(ctx context.Context, req RegisterRequest, ipAddress, userAgent string) (TokenResponse, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return TokenResponse{}, err
	}

	userID, err := s.repo.CreateUser(ctx, req.Name, req.Email, nil)
	if err != nil {
		return TokenResponse{}, err
	}

	hashStr := string(hash)
	if err := s.repo.CreateAccount(ctx, userID, "credential", req.Email, &hashStr); err != nil {
		return TokenResponse{}, err
	}

	return s.createSession(ctx, userID, ipAddress, userAgent)
}

func (s *Service) Login(ctx context.Context, req LoginRequest, ipAddress, userAgent string) (TokenResponse, error) {
	u, err := s.repo.GetUserByEmail(ctx, req.Email)
	if err != nil {
		if errors.Is(err, user.ErrNotFound) {
			return TokenResponse{}, ErrInvalidCredentials
		}
		return TokenResponse{}, err
	}

	if u.Banned {
		return TokenResponse{}, ErrUserBanned
	}

	acct, err := s.repo.GetAccountByProvider(ctx, "credential", req.Email)
	if err != nil {
		return TokenResponse{}, ErrInvalidCredentials
	}

	if !acct.Password.Valid {
		return TokenResponse{}, ErrInvalidCredentials
	}

	if err := bcrypt.CompareHashAndPassword([]byte(acct.Password.String), []byte(req.Password)); err != nil {
		return TokenResponse{}, ErrInvalidCredentials
	}

	return s.createSession(ctx, u.ID, ipAddress, userAgent)
}

func (s *Service) Refresh(ctx context.Context, refreshToken string) (TokenResponse, error) {
	if refreshToken == "" {
		return TokenResponse{}, ErrInvalidToken
	}
	sess, err := s.repo.GetSessionByToken(ctx, refreshToken)
	if err != nil {
		return TokenResponse{}, ErrInvalidToken
	}

	if sess.ExpiresAt.Time.Before(time.Now()) {
		s.repo.DeleteSession(ctx, sess.ID)
		return TokenResponse{}, ErrExpiredToken
	}

	return s.rotateSession(ctx, refreshToken, sess.UserID, sess.IpAddress.String, sess.UserAgent.String)
}

func (s *Service) rotateSession(ctx context.Context, token, userID, ipAddress, userAgent string) (TokenResponse, error) {
	u, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return TokenResponse{}, err
	}
	_, refreshToken, err := s.repo.RotateSession(ctx, token, userID, ipAddress, userAgent, time.Now().Add(refreshTokenTTL))
	if err != nil {
		return TokenResponse{}, ErrInvalidToken
	}
	accessToken, err := s.jwt.Sign(userID)
	if err != nil {
		return TokenResponse{}, err
	}
	sessionContext, err := s.repo.GetSessionContextByToken(ctx, refreshToken)
	if err != nil {
		return TokenResponse{}, err
	}
	return TokenResponse{AccessToken: accessToken, RefreshToken: refreshToken, User: toAuthUser(u), Session: toAuthSession(sessionContext)}, nil
}

func (s *Service) GetSession(ctx context.Context, userID, refreshToken string) (SessionResponse, error) {
	if refreshToken == "" {
		return SessionResponse{}, ErrInvalidToken
	}
	u, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return SessionResponse{}, err
	}

	sess, err := s.repo.GetSessionContextByToken(ctx, refreshToken)
	if err != nil {
		return SessionResponse{}, ErrInvalidToken
	}
	if sess.UserID != userID || sess.ExpiresAt.Before(time.Now()) {
		return SessionResponse{}, ErrInvalidToken
	}
	return SessionResponse{User: toAuthUser(u), Session: toAuthSession(sess)}, nil
}

func (s *Service) SetActiveOrganization(ctx context.Context, userID, refreshToken, organizationID string) (SessionResponse, error) {
	if refreshToken == "" {
		return SessionResponse{}, ErrInvalidToken
	}
	sess, err := s.repo.SetActiveOrganization(ctx, refreshToken, userID, organizationID)
	if err != nil {
		return SessionResponse{}, err
	}
	u, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return SessionResponse{}, err
	}
	return SessionResponse{User: toAuthUser(u), Session: toAuthSession(sess)}, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	if refreshToken == "" {
		return nil
	}
	return s.repo.DeleteSessionByToken(ctx, refreshToken)
}

func (s *Service) ChangePassword(ctx context.Context, userID string, req ChangePasswordRequest) error {
	account, err := s.repo.GetCredentialAccountByUserID(ctx, userID)
	if err != nil || !account.Password.Valid {
		return ErrIncorrectPassword
	}
	if err := bcrypt.CompareHashAndPassword([]byte(account.Password.String), []byte(req.CurrentPassword)); err != nil {
		return ErrIncorrectPassword
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return err
	}
	return s.repo.UpdateAccountPassword(ctx, account.ID, string(hash))
}

func (s *Service) ListSessions(ctx context.Context, userID, refreshToken string) ([]ManagedSession, error) {
	if refreshToken == "" {
		return nil, ErrInvalidToken
	}
	current, err := s.repo.GetSessionContextByToken(ctx, refreshToken)
	if err != nil || current.UserID != userID {
		return nil, ErrInvalidToken
	}
	sessions, err := s.repo.ListSessionsByUserID(ctx, userID)
	if err != nil {
		return nil, err
	}
	result := make([]ManagedSession, len(sessions))
	for i, session := range sessions {
		result[i] = ManagedSession{
			ID:        session.ID,
			ExpiresAt: formatTime(session.ExpiresAt.Time),
			CreatedAt: formatTime(session.CreatedAt.Time),
			UpdatedAt: formatTime(session.UpdatedAt.Time),
			IPAddress: pgtextPtr(&session.IpAddress),
			UserAgent: pgtextPtr(&session.UserAgent),
			Current:   session.ID == current.ID,
		}
	}
	return result, nil
}

func (s *Service) RevokeSession(ctx context.Context, userID, refreshToken, sessionID string) error {
	if refreshToken == "" {
		return ErrInvalidToken
	}
	current, err := s.repo.GetSessionContextByToken(ctx, refreshToken)
	if err != nil || current.UserID != userID {
		return ErrInvalidToken
	}
	if current.ID == sessionID {
		return ErrCurrentSession
	}
	sessions, err := s.repo.ListSessionsByUserID(ctx, userID)
	if err != nil {
		return err
	}
	for _, session := range sessions {
		if session.ID == sessionID {
			return s.repo.DeleteSession(ctx, sessionID)
		}
	}
	return ErrSessionNotFound
}

func (s *Service) createSession(ctx context.Context, userID, ipAddress, userAgent string) (TokenResponse, error) {
	u, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return TokenResponse{}, err
	}

	_, refreshToken, err := s.repo.CreateSession(ctx, userID, ipAddress, userAgent, time.Now().Add(refreshTokenTTL))
	if err != nil {
		return TokenResponse{}, err
	}

	accessToken, err := s.jwt.Sign(u.ID)
	if err != nil {
		return TokenResponse{}, err
	}

	sessionContext, err := s.repo.GetSessionContextByToken(ctx, refreshToken)
	if err != nil {
		return TokenResponse{}, err
	}
	return TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
		User:         toAuthUser(u),
		Session:      toAuthSession(sessionContext),
	}, nil
}

func (s *Service) AccessToken(userID string) (string, error) { return s.jwt.Sign(userID) }

func toAuthUser(u user.User) AuthUser {
	var banExpires *string
	if u.BanExpires != nil {
		value := formatTime(*u.BanExpires)
		banExpires = &value
	}
	return AuthUser{
		ID:               u.ID,
		Name:             u.Name,
		Email:            u.Email,
		EmailVerified:    u.EmailVerified,
		Image:            u.Image,
		Role:             u.Role,
		Banned:           u.Banned,
		BanReason:        u.BanReason,
		BanExpires:       banExpires,
		TwoFactorEnabled: u.TwoFactorEnabled,
		CreatedAt:        formatTime(u.CreatedAt),
		UpdatedAt:        formatTime(u.UpdatedAt),
	}
}

func toAuthSession(s SessionContext) AuthSession {
	return AuthSession{
		ID:                     s.ID,
		ExpiresAt:              formatTime(s.ExpiresAt),
		CreatedAt:              formatTime(s.CreatedAt),
		UpdatedAt:              formatTime(s.UpdatedAt),
		IPAddress:              s.IPAddress,
		UserAgent:              s.UserAgent,
		UserID:                 s.UserID,
		ImpersonatedBy:         s.ImpersonatedBy,
		ActiveOrganizationID:   s.ActiveOrganizationID,
		ActiveOrganizationRole: s.ActiveOrganizationRole,
		ActiveTeamID:           s.ActiveTeamID,
	}
}

func formatTime(value time.Time) string {
	return value.UTC().Format(time.RFC3339Nano)
}
