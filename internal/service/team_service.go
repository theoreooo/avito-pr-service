package service

import (
	"avito-pr-service/internal/model"
	"avito-pr-service/internal/repository"
	"context"
)

type TeamService struct {
	teamRepo repository.TeamRepository
	userRepo repository.UserRepository
	prRepo   repository.PRRepository
}

func NewTeamService(teamRepo repository.TeamRepository, userRepo repository.UserRepository, prRepo repository.PRRepository) *TeamService {
	return &TeamService{
		teamRepo: teamRepo,
		userRepo: userRepo,
		prRepo:   prRepo,
	}
}

func (s *TeamService) GetTeam(ctx context.Context, name string) (*model.Team, error) {
	return s.teamRepo.GetTeam(ctx, name)
}

func (s *TeamService) CreateTeam(ctx context.Context, team *model.Team) (*model.Team, error) {
	return s.teamRepo.AddTeam(ctx, team)
}
