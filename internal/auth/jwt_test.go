package auth

import "testing"

func TestNewTokenAndParseToken(t *testing.T) {
	secret := "test-secret"
	userID := "550e8400-e29b-41d4-a716-446655440000"
	token, err := NewToken(secret, userID)
	if err != nil {
		t.Fatalf("NewToken: %v", err)
	}
	if token == "" {
		t.Fatal("token is empty")
	}
	parsed, err := ParseToken(secret, token)
	if err != nil {
		t.Fatalf("ParseToken: %v", err)
	}
	if parsed != userID {
		t.Errorf("ParseToken got userID %q, want %q", parsed, userID)
	}
}

func TestParseToken_Invalid(t *testing.T) {
	_, err := ParseToken("secret", "invalid.jwt.here")
	if err != ErrInvalidToken {
		t.Errorf("ParseToken(invalid) = %v, want ErrInvalidToken", err)
	}
}

func TestParseToken_WrongSecret(t *testing.T) {
	token, _ := NewToken("secret1", "user-id")
	_, err := ParseToken("secret2", token)
	if err != ErrInvalidToken {
		t.Errorf("ParseToken(wrong secret) = %v, want ErrInvalidToken", err)
	}
}
