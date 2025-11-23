package int_errors

import "errors"

var (
	ErrPRExists                = errors.New("pull request already exists")
	ErrUserNotFound            = errors.New("user not found")
	ErrPRNotFound              = errors.New("pull request not found")
	ErrTeamNotFound            = errors.New("team not found")
	ErrTeamExists              = errors.New("team already exists")
	ErrPRMerged                = errors.New("pull request already merged")
	ErrReviewerNotAssigned     = errors.New("reviewer is not assigned to pull request")
	ErrNoReplacementCandidate  = errors.New("no active candidate available")
	ErrUserHasOpenPullRequests = errors.New("user has open pull requests")
	ErrAuthorNotFound          = errors.New("author not found")
	ErrAuthorNotActive         = errors.New("author is not active")
	ErrNoTeamFound             = errors.New("author has no team")
)
