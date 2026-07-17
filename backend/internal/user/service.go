package user

import (
	"bytes"
	"context"
	"fmt"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/eci4ever/dc-go/internal/storage"
	"github.com/google/uuid"
)

const (
	maxAvatarUploadBytes = 2 << 20
	maxAvatarStoredBytes = 5 << 20
	maxAvatarDimension   = 2048
)

type userRepository interface {
	GetByID(context.Context, string) (User, error)
	List(context.Context) ([]User, error)
	Update(context.Context, string, string) (User, error)
	UpdateAvatar(context.Context, string, string, string, string, time.Time) (User, error)
	ClearAvatar(context.Context, string) (User, error)
	UpdateRole(context.Context, string, Role) (User, error)
	Delete(context.Context, string) error
}

type Service struct {
	repo        userRepository
	avatarStore storage.ObjectStore
}

func NewService(repo userRepository, avatarStore storage.ObjectStore) *Service {
	return &Service{repo: repo, avatarStore: avatarStore}
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
	u, err := s.repo.Update(ctx, id, req.Name)
	if err != nil {
		return UserResponse{}, err
	}
	return toResponse(u), nil
}

type AvatarFile struct {
	Data        []byte
	ContentType string
	ETag        string
}

func (s *Service) UploadAvatar(ctx context.Context, id string, reader io.Reader, size int64) (UserResponse, error) {
	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return UserResponse{}, err
	}

	data, contentType, extension, err := sanitizeAvatar(reader, size)
	if err != nil {
		return UserResponse{}, err
	}
	key := fmt.Sprintf("avatars/%s/%s.%s", id, uuid.NewString(), extension)
	if err := s.avatarStore.Put(ctx, key, contentType, bytes.NewReader(data), int64(len(data))); err != nil {
		slog.Warn("avatar upload failed", "key", key, "error", err)
		return UserResponse{}, fmt.Errorf("%w: %v", ErrAvatarUnavailable, err)
	}

	updatedAt := time.Now().UTC()
	imageURL := fmt.Sprintf("/api/v1/users/%s/avatar?v=%d", id, updatedAt.UnixNano())
	updated, err := s.repo.UpdateAvatar(ctx, id, imageURL, key, contentType, updatedAt)
	if err != nil {
		if cleanupErr := s.avatarStore.Delete(ctx, key); cleanupErr != nil {
			slog.Warn("failed to clean up unreferenced avatar", "key", key, "error", cleanupErr)
		}
		return UserResponse{}, err
	}

	if current.AvatarKey != nil && *current.AvatarKey != key {
		if err := s.avatarStore.Delete(ctx, *current.AvatarKey); err != nil {
			slog.Warn("failed to delete replaced avatar", "key", *current.AvatarKey, "error", err)
		}
	}
	return toResponse(updated), nil
}

func (s *Service) GetAvatar(ctx context.Context, id string) (AvatarFile, error) {
	u, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return AvatarFile{}, err
	}
	if u.AvatarKey == nil || u.AvatarContentType == nil {
		return AvatarFile{}, ErrNoAvatar
	}

	object, err := s.avatarStore.Get(ctx, *u.AvatarKey)
	if err != nil {
		slog.Warn("avatar download failed", "key", *u.AvatarKey, "error", err)
		return AvatarFile{}, fmt.Errorf("%w: %v", ErrAvatarUnavailable, err)
	}
	defer object.Body.Close()
	data, err := io.ReadAll(io.LimitReader(object.Body, maxAvatarStoredBytes+1))
	if err != nil || len(data) > maxAvatarStoredBytes {
		return AvatarFile{}, ErrAvatarUnavailable
	}
	return AvatarFile{Data: data, ContentType: *u.AvatarContentType, ETag: object.ETag}, nil
}

func (s *Service) RemoveAvatar(ctx context.Context, id string) (UserResponse, error) {
	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return UserResponse{}, err
	}
	updated, err := s.repo.ClearAvatar(ctx, id)
	if err != nil {
		return UserResponse{}, err
	}
	if current.AvatarKey != nil {
		if err := s.avatarStore.Delete(ctx, *current.AvatarKey); err != nil {
			slog.Warn("failed to delete removed avatar", "key", *current.AvatarKey, "error", err)
		}
	}
	return toResponse(updated), nil
}

func (s *Service) UpdateRole(ctx context.Context, id, actorID string, role Role) (UserResponse, error) {
	if id == actorID {
		return UserResponse{}, ErrSelfRole
	}
	u, err := s.repo.UpdateRole(ctx, id, role)
	if err != nil {
		return UserResponse{}, err
	}
	return toResponse(u), nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}
	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}
	if current.AvatarKey != nil {
		if err := s.avatarStore.Delete(ctx, *current.AvatarKey); err != nil {
			slog.Warn("failed to delete avatar for removed user", "key", *current.AvatarKey, "error", err)
		}
	}
	return nil
}

func sanitizeAvatar(reader io.Reader, size int64) ([]byte, string, string, error) {
	if size > maxAvatarUploadBytes {
		return nil, "", "", ErrAvatarTooLarge
	}
	raw, err := io.ReadAll(io.LimitReader(reader, maxAvatarUploadBytes+1))
	if err != nil {
		return nil, "", "", ErrInvalidAvatar
	}
	if len(raw) == 0 {
		return nil, "", "", ErrInvalidAvatar
	}
	if len(raw) > maxAvatarUploadBytes {
		return nil, "", "", ErrAvatarTooLarge
	}

	contentType := http.DetectContentType(raw)
	if contentType != "image/jpeg" && contentType != "image/png" {
		return nil, "", "", ErrInvalidAvatar
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil || config.Width < 1 || config.Height < 1 || config.Width > maxAvatarDimension || config.Height > maxAvatarDimension {
		return nil, "", "", ErrInvalidAvatar
	}
	if (contentType == "image/jpeg" && format != "jpeg") || (contentType == "image/png" && format != "png") {
		return nil, "", "", ErrInvalidAvatar
	}

	decoded, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, "", "", ErrInvalidAvatar
	}
	var sanitized bytes.Buffer
	extension := "jpg"
	if contentType == "image/jpeg" {
		err = jpeg.Encode(&sanitized, decoded, &jpeg.Options{Quality: 85})
	} else {
		extension = "png"
		err = (&png.Encoder{CompressionLevel: png.BestSpeed}).Encode(&sanitized, decoded)
	}
	if err != nil || sanitized.Len() > maxAvatarStoredBytes {
		return nil, "", "", ErrInvalidAvatar
	}
	return sanitized.Bytes(), contentType, extension, nil
}
