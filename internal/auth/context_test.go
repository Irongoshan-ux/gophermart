package auth

import (
	"context"
	"testing"
)

func TestUserIDFromContext(t *testing.T) {
	ctx := context.Background()
	_, err := UserIDFromContext(ctx)
	if err != ErrNoUserIDInContext {
		t.Errorf("empty context: got %v", err)
	}
	ctx = WithUserID(ctx, "user-123")
	id, err := UserIDFromContext(ctx)
	if err != nil {
		t.Fatalf("WithUserID then FromContext: %v", err)
	}
	if id != "user-123" {
		t.Errorf("got %q, want user-123", id)
	}
}
