package user

import (
	"context"
	"errors"
)

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) Create(ctx context.Context, req CreateUserRequest) (UserResponse, error) {
	u, err := s.repo.Create(ctx, req.Name, req.Email)
	if err != nil {
		return UserResponse{}, err
	}
	return toResponse(u), nil
}

func (s *Service) GetByID(ctx context.Context, id int32) (UserResponse, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return UserResponse{}, err
	}
	return toResponse(u), nil
}

func (s *Service) List(ctx context.Context) ([]UserResponse, error) {
	users, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	resp := make([]UserResponse, len(users))
	for i, u := range users {
		resp[i] = toResponse(u)
	}
	return resp, nil
}

func (s *Service) Update(ctx context.Context, id int32, req UpdateUserRequest) (UserResponse, error) {
	u, err := s.repo.Update(ctx, id, req.Name, req.Email)
	if err != nil {
		return UserResponse{}, err
	}
	return toResponse(u), nil
}

func (s *Service) Delete(ctx context.Context, id int32) error {
	return s.repo.Delete(ctx, id)
}

func (s *Service) Ping(ctx context.Context) error {
	_, err := s.repo.GetByID(ctx, 0)
	if err != nil && !errors.Is(err, ErrNotFound) {
		return err
	}
	return nil
}
