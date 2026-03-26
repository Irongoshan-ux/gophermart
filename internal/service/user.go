// Package service implements business logic for gophermart.
package service

import (
	"context"
	"fmt"

	"gophermart/internal/model"
	"gophermart/internal/repository"
	"golang.org/x/crypto/bcrypt"
)

type UserService struct {
	users repository.UserRepository
}

func NewUserService(users repository.UserRepository) *UserService {
	return &UserService{users: users}
}

func (s *UserService) Register(ctx context.Context, login, password string) (*model.User, error) {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hash password: %w", err)
	}
	return s.users.Create(ctx, login, string(hash))
}

func (s *UserService) Login(ctx context.Context, login, password string) (*model.User, error) {
	u, err := s.users.GetByLogin(ctx, login)
	if err != nil {
		if err == repository.ErrNotFound {
			return nil, repository.ErrNotFound
		}
		return nil, err
	}
	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(password)); err != nil {
		return nil, repository.ErrNotFound
	}
	return u, nil
}
