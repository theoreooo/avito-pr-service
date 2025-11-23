CREATE INDEX IF NOT EXISTS idx_users_team_active 
    ON users(team_name, is_active) WHERE is_active = TRUE;

CREATE INDEX IF NOT EXISTS idx_users_team ON users(team_name);
CREATE INDEX IF NOT EXISTS idx_pr_reviewers_reviewer ON pr_reviewers(reviewer_id);
CREATE INDEX IF NOT EXISTS idx_pr_status ON pull_requests(status);
CREATE INDEX IF NOT EXISTS idx_pr_author ON pull_requests(author_id);