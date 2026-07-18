package organization

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

	"github.com/google/uuid"
)

const (
	maxLogoUploadBytes = 2 << 20
	maxLogoStoredBytes = 5 << 20
	maxLogoDimension   = 2048
)

type LogoFile struct {
	Data        []byte
	ContentType string
	ETag        string
}

func (s *Service) UploadLogo(ctx context.Context, id string, reader io.Reader, size int64) (Organization, error) {
	current, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return Organization{}, err
	}

	data, contentType, extension, err := sanitizeLogo(reader, size)
	if err != nil {
		return Organization{}, err
	}
	key := fmt.Sprintf("organization-logos/%s/%s.%s", id, uuid.NewString(), extension)
	if err := s.logoStore.Put(ctx, key, contentType, bytes.NewReader(data), int64(len(data))); err != nil {
		slog.Warn("organization logo upload failed", "key", key, "error", err)
		return Organization{}, fmt.Errorf("%w: %v", ErrLogoUnavailable, err)
	}

	updatedAt := time.Now().UTC()
	logoURL := fmt.Sprintf("/api/v1/organizations/%s/logo?v=%d", id, updatedAt.UnixNano())
	updated, err := s.repo.UpdateLogo(ctx, id, logoURL, key, contentType, updatedAt)
	if err != nil {
		if cleanupErr := s.logoStore.Delete(ctx, key); cleanupErr != nil {
			slog.Warn("failed to clean up unreferenced organization logo", "key", key, "error", cleanupErr)
		}
		return Organization{}, err
	}

	if current.LogoKey != nil && *current.LogoKey != key {
		if err := s.logoStore.Delete(ctx, *current.LogoKey); err != nil {
			slog.Warn("failed to delete replaced organization logo", "key", *current.LogoKey, "error", err)
		}
	}
	return updated, nil
}

func (s *Service) GetLogo(ctx context.Context, id string) (LogoFile, error) {
	org, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return LogoFile{}, err
	}
	if org.LogoKey == nil || org.LogoContentType == nil {
		return LogoFile{}, ErrNoLogo
	}

	object, err := s.logoStore.Get(ctx, *org.LogoKey)
	if err != nil {
		slog.Warn("organization logo download failed", "key", *org.LogoKey, "error", err)
		return LogoFile{}, fmt.Errorf("%w: %v", ErrLogoUnavailable, err)
	}
	defer object.Body.Close()
	data, err := io.ReadAll(io.LimitReader(object.Body, maxLogoStoredBytes+1))
	if err != nil || len(data) > maxLogoStoredBytes {
		return LogoFile{}, ErrLogoUnavailable
	}
	return LogoFile{Data: data, ContentType: *org.LogoContentType, ETag: object.ETag}, nil
}

func sanitizeLogo(reader io.Reader, size int64) ([]byte, string, string, error) {
	if size > maxLogoUploadBytes {
		return nil, "", "", ErrLogoTooLarge
	}
	raw, err := io.ReadAll(io.LimitReader(reader, maxLogoUploadBytes+1))
	if err != nil || len(raw) == 0 {
		return nil, "", "", ErrInvalidLogo
	}
	if len(raw) > maxLogoUploadBytes {
		return nil, "", "", ErrLogoTooLarge
	}

	contentType := http.DetectContentType(raw)
	if contentType != "image/jpeg" && contentType != "image/png" {
		return nil, "", "", ErrInvalidLogo
	}
	config, format, err := image.DecodeConfig(bytes.NewReader(raw))
	if err != nil || config.Width < 1 || config.Height < 1 || config.Width > maxLogoDimension || config.Height > maxLogoDimension {
		return nil, "", "", ErrInvalidLogo
	}
	if (contentType == "image/jpeg" && format != "jpeg") || (contentType == "image/png" && format != "png") {
		return nil, "", "", ErrInvalidLogo
	}

	decoded, _, err := image.Decode(bytes.NewReader(raw))
	if err != nil {
		return nil, "", "", ErrInvalidLogo
	}
	var sanitized bytes.Buffer
	extension := "jpg"
	if contentType == "image/jpeg" {
		err = jpeg.Encode(&sanitized, decoded, &jpeg.Options{Quality: 85})
	} else {
		extension = "png"
		err = (&png.Encoder{CompressionLevel: png.BestSpeed}).Encode(&sanitized, decoded)
	}
	if err != nil || sanitized.Len() > maxLogoStoredBytes {
		return nil, "", "", ErrInvalidLogo
	}
	return sanitized.Bytes(), contentType, extension, nil
}
