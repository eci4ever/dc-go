package user

import (
	"bytes"
	"context"
	"errors"
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"io"
	"strings"
	"testing"
	"time"

	"github.com/eci4ever/dc-go/internal/storage"
)

type fakeUserRepository struct {
	user            User
	updateAvatarErr error
	clearAvatarErr  error
}

func (r *fakeUserRepository) GetByID(context.Context, string) (User, error) { return r.user, nil }
func (r *fakeUserRepository) List(context.Context) ([]User, error)          { return []User{r.user}, nil }
func (r *fakeUserRepository) Update(_ context.Context, _ string, name string) (User, error) {
	r.user.Name = name
	return r.user, nil
}
func (r *fakeUserRepository) UpdateAvatar(_ context.Context, _, imageURL, key, contentType string, updatedAt time.Time) (User, error) {
	if r.updateAvatarErr != nil {
		return User{}, r.updateAvatarErr
	}
	r.user.Image = &imageURL
	r.user.AvatarKey = &key
	r.user.AvatarContentType = &contentType
	r.user.AvatarUpdatedAt = &updatedAt
	return r.user, nil
}
func (r *fakeUserRepository) ClearAvatar(context.Context, string) (User, error) {
	if r.clearAvatarErr != nil {
		return User{}, r.clearAvatarErr
	}
	r.user.Image = nil
	r.user.AvatarKey = nil
	r.user.AvatarContentType = nil
	r.user.AvatarUpdatedAt = nil
	return r.user, nil
}
func (r *fakeUserRepository) UpdateRole(_ context.Context, _ string, role Role) (User, error) {
	r.user.Role = role
	return r.user, nil
}
func (r *fakeUserRepository) Delete(context.Context, string) error { return nil }

type fakeObjectStore struct {
	putKey      string
	putType     string
	putData     []byte
	putErr      error
	getObject   storage.Object
	getErr      error
	deletedKeys []string
	deleteErr   error
}

func (s *fakeObjectStore) Put(_ context.Context, key, contentType string, body io.Reader, _ int64) error {
	if s.putErr != nil {
		return s.putErr
	}
	s.putKey = key
	s.putType = contentType
	s.putData, _ = io.ReadAll(body)
	return nil
}
func (s *fakeObjectStore) Get(context.Context, string) (storage.Object, error) {
	return s.getObject, s.getErr
}
func (s *fakeObjectStore) Delete(_ context.Context, key string) error {
	s.deletedKeys = append(s.deletedKeys, key)
	return s.deleteErr
}

func TestUploadAvatarSanitizesAndReplacesObject(t *testing.T) {
	oldKey := "avatars/user-id/old.png"
	repo := &fakeUserRepository{user: User{ID: "user-id", Name: "User", AvatarKey: &oldKey}}
	store := &fakeObjectStore{}
	service := NewService(repo, store)
	imageData := testPNG(t, 16, 16)

	result, err := service.UploadAvatar(context.Background(), "user-id", bytes.NewReader(imageData), int64(len(imageData)))
	if err != nil {
		t.Fatal(err)
	}
	if store.putType != "image/png" || !strings.HasPrefix(store.putKey, "avatars/user-id/") || !strings.HasSuffix(store.putKey, ".png") {
		t.Fatalf("unexpected stored avatar: key=%q type=%q", store.putKey, store.putType)
	}
	if _, _, err := image.Decode(bytes.NewReader(store.putData)); err != nil {
		t.Fatalf("stored avatar is not decodable: %v", err)
	}
	if result.Image == nil || !strings.HasPrefix(*result.Image, "/api/v1/users/user-id/avatar?v=") {
		t.Fatalf("unexpected image URL: %v", result.Image)
	}
	if len(store.deletedKeys) != 1 || store.deletedKeys[0] != oldKey {
		t.Fatalf("replaced avatar was not deleted: %v", store.deletedKeys)
	}
}

func TestUploadAvatarDeletesNewObjectWhenDatabaseUpdateFails(t *testing.T) {
	repo := &fakeUserRepository{user: User{ID: "user-id"}, updateAvatarErr: errors.New("database unavailable")}
	store := &fakeObjectStore{}
	service := NewService(repo, store)
	imageData := testPNG(t, 4, 4)

	if _, err := service.UploadAvatar(context.Background(), "user-id", bytes.NewReader(imageData), int64(len(imageData))); err == nil {
		t.Fatal("expected upload to fail")
	}
	if len(store.deletedKeys) != 1 || store.deletedKeys[0] != store.putKey {
		t.Fatalf("new object was not cleaned up: put=%q deleted=%v", store.putKey, store.deletedKeys)
	}
}

func TestSanitizeAvatarRejectsInvalidInput(t *testing.T) {
	if _, _, _, err := sanitizeAvatar(bytes.NewReader(nil), 0); !errors.Is(err, ErrInvalidAvatar) {
		t.Fatalf("empty file error = %v", err)
	}
	if _, _, _, err := sanitizeAvatar(strings.NewReader("not an image"), 12); !errors.Is(err, ErrInvalidAvatar) {
		t.Fatalf("invalid file error = %v", err)
	}
	if _, _, _, err := sanitizeAvatar(bytes.NewReader(nil), maxAvatarUploadBytes+1); !errors.Is(err, ErrAvatarTooLarge) {
		t.Fatalf("oversized file error = %v", err)
	}
	largeDimensions := testPNG(t, maxAvatarDimension+1, 1)
	if _, _, _, err := sanitizeAvatar(bytes.NewReader(largeDimensions), int64(len(largeDimensions))); !errors.Is(err, ErrInvalidAvatar) {
		t.Fatalf("large dimensions error = %v", err)
	}
}

func TestSanitizeAvatarSupportsJPEG(t *testing.T) {
	img := image.NewRGBA(image.Rect(0, 0, 8, 8))
	img.Set(0, 0, color.RGBA{R: 120, G: 80, B: 200, A: 255})
	var input bytes.Buffer
	if err := jpeg.Encode(&input, img, nil); err != nil {
		t.Fatal(err)
	}

	data, contentType, extension, err := sanitizeAvatar(bytes.NewReader(input.Bytes()), int64(input.Len()))
	if err != nil {
		t.Fatal(err)
	}
	if contentType != "image/jpeg" || extension != "jpg" {
		t.Fatalf("unexpected JPEG metadata: contentType=%q extension=%q", contentType, extension)
	}
	if _, format, err := image.Decode(bytes.NewReader(data)); err != nil || format != "jpeg" {
		t.Fatalf("stored JPEG is not decodable: format=%q error=%v", format, err)
	}
}

func TestGetAndRemoveAvatar(t *testing.T) {
	key := "avatars/user-id/avatar.png"
	contentType := "image/png"
	repo := &fakeUserRepository{user: User{ID: "user-id", AvatarKey: &key, AvatarContentType: &contentType}}
	store := &fakeObjectStore{getObject: storage.Object{
		Body:        io.NopCloser(strings.NewReader("image-data")),
		ContentType: contentType,
		ETag:        "etag-value",
	}}
	service := NewService(repo, store)

	avatar, err := service.GetAvatar(context.Background(), "user-id")
	if err != nil || string(avatar.Data) != "image-data" || avatar.ETag != "etag-value" {
		t.Fatalf("GetAvatar() = %+v, %v", avatar, err)
	}
	result, err := service.RemoveAvatar(context.Background(), "user-id")
	if err != nil || result.Image != nil {
		t.Fatalf("RemoveAvatar() = %+v, %v", result, err)
	}
	if len(store.deletedKeys) != 1 || store.deletedKeys[0] != key {
		t.Fatalf("removed avatar was not deleted: %v", store.deletedKeys)
	}
}

func testPNG(t *testing.T, width, height int) []byte {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	img.Set(0, 0, color.RGBA{R: 120, G: 80, B: 200, A: 255})
	var data bytes.Buffer
	if err := png.Encode(&data, img); err != nil {
		t.Fatal(err)
	}
	return data.Bytes()
}
