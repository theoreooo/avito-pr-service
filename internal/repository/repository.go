package repository

import (
	"avito-pr-service/internal/model"
	"context"
)

type PRRepository interface {
	CreatePR(ctx context.Context, prID, title, authorID string) (*model.PullRequest, error)
	MergePR(ctx context.Context, prID string) (*model.PullRequest, error)
	ReassignReviewer(ctx context.Context, prID, oldUserID string) (*model.PullRequest, string, error)
}

type UserRepository interface {
	GetUserReviews(ctx context.Context, userID string) ([]model.PullRequest, error)
	SetIsActive(ctx context.Context, userID string, isActive bool) (*model.User, error)
}

type TeamRepository interface {
	GetTeam(ctx context.Context, teamName string) (*model.Team, error)
	AddTeam(ctx context.Context, team *model.Team) (*model.Team, error)
}

type StatisticsRepository interface {
	GetStatistics(ctx context.Context) (*model.Statistics, error)
}
