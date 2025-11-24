package models

type TeamMemberResp struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	IsActive bool   `json:"is_active"`
}

type TeamResp struct {
	TeamName string           `json:"team_name"`
	Members  []TeamMemberResp `json:"members"`
}

type UserResp struct {
	UserID   string `json:"user_id"`
	Username string `json:"username"`
	TeamName string `json:"team_name"`
	IsActive bool   `json:"is_active"`
}

type PullRequestResp struct {
	PullRequestID     string   `json:"pull_request_id" db:"id"`
	PullRequestName   string   `json:"pull_request_name" db:"title"`
	AuthorID          string   `json:"author_id" db:"author_id"`
	Status            string   `json:"status" db:"status"`
	AssignedReviewers []string `json:"assigned_reviewers" db:"-"`
	Team_name         string   `json:"team_name" db:"team_name"`
}

type PullRequestShortResp struct {
    PullRequestID   string `db:"pull_request_id" json:"pull_request_id"`
    PullRequestName string `db:"pull_request_name" json:"pull_request_name"`
    AuthorID        string `db:"author_id" json:"author_id"`
    Status          string `db:"status" json:"status"`
}
