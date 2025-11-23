package model

type ReviewerStats struct {
	UserID        string `json:"user_id"`
	Username      string `json:"username"`
	TeamName      string `json:"team_name"`
	TotalReviews  int    `json:"total_reviews"`
	OpenReviews   int    `json:"open_reviews"`
	MergedReviews int    `json:"merged_reviews"`
}

type PRStats struct {
	PullRequestID   string `json:"pull_request_id"`
	PullRequestName string `json:"pull_request_name"`
	AuthorID        string `json:"author_id"`
	Status          string `json:"status"`
	ReviewersCount  int    `json:"reviewers_count"`
}

type Statistics struct {
	TotalPRs       int             `json:"total_prs"`
	TotalReviewers int             `json:"total_reviewers"`
	ReviewersStats []ReviewerStats `json:"reviewers_stats"`
	PRStats        []PRStats       `json:"pr_stats"`
}
