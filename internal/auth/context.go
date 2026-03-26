package auth

import (
	"context"
	"errors"
)

type contextKey string

const (
	contextKeyUserID contextKey = "user_id"
)

var (
	ErrNoUserIDInContext      = errors.New("user id not in context")
	ErrInvalidUserIDInContext  = errors.New("user id in context has invalid type")
)

func UserIDFromContext(ctx context.Context) (string, error) {
	v := ctx.Value(contextKeyUserID)
	if v == nil {
		return "", ErrNoUserIDInContext
	}
	s, ok := v.(string)
	if !ok {
		return "", ErrInvalidUserIDInContext
	}
	return s, nil
}

func WithUserID(ctx context.Context, userID string) context.Context {
	return context.WithValue(ctx, contextKeyUserID, userID)
}
