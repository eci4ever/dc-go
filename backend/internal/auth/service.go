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

func (s *Service) Register(ctx context.Context, req RegisterRequest) (TokenResponse, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return TokenResponse{}, err
	}

	userID, err := s.repo.CreateUser(ctx, req.Name, req.Email, nil, nil)
	if err != nil {
		return TokenResponse{}, err
	}

	hashStr := string(hash)
	if err := s.repo.CreateAccount(ctx, userID, "credential", req.Email, &hashStr); err != nil {
		return TokenResponse{}, err
	}

	return s.createSession(ctx, userID, "", "")
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
	sess, err := s.repo.GetSessionByToken(ctx, refreshToken)
	if err != nil {
		return TokenResponse{}, ErrInvalidToken
	}

	if sess.ExpiresAt.Time.Before(time.Now()) {
		s.repo.DeleteSession(ctx, sess.ID)
		return TokenResponse{}, ErrExpiredToken
	}

	s.repo.DeleteSession(ctx, sess.ID)

	return s.createSession(ctx, sess.UserID, sess.IpAddress.String, sess.UserAgent.String)
}

func (s *Service) GetSession(ctx context.Context, userID string) (SessionResponse, error) {
	u, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return SessionResponse{}, err
	}

	sessions, err := s.repo.ListSessionsByUserID(ctx, userID)
	if err != nil {
		return SessionResponse{}, err
	}

	var active *AuthSession
	for i := range sessions {
		if sessions[i].ExpiresAt.Time.After(time.Now()) {
			active = &AuthSession{
				ID:        sessions[i].ID,
				ExpiresAt: sessions[i].ExpiresAt.Time.Format(time.RFC3339),
				CreatedAt: sessions[i].CreatedAt.Time.Format(time.RFC3339),
				UserID:    sessions[i].UserID,
			}
			break
		}
	}

	au := toAuthUser(u)

	if active == nil {
		return SessionResponse{User: au}, nil
	}

	return SessionResponse{
		User:    au,
		Session: *active,
	}, nil
}

func (s *Service) Logout(ctx context.Context, refreshToken string) error {
	return s.repo.DeleteSessionByToken(ctx, refreshToken)
}

func (s *Service) createSession(ctx context.Context, userID, ipAddress, userAgent string) (TokenResponse, error) {
	u, err := s.repo.GetUserByID(ctx, userID)
	if err != nil {
		return TokenResponse{}, err
	}

	sess, err := s.repo.CreateSession(ctx, userID, ipAddress, userAgent, time.Now().Add(refreshTokenTTL))
	if err != nil {
		return TokenResponse{}, err
	}

	accessToken, err := s.jwt.Sign(u.ID)
	if err != nil {
		return TokenResponse{}, err
	}

	return TokenResponse{
		AccessToken:  accessToken,
		RefreshToken: sess.Token,
		User:         toAuthUser(u),
		Session: AuthSession{
			ID:        sess.ID,
			ExpiresAt: sess.ExpiresAt.Time.Format(time.RFC3339),
			CreatedAt: sess.CreatedAt.Time.Format(time.RFC3339),
			UserID:    sess.UserID,
		},
	}, nil
}

func toAuthUser(u user.User) AuthUser {
	return AuthUser{
		ID:            u.ID,
		Name:          u.Name,
		Email:         u.Email,
		EmailVerified: u.EmailVerified,
		Image:         u.Image,
		Role:          u.Role,
		Banned:        u.Banned,
		CreatedAt:     u.CreatedAt.Format(time.RFC3339),
		UpdatedAt:     u.UpdatedAt.Format(time.RFC3339),
	}
}
