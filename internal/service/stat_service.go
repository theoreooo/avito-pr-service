package service

import (
	"context"

	"avito-pr-service/internal/model"
	"avito-pr-service/internal/repository"
)

type StatisticsService struct {
	repo repository.StatisticsRepository
}

func NewStatisticsService(repo repository.StatisticsRepository) *StatisticsService {
	return &StatisticsService{repo: repo}
}

func (s *StatisticsService) GetStatistics(ctx context.Context) (*model.Statistics, error) {
	return s.repo.GetStatistics(ctx)
}
