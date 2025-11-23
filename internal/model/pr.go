package model

import "time"

const (
	StatusOpen   string = "OPEN"
	StatusMerged string = "MERGED"
)

type PullRequest struct {
	PRID              string
	PRName            string
	AuthorID          string
	Status            string
	AssignedReviewers []string
	CreatedAt         time.Time
	MergedAt          *time.Time
}

type CreatePRRequest struct {
	PRID     string
	PRName   string
	AuthorID string
}
