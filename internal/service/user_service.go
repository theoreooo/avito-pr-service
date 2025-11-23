package service

import (
	"avito-pr-service/internal/model"
	"avito-pr-service/internal/repository"
	"context"
)

type UserService struct {
	userRepo repository.UserRepository
}

func NewUserService(userRepo repository.UserRepository) *UserService {
	return &UserService{userRepo: userRepo}
}

func (s *UserService) GetByReviewer(ctx context.Context, id string) ([]model.PullRequest, error) {
	return s.userRepo.GetUserReviews(ctx, id)
}

func (s *UserService) SetIsActive(ctx context.Context, userID string, isActive bool) (*model.User, error) {
	return s.userRepo.SetIsActive(ctx, userID, isActive)
}
