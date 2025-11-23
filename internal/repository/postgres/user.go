package postgres

import (
	"avito-pr-service/internal/int_errors"
	"avito-pr-service/internal/model"
	"context"
	"errors"
	"fmt"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type UserRepository struct {
	pool *pgxpool.Pool
}

func NewUserRepository(pool *pgxpool.Pool) *UserRepository {
	return &UserRepository{pool: pool}
}

func (r *UserRepository) GetUserReviews(ctx context.Context, userID string) ([]model.PullRequest, error) {
	const op = "UserRepository.GetUserReviews"

	query := `
		SELECT 
			pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status
		FROM pull_requests pr
		JOIN pr_reviewers prr ON pr.pull_request_id = prr.pull_request_id
		WHERE prr.reviewer_id = $1
	`

	rows, err := r.pool.Query(ctx, query, userID)
	if err != nil {
		return nil, fmt.Errorf("%s: query failed: %w", op, err)
	}
	defer rows.Close()

	var prs []model.PullRequest
	for rows.Next() {
		var pr model.PullRequest

		if err := rows.Scan(&pr.PRID, &pr.PRName, &pr.AuthorID, &pr.Status); err != nil {
			return nil, fmt.Errorf("%s: scan failed: %w", op, err)
		}

		prs = append(prs, pr)
	}
	
	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("%s: iteration error: %w", op, err)
	}

	return prs, nil
}

func (r *UserRepository) SetIsActive(ctx context.Context, userID string, isActive bool) (*model.User, error) {
	const op = "UserRepository.SetIsActive"

	row := r.pool.QueryRow(ctx, `
		UPDATE users
		SET is_active = $2
		WHERE user_id = $1
		RETURNING user_id, username, team_name, is_active
	`, userID, isActive)

	var user model.User

	err := row.Scan(&user.ID, &user.Username, &user.TeamName, &user.IsActive)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, int_errors.ErrUserNotFound
		}
		return nil, fmt.Errorf("%s: update query failed: %w", op, err)
	}

	return &user, nil
}
