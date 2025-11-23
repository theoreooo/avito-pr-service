package postgres

import (
	"avito-pr-service/internal/int_errors"
	"avito-pr-service/internal/model"
	"context"
	"fmt"

	"github.com/jackc/pgx/v5/pgxpool"
)

type TeamRepository struct {
	pool     *pgxpool.Pool
	userRepo *UserRepository
}

func NewTeamRepository(pool *pgxpool.Pool, userRepo *UserRepository) *TeamRepository {
	return &TeamRepository{pool: pool, userRepo: userRepo}
}

func (r *TeamRepository) GetTeam(ctx context.Context, teamName string) (*model.Team, error) {
	const op = "TeamRepository.GetTeam"

	rows, err := r.pool.Query(ctx, `
		SELECT user_id, username, is_active
		FROM users
		WHERE team_name = $1
		ORDER BY username
	`, teamName)
	if err != nil {
		return nil, fmt.Errorf("%s: query failed: %w", op, err)
	}
	defer rows.Close()

	var members []*model.User
	for rows.Next() {
		var member model.User
		if err := rows.Scan(&member.ID, &member.Username, &member.IsActive); err != nil {
			return nil, fmt.Errorf("%s: scan failed: %w", op, err)
		}
		members = append(members, &member)
	}

	if len(members) == 0 {
		return nil, int_errors.ErrTeamNotFound
	}

	team := &model.Team{
		TeamName: teamName,
		Members:  members,
	}

	return team, nil
}

func (r *TeamRepository) AddTeam(ctx context.Context, team *model.Team) (*model.Team, error) {
	const op = "TeamsRepository.AddTeam"

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, fmt.Errorf("%s: start transaction: %w", op, err)
	}
	defer tx.Rollback(ctx)

	// Insert or ignore if already exists
	_, err = tx.Exec(ctx, `INSERT INTO teams (team_name) VALUES ($1) ON CONFLICT (team_name) DO NOTHING`, team.TeamName)
	if err != nil {
		return nil, fmt.Errorf("%s: insert team: %w", op, err)
	}

	for _, member := range team.Members {
		_, err := tx.Exec(ctx, `
			INSERT INTO users (user_id, username, team_name, is_active)
			VALUES ($1, $2, $3, $4)
			ON CONFLICT (user_id) 
			DO UPDATE SET 
				username = EXCLUDED.username, 
				team_name = EXCLUDED.team_name, 
				is_active = EXCLUDED.is_active
		`, member.ID, member.Username, team.TeamName, member.IsActive)

		if err != nil {
			return nil, fmt.Errorf("%s: upsert user %s: %w", op, member.ID, err)
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, fmt.Errorf("%s: commit: %w", op, err)
	}

	return team, nil
}
