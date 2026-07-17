package auth

import "testing"

func TestJWTSignAndVerify(t *testing.T) {
	service := NewJWTService("A-Strong-Secret-With-Upper-Lower-123!", "dc-go", "dc-go")
	token, err := service.Sign("user-id")
	if err != nil {
		t.Fatal(err)
	}
	claims, err := service.Verify(token)
	if err != nil {
		t.Fatal(err)
	}
	if claims.Subject != "user-id" {
		t.Fatalf("unexpected claims: %+v", claims)
	}
}

func TestJWTVerifyRejectsWrongContext(t *testing.T) {
	signer := NewJWTService("A-Strong-Secret-With-Upper-Lower-123!", "dc-go", "dc-go")
	token, err := signer.Sign("user-id")
	if err != nil {
		t.Fatal(err)
	}
	for _, verifier := range []*JWTService{
		NewJWTService("A-Different-Strong-Secret-Value-456!", "dc-go", "dc-go"),
		NewJWTService("A-Strong-Secret-With-Upper-Lower-123!", "other", "dc-go"),
		NewJWTService("A-Strong-Secret-With-Upper-Lower-123!", "dc-go", "other"),
	} {
		if _, err := verifier.Verify(token); err == nil {
			t.Fatal("Verify() should reject a mismatched token context")
		}
	}
}

func TestRefreshTokenGenerationAndHashing(t *testing.T) {
	first, err := newRefreshToken()
	if err != nil {
		t.Fatal(err)
	}
	second, err := newRefreshToken()
	if err != nil {
		t.Fatal(err)
	}
	if first == second {
		t.Fatal("refresh tokens must be unique")
	}
	if hashToken(first) == first || len(hashToken(first)) != 64 {
		t.Fatal("refresh tokens must be stored as SHA-256 hashes")
	}
	if hashToken(first) != hashToken(first) {
		t.Fatal("token hashing must be deterministic")
	}
}
