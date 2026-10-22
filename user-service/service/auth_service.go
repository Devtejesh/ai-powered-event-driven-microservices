package service

import (
	"context"
	"fmt"

	"example.com/go-commerce-platform/user-service/models"
	"example.com/go-commerce-platform/user-service/repository"

	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	Repo *repository.UserReppository
}

func NewAuthService(repo *repository.UserReppository) *AuthService {
	return &AuthService{
		Repo: repo,
	}
}

func (s *AuthService) RegisterUser(ctx context.Context, name string, email string, password string) error {
	if name == "" {
		return fmt.Errorf("name is required")
	}
	if email == "" {
		return fmt.Errorf("email is required")
	}
	if password == "" {
		return fmt.Errorf("password is required")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &models.User{
		Name:         name,
		Email:        email,
		PasswordHash: string(hashedPassword),
	}
	return s.Repo.CreateUser(ctx, user)
}

func (s *AuthService) LoginUser(ctx context.Context, email string, password string) (string, error) {
	if email == "" {
		return "", fmt.Errorf("email is required")
	}
	if password == "" {
		return "", fmt.Errorf("password is required	")
	}
	user, err := s.Repo.GetUserByEmail(ctx, email)
	if err != nil {
		return "", fmt.Errorf("invalid email or password")
	}
	err = bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(password))
	if err != nil {
		return "", fmt.Errorf("invalid email or password")
	}
	token, err := GenerateToken(user.ID)
	if err != nil {
		return "", fmt.Errorf("failed to generate token: %w", err)
	}
	return token, nil
}
