package postgres

import (
	"avito-pr-service/internal/int_errors"
	"avito-pr-service/internal/model"
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Storage struct {
	pool *pgxpool.Pool
}

func (s *Storage) Pool() *pgxpool.Pool { return s.pool }
func (s *Storage) Close()              { s.pool.Close() }

func New(ctx context.Context, connStr string) (*Storage, error) {
	const op = "storage.postgres.New"

	config, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	config.MaxConns = 25
	config.MinConns = 2
	config.MaxConnLifetime = time.Hour
	config.HealthCheckPeriod = time.Minute

	pool, err := pgxpool.NewWithConfig(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("%s: %w", op, err)
	}

	if err = pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("%s: ping: %w", op, err)
	}

	return &Storage{pool: pool}, nil
}

type PRRepository struct {
	pool *pgxpool.Pool
}

func NewPRRepository(pool *pgxpool.Pool) *PRRepository {
	return &PRRepository{pool: pool}
}

func (r *PRRepository) CreatePR(ctx context.Context, prID, pullRequestName, authorID string) (*model.PullRequest, error) {
	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer tx.Rollback(ctx)

	var exists bool
	err = tx.QueryRow(ctx, "SELECT EXISTS(SELECT 1 FROM pull_requests WHERE pull_request_id = $1)", prID).Scan(&exists)
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, int_errors.ErrPRExists
	}

	var teamName string
	err = tx.QueryRow(ctx, "SELECT team_name FROM users WHERE user_id = $1", authorID).Scan(&teamName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, int_errors.ErrUserNotFound
		}
		return nil, err
	}

	createdAt := time.Now()
	_, err = tx.Exec(ctx, `
		INSERT INTO pull_requests (pull_request_id, pull_request_name, author_id, status, created_at)
		VALUES ($1, $2, $3, 'OPEN', $4)
	`, prID, pullRequestName, authorID, createdAt)
	if err != nil {
		return nil, err
	}

	rows, err := tx.Query(ctx, `
		SELECT user_id FROM users 
		WHERE team_name = $1 
		  AND is_active = true 
		  AND user_id != $2
		ORDER BY RANDOM() 
		LIMIT 2
	`, teamName, authorID)
	if err != nil {
		return nil, err
	}

	var reviewers []string
	for rows.Next() {
		var rID string
		if err := rows.Scan(&rID); err != nil {
			return nil, err
		}
		reviewers = append(reviewers, rID)
	}
	rows.Close()

	for _, revID := range reviewers {
		_, err := tx.Exec(ctx, `INSERT INTO pr_reviewers (pull_request_id, reviewer_id) VALUES ($1, $2)`, prID, revID)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, err
	}

	return &model.PullRequest{
		PRID:              prID,
		PRName:            pullRequestName,
		AuthorID:          authorID,
		Status:            "OPEN",
		AssignedReviewers: reviewers,
		CreatedAt:         createdAt,
	}, nil
}

func (r *PRRepository) MergePR(ctx context.Context, prID string) (*model.PullRequest, error) {
	row := r.pool.QueryRow(ctx, `
		UPDATE pull_requests
		SET status = 'MERGED',
		    merged_at = COALESCE(merged_at, NOW()) 
		WHERE pull_request_id = $1
		RETURNING pull_request_id, pull_request_name, author_id, status, created_at, merged_at
	`, prID)

	var pr model.PullRequest
	err := row.Scan(&pr.PRID, &pr.PRName, &pr.AuthorID, &pr.Status, &pr.CreatedAt, &pr.MergedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, int_errors.ErrPRNotFound
		}
		return nil, err
	}

	pr.AssignedReviewers, err = r.getReviewers(ctx, prID)
	if err != nil {
		return nil, err
	}

	return &pr, nil
}

func (r *PRRepository) ReassignReviewer(ctx context.Context, prID, oldUserID string) (*model.PullRequest, string, error) {
	const op = "PRRepository.ReassignReviewer"

	tx, err := r.pool.Begin(ctx)
	if err != nil {
		return nil, "", fmt.Errorf("%s: start transaction: %w", op, err)
	}
	defer tx.Rollback(ctx)

	var status, authorID, prName string
	var createdAt time.Time
	var mergedAt *time.Time
	err = tx.QueryRow(ctx, `SELECT status, author_id, pull_request_name, created_at, merged_at FROM pull_requests WHERE pull_request_id = $1 FOR UPDATE`, prID).Scan(&status, &authorID, &prName, &createdAt, &mergedAt)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", int_errors.ErrPRNotFound
		}
		return nil, "", fmt.Errorf("%s: select pr: %w", op, err)
	}

	if status == "MERGED" {
		return nil, "", int_errors.ErrPRMerged
	}

	res, err := tx.Exec(ctx, "DELETE FROM pr_reviewers WHERE pull_request_id = $1 AND reviewer_id = $2", prID, oldUserID)
	if err != nil {
		return nil, "", fmt.Errorf("%s: delete old reviewer: %w", op, err)
	}
	if res.RowsAffected() == 0 {
		return nil, "", int_errors.ErrReviewerNotAssigned
	}

	var teamName string
	err = tx.QueryRow(ctx, "SELECT team_name FROM users WHERE user_id = $1", oldUserID).Scan(&teamName)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", int_errors.ErrUserNotFound
		}
		return nil, "", fmt.Errorf("%s: select team name: %w", op, err)
	}

	currentReviewers, err := r.getReviewersTx(ctx, tx, prID)
	if err != nil {
		return nil, "", fmt.Errorf("%s: get current reviewers: %w", op, err)
	}

	excludeIDs := []string{authorID, oldUserID}
	excludeIDs = append(excludeIDs, currentReviewers...)

	var newReviewerID string
	err = tx.QueryRow(ctx, `
		SELECT user_id FROM users
		WHERE team_name = $1
		  AND is_active = true
		  AND user_id != ALL($2)
		ORDER BY RANDOM()
		LIMIT 1
	`, teamName, excludeIDs).Scan(&newReviewerID)

	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, "", int_errors.ErrNoReplacementCandidate
		}
		return nil, "", fmt.Errorf("%s: select replacement: %w", op, err)
	}

	_, err = tx.Exec(ctx, "INSERT INTO pr_reviewers (pull_request_id, reviewer_id) VALUES ($1, $2)", prID, newReviewerID)
	if err != nil {
		return nil, "", fmt.Errorf("%s: insert new reviewer: %w", op, err)
	}

	if err := tx.Commit(ctx); err != nil {
		return nil, "", fmt.Errorf("%s: commit: %w", op, err)
	}

	reviewersAfter := append(currentReviewers, newReviewerID)

	pr := &model.PullRequest{
		PRID:              prID,
		PRName:            prName,
		AuthorID:          authorID,
		Status:            status,
		AssignedReviewers: reviewersAfter,
		CreatedAt:         createdAt,
		MergedAt:          mergedAt,
	}

	return pr, newReviewerID, nil
}

func (r *PRRepository) getReviewers(ctx context.Context, prID string) ([]string, error) {
	rows, err := r.pool.Query(ctx, "SELECT reviewer_id FROM pr_reviewers WHERE pull_request_id = $1", prID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		rows.Scan(&id)
		ids = append(ids, id)
	}
	return ids, nil
}

func (r *PRRepository) getReviewersTx(ctx context.Context, tx pgx.Tx, prID string) ([]string, error) {
	rows, err := tx.Query(ctx, "SELECT reviewer_id FROM pr_reviewers WHERE pull_request_id = $1", prID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ids []string
	for rows.Next() {
		var id string
		rows.Scan(&id)
		ids = append(ids, id)
	}
	return ids, nil
}
