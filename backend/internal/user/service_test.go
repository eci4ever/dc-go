package user

import (
	"context"
	"errors"
	"testing"
)

func TestUpdateRoleRejectsSelfChangeBeforeRepositoryAccess(t *testing.T) {
	service := NewService(nil)
	_, err := service.UpdateRole(context.Background(), "same-user", "same-user", RoleAdmin)
	if !errors.Is(err, ErrSelfRole) {
		t.Fatalf("UpdateRole() error = %v, want ErrSelfRole", err)
	}
}
