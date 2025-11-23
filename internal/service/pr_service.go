package service

import (
	"avito-pr-service/internal/model"
	"avito-pr-service/internal/repository"
	"context"
)

type PRService struct {
	prRepo repository.PRRepository
}

func NewPRService(prRepo repository.PRRepository) *PRService {
	return &PRService{prRepo: prRepo}
}

func (s *PRService) CreatePR(ctx context.Context, req model.CreatePRRequest) (*model.PullRequest, error) {
	return s.prRepo.CreatePR(ctx, req.PRID, req.PRName, req.AuthorID)
}

func (s *PRService) MergePR(ctx context.Context, prID string) (*model.PullRequest, error) {
	return s.prRepo.MergePR(ctx, prID)
}

func (s *PRService) ReassignReviewer(ctx context.Context, prID, oldUserID string) (*model.PullRequest, string, error) {
	return s.prRepo.ReassignReviewer(ctx, prID, oldUserID)
}
