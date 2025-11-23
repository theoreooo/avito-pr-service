package postgres

import (
	"context"
	"fmt"

	"avito-pr-service/internal/model"

	"github.com/jackc/pgx/v5/pgxpool"
)

type StatisticsRepository struct {
	db *pgxpool.Pool
}

func NewStatisticsRepository(db *pgxpool.Pool) *StatisticsRepository {
	return &StatisticsRepository{db: db}
}

func (r *StatisticsRepository) GetStatistics(ctx context.Context) (*model.Statistics, error) {
	const op = "StatisticsRepository.GetStatistics"

	stats := &model.Statistics{
		ReviewersStats: []model.ReviewerStats{},
		PRStats:        []model.PRStats{},
	}

	reviewersStatsRows, err := r.db.Query(ctx, `
        SELECT 
            u.user_id,
            u.username,
            u.team_name,
            COUNT(DISTINCT pr.pull_request_id) as total_reviews,
            COUNT(DISTINCT CASE WHEN pr.status = 'OPEN' THEN pr.pull_request_id END) as open_reviews,
            COUNT(DISTINCT CASE WHEN pr.status = 'MERGED' THEN pr.pull_request_id END) as merged_reviews
        FROM users u
        LEFT JOIN pr_reviewers prr ON u.user_id = prr.reviewer_id
        LEFT JOIN pull_requests pr ON prr.pull_request_id = pr.pull_request_id
        GROUP BY u.user_id, u.username, u.team_name
        ORDER BY total_reviews DESC
    `)
	if err != nil {
		return nil, fmt.Errorf("%s: query reviewers stats: %w", op, err)
	}
	defer reviewersStatsRows.Close()

	for reviewersStatsRows.Next() {
		var rs model.ReviewerStats
		err := reviewersStatsRows.Scan(
			&rs.UserID,
			&rs.Username,
			&rs.TeamName,
			&rs.TotalReviews,
			&rs.OpenReviews,
			&rs.MergedReviews,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: scan reviewers stats: %w", op, err)
		}
		stats.ReviewersStats = append(stats.ReviewersStats, rs)
	}

	prStatsRows, err := r.db.Query(ctx, `
        SELECT 
            pr.pull_request_id,
            pr.pull_request_name,
            pr.author_id,
            pr.status,
            COUNT(prr.reviewer_id) as reviewers_count
        FROM pull_requests pr
        LEFT JOIN pr_reviewers prr ON pr.pull_request_id = prr.pull_request_id
        GROUP BY pr.pull_request_id, pr.pull_request_name, pr.author_id, pr.status
        ORDER BY pr.created_at DESC
    `)
	if err != nil {
		return nil, fmt.Errorf("%s: query pr stats: %w", op, err)
	}
	defer prStatsRows.Close()

	for prStatsRows.Next() {
		var ps model.PRStats
		err := prStatsRows.Scan(
			&ps.PullRequestID,
			&ps.PullRequestName,
			&ps.AuthorID,
			&ps.Status,
			&ps.ReviewersCount,
		)
		if err != nil {
			return nil, fmt.Errorf("%s: scan pr stats: %w", op, err)
		}
		stats.PRStats = append(stats.PRStats, ps)
	}

	var totalPRs, totalReviewers int
	err = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM pull_requests`).Scan(&totalPRs)
	if err != nil {
		return nil, fmt.Errorf("%s: query total prs: %w", op, err)
	}

	err = r.db.QueryRow(ctx, `SELECT COUNT(*) FROM users`).Scan(&totalReviewers)
	if err != nil {
		return nil, fmt.Errorf("%s: query total reviewers: %w", op, err)
	}

	stats.TotalPRs = totalPRs
	stats.TotalReviewers = totalReviewers

	return stats, nil
}
