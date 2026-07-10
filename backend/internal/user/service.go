package user

import "context"

type Service struct {
	repo *Repository
}

func NewService(repo *Repository) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetByID(ctx context.Context, id string) (UserResponse, error) {
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
	return toResponses(users), nil
}

func (s *Service) Update(ctx context.Context, id string, req UpdateUserRequest) (UserResponse, error) {
	u, err := s.repo.Update(ctx, id, req.Name, req.Email, req.Image, nil)
	if err != nil {
		return UserResponse{}, err
	}
	return toResponse(u), nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	return s.repo.Delete(ctx, id)
}
