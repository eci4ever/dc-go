package auth

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"dc-express/internal/user"
)

func TestSessionResponseJSONContract(t *testing.T) {
	response := SessionResponse{
		Session: AuthSession{
			ID:        "session-id",
			ExpiresAt: "2026-07-17T14:29:49.778Z",
			CreatedAt: "2026-07-10T14:29:49.778Z",
			UpdatedAt: "2026-07-10T14:29:49.778Z",
			UserID:    "user-id",
		},
		User: AuthUser{
			ID:               "user-id",
			Name:             "User",
			Email:            "user@example.com",
			Role:             user.RoleUser,
			TwoFactorEnabled: false,
		},
	}

	encoded, err := json.Marshal(response)
	if err != nil {
		t.Fatal(err)
	}
	jsonText := string(encoded)
	for _, field := range []string{
		`"activeOrganizationId":null`,
		`"activeOrganizationRole":null`,
		`"activeTeamId":null`,
		`"ipAddress":null`,
		`"banReason":null`,
		`"twoFactorEnabled":false`,
		`"emailVerified":false`,
	} {
		if !strings.Contains(jsonText, field) {
			t.Errorf("expected JSON to contain %s: %s", field, jsonText)
		}
	}
	if strings.Contains(strings.ToLower(jsonText), "token") {
		t.Fatalf("session response must not expose a token: %s", jsonText)
	}
}

func TestFormatTimeUsesUTCAndRFC3339(t *testing.T) {
	value := time.Date(2026, 7, 10, 22, 29, 49, 778000000, time.FixedZone("MYT", 8*60*60))
	if got, want := formatTime(value), "2026-07-10T14:29:49.778Z"; got != want {
		t.Fatalf("formatTime() = %q, want %q", got, want)
	}
}
