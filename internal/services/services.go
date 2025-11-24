package services

import (
	"database/sql"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/example/prreview/internal/repo"
	"github.com/lib/pq"
)

type Services struct {
	PR *PRService
}

var rnd = rand.New(rand.NewSource(time.Now().UnixNano()))

func NewServices(r *repo.SQLRepo) *Services {
    return &Services{PR: &PRService{repo: r}}
}

type PRService struct {
	repo *repo.SQLRepo
}

var (
	ErrPRExists      = errors.New("pr exists")
	ErrAuthorMissing = errors.New("author not found or has no team")
)

func (s *PRService) CreatePR(prID, title, authorID string) (map[string]interface{}, error) {
	tx, err := s.repo.Beginx()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var prExists bool
	if err := tx.Get(&prExists, "SELECT EXISTS(SELECT 1 FROM prs WHERE id=$1)", prID); err != nil {
		return nil, err
	}
	if prExists {
		return nil, ErrPRExists
	}

	authExists, err := s.repo.UserExistsTx(tx, authorID)
	if err != nil {
		return nil, err
	}
	if !authExists {
		return nil, ErrAuthorMissing
	}

	var team sql.NullString
	if err := tx.Get(&team, "SELECT team_name FROM team_members WHERE user_id=$1 LIMIT 1", authorID); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, ErrAuthorMissing
		}
		return nil, err
	}
	if !team.Valid {
		return nil, ErrAuthorMissing
	}

	createdAt := sql.NullTime{Time: time.Now().UTC(), Valid: true}
	if err := s.repo.InsertPRTx(tx, prID, title, authorID, team.String, createdAt); err != nil {
		return nil, err
	}

	candidates, err := s.repo.SelectActiveReviewersTx(tx, team.String, authorID)
	if err != nil {
		return nil, err
	}
	pick := pickRandomN(candidates, 2)
	for _, uid := range pick {
		if err := s.repo.AddPRReviewerTx(tx, prID, uid); err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}
	
	prModel, _ := s.repo.GetPR(prID)
	pr := map[string]interface{}{
		"id":                 prModel.PullRequestID,
		"title":              prModel.PullRequestName,
		"author":             prModel.AuthorID,
		"status":             prModel.Status,
		"assigned_reviewers": prModel.AssignedReviewers,
		"team_name":          prModel.Team_name,
	}
	return pr, err
}

func (s *PRService) MergePR(prID string) (map[string]interface{}, error) {
    tx, err := s.repo.Beginx()
    if err != nil {
        return nil, err
    }
    defer func() { _ = tx.Rollback() }()

    var status string
    if err := tx.Get(&status, "SELECT status FROM prs WHERE id=$1 FOR UPDATE", prID); err != nil {
        if errors.Is(err, sql.ErrNoRows) {
            return nil, fmt.Errorf("PR not found")
        }
        return nil, err
    }

    if status != "MERGED" {
        if _, err := tx.Exec("UPDATE prs SET status='MERGED' WHERE id=$1", prID); err != nil {
            return nil, err
        }
    }

    if err := tx.Commit(); err != nil {
        return nil, err
    }

    prModel, err := s.repo.GetPR(prID)
    if err != nil {
        return nil, err
    }

    pr := map[string]interface{}{
        "id":       prModel.PullRequestID,
        "title":    prModel.PullRequestName,
        "author":   prModel.AuthorID,
        "status":   prModel.Status,
        "reviewers": prModel.AssignedReviewers,
    }

    return pr, nil
}


func (s *PRService) Reassign(prID, oldUser string) (string, map[string]interface{}, error) {
	tx, err := s.repo.Beginx()
	if err != nil {
		return "", nil, err
	}
	defer func() { _ = tx.Rollback() }()

	var status string
	if err := tx.Get(&status, "SELECT status FROM prs WHERE id=$1 FOR UPDATE", prID); err != nil {
		return "", nil, err
	}
	if status == "MERGED" {
		return "", nil, errors.New("PR_MERGED")
	}

	var team string
	if err := tx.Get(&team, "SELECT team_name FROM prs WHERE id=$1", prID); err != nil {
		return "", nil, err
	}

	var authorID string
	if err := tx.Get(&authorID, "SELECT author_id FROM prs WHERE id=$1", prID); err != nil {
		return "", nil, err
	}

	var isAssigned bool
	if err := tx.Get(&isAssigned, "SELECT EXISTS(SELECT 1 FROM pr_reviewers WHERE pr_id=$1 AND user_id=$2)", prID, oldUser); err != nil {
		return "", nil, err
	}
	if !isAssigned {
		return "", nil, errors.New("NOT_ASSIGNED")
	}

	var current []string
	if err := tx.Select(&current, "SELECT user_id FROM pr_reviewers WHERE pr_id=$1", prID); err != nil {
		return "", nil, err
	}

	var candidates []string
	query := `SELECT u.id
			FROM users u
			JOIN team_members tm ON tm.user_id = u.id
			WHERE tm.team_name = $1
				AND u.is_active = true
				AND u.id <> ALL($2::text[])
				AND u.id <> $3`

	if err := tx.Select(&candidates, query, team, pq.Array(current), authorID); err != nil {
		return "", nil, err
	}
	if len(candidates) == 0 {
		return "", nil, errors.New("NO_CANDIDATE")
	}
	newID := candidates[rnd.Intn(len(candidates))]

	if _, err := tx.Exec("DELETE FROM pr_reviewers WHERE pr_id=$1 AND user_id=$2", prID, oldUser); err != nil {
		return "", nil, err
	}
	if _, err := tx.Exec("INSERT INTO pr_reviewers(pr_id,user_id) VALUES($1,$2)", prID, newID); err != nil {
		return "", nil, err
	}

	if err := tx.Commit(); err != nil {
		return "", nil, err
	}
	prModel, err := s.repo.GetPR(prID)
	pr := map[string]interface{}{
		"id":     prModel.PullRequestID,
		"title":  prModel.PullRequestName,
		"author": prModel.AuthorID,
	}
	return newID, pr, err
}

func pickRandomN(src []string, n int) []string {
	if len(src) <= n {
		out := make([]string, len(src))
		copy(out, src)
		return out
	}
	out := make([]string, 0, n)
	idx := rnd.Perm(len(src))[:n]
	for _, i := range idx {
		out = append(out, src[i])
	}
	return out
}
