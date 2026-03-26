package service

import (
	"context"
	"testing"
	"time"

	"gophermart/internal/model"
	"gophermart/internal/repository"
	"gophermart/internal/repository/mocks"
	"go.uber.org/mock/gomock"
	"golang.org/x/crypto/bcrypt"
)

func TestUserService_Register(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := repository_mocks.NewMockUserRepository(ctrl)
	svc := NewUserService(repo)

	ctx := context.Background()
	repo.EXPECT().
		Create(ctx, "user1", gomock.Any()).
		Return(&model.User{ID: "id1", Login: "user1", PasswordHash: "hash", CreatedAt: time.Now()}, nil)

	user, err := svc.Register(ctx, "user1", "pass")
	if err != nil {
		t.Fatalf("Register: %v", err)
	}
	if user.Login != "user1" || user.ID != "id1" {
		t.Errorf("got user %+v", user)
	}
}

func TestUserService_Register_Conflict(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := repository_mocks.NewMockUserRepository(ctrl)
	svc := NewUserService(repo)

	ctx := context.Background()
	repo.EXPECT().
		Create(ctx, "user1", gomock.Any()).
		Return(nil, repository.ErrConflict)

	_, err := svc.Register(ctx, "user1", "pass")
	if err != repository.ErrConflict {
		t.Errorf("got err %v", err)
	}
}

func TestUserService_Login(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := repository_mocks.NewMockUserRepository(ctrl)
	svc := NewUserService(repo)

	ctx := context.Background()
	hash, _ := bcrypt.GenerateFromPassword([]byte("pass"), bcrypt.DefaultCost)
	repo.EXPECT().
		GetByLogin(ctx, "gooduser").
		Return(&model.User{ID: "id1", Login: "gooduser", PasswordHash: string(hash), CreatedAt: time.Now()}, nil)

	user, err := svc.Login(ctx, "gooduser", "pass")
	if err != nil {
		t.Fatalf("Login: %v", err)
	}
	if user.Login != "gooduser" {
		t.Errorf("got %q", user.Login)
	}
}

func TestUserService_Login_WrongPassword(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := repository_mocks.NewMockUserRepository(ctrl)
	svc := NewUserService(repo)

	ctx := context.Background()
	hash, _ := bcrypt.GenerateFromPassword([]byte("right"), bcrypt.DefaultCost)
	repo.EXPECT().
		GetByLogin(ctx, "user1").
		Return(&model.User{ID: "id1", Login: "user1", PasswordHash: string(hash), CreatedAt: time.Now()}, nil)

	_, err := svc.Login(ctx, "user1", "wrong")
	if err != repository.ErrNotFound {
		t.Errorf("got err %v", err)
	}
}

func TestUserService_Login_UserNotFound(t *testing.T) {
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	repo := repository_mocks.NewMockUserRepository(ctrl)
	svc := NewUserService(repo)

	ctx := context.Background()
	repo.EXPECT().
		GetByLogin(ctx, "nobody").
		Return(nil, repository.ErrNotFound)

	_, err := svc.Login(ctx, "nobody", "pass")
	if err != repository.ErrNotFound {
		t.Errorf("got err %v", err)
	}
}
