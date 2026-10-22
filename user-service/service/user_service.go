package service

import (
	"example.com/go-commerce-platform/user-service/models"
	"example.com/go-commerce-platform/user-service/repository"
	"context"
	"errors"
	"fmt"
)

type UserService struct {
	Repo *repository.UserReppository
}

func NewUserService(repo *repository.UserReppository) *UserService {
	return &UserService{
		Repo: repo,
	}
}

func (s *UserService) CreateUser(ctx context.Context, user *models.User) error {
	if user.Name == "" {
		return errors.New("name is required")
	}
	if user.Email == "" {
		return errors.New("email is required")
	}
	return s.Repo.CreateUser(ctx, user)
}

func (s *UserService) GetUserById(ctx context.Context, id int) (*models.User, error) {
	if id <= 0 {
		return nil, errors.New("invalid user id")
	}
	return s.Repo.GetUserById(ctx, id)
}
func (s *UserService) DeleteUserBYId(ctx context.Context, id int) error {
	if id <= 0 {
		return errors.New("invalid user id")
	}
	return s.Repo.DeleteUserBYId(ctx, id)
}
func (s *UserService) GetAllUsers(
	ctx context.Context, name string, page int, limit int) ([]*models.User, error) {
	if page == 0 {
		page = 1
	}
	if limit == 0 {
		limit = 10
	}
	if page < 1 {
		return nil, fmt.Errorf("page must be greater than 0")
	}
	if limit < 1 || limit > 100 {
		return nil, fmt.Errorf("limit must be between 1 and 100")
	}
	offset := (page - 1) * limit
	users, err := s.Repo.GetAllUsers(ctx, name, limit, offset)
	if err != nil {
		return nil, err
	}
	return users, nil
}
