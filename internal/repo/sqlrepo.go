package repo

import (
	"database/sql"
	"errors"
	"fmt"
	"os"

	"github.com/jmoiron/sqlx"

	"github.com/example/prreview/internal/models"
)

func RunMigrations(db *sqlx.DB) error {
	var exists bool
	err := db.Get(&exists, "SELECT to_regclass('public.prs') IS NOT NULL")
	if err != nil {
		return err
	}
	if exists {
		return nil
	}
	sqlBytes, err := os.ReadFile("migrations/0001_init.sql")
	if err != nil {
		return err
	}
	_, err = db.Exec(string(sqlBytes))
	return err
}

// --- Team ---

func (r *SQLRepo) GetTeamByName(teamName string) (*models.TeamResp, error) {
	var exists bool
	if err := r.DB.Get(&exists, "SELECT EXISTS(SELECT 1 FROM teams WHERE name=$1)", teamName); err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("team not found")
	}

	rows, err := r.DB.Queryx(`
		SELECT u.id, u.name, u.is_active
		FROM users u
		JOIN team_members tm ON tm.user_id = u.id
		WHERE tm.team_name = $1
	`, teamName)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var members []models.TeamMemberResp
	for rows.Next() {
		var m models.TeamMemberResp
		if err := rows.Scan(&m.UserID, &m.Username, &m.IsActive); err != nil {
			return nil, err
		}
		members = append(members, m)
	}

	return &models.TeamResp{
		TeamName: teamName,
		Members:  members,
	}, nil
}

func (r *SQLRepo) CreateTeam(teamName string, members []models.TeamMemberResp) (*models.TeamResp, error) {
	tx, err := r.DB.Beginx()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()


	var exists bool
	if err := tx.Get(&exists, "SELECT EXISTS(SELECT 1 FROM teams WHERE name=$1)", teamName); err != nil {
		return nil, err
	}
	if exists {
		return nil, errors.New("team already exists")
	}

	if _, err := tx.Exec("INSERT INTO teams(name) VALUES($1)", teamName); err != nil {
		return nil, err
	}

	for _, m := range members {
		if m.UserID == "" {
			continue
		}
		_, err := tx.Exec(`
			INSERT INTO users(id, name, is_active)
			VALUES($1,$2,$3)
			ON CONFLICT (id) DO UPDATE SET name=EXCLUDED.name, is_active=EXCLUDED.is_active
		`, m.UserID, m.Username, m.IsActive)
		if err != nil {
			return nil, err
		}

		_, err = tx.Exec(`
			INSERT INTO team_members(team_name, user_id) VALUES($1,$2)
			ON CONFLICT DO NOTHING
		`, teamName, m.UserID)
		if err != nil {
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.GetTeamByName(teamName)
}

// --- Users ---

func (r *SQLRepo) SetUserActive(userID string, isActive bool) (*models.UserResp, error) {
	tx, err := r.DB.Beginx()
	if err != nil {
		return nil, err
	}
	defer func() { _ = tx.Rollback() }()


	var exists bool
	if err := tx.Get(&exists, "SELECT EXISTS(SELECT 1 FROM users WHERE id=$1)", userID); err != nil {
		return nil, err
	}
	if !exists {
		return nil, errors.New("user not found")
	}

	if _, err := tx.Exec("UPDATE users SET is_active=$1 WHERE id=$2", isActive, userID); err != nil {
		return nil, err
	}

	var username string
	if err := tx.Get(&username, "SELECT name FROM users WHERE id=$1", userID); err != nil {
		return nil, err
	}

	var teamName sql.NullString
	if err := tx.Get(&teamName, "SELECT team_name FROM team_members WHERE user_id=$1 LIMIT 1", userID); err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	u := &models.UserResp{
		UserID:   userID,
		Username: username,
		IsActive: isActive,
	}
	if teamName.Valid {
		u.TeamName = teamName.String
	} else {
		u.TeamName = ""
	}

	return u, nil
}

// --- Pull Requests ---

func (r *SQLRepo) GetReviewsForUser(userID string) ([]models.PullRequestShortResp, error) {
	var prs []models.PullRequestShortResp
	err := r.DB.Select(&prs, `
		SELECT p.id AS pull_request_id, p.title AS pull_request_name, p.author_id, p.status
		FROM prs p
		JOIN pr_reviewers r ON r.pr_id = p.id
		WHERE r.user_id = $1
	`, userID)
	return prs, err
}

type SQLRepo struct {
	DB *sqlx.DB
}

func NewSQLRepo(db *sqlx.DB) *SQLRepo {
	return &SQLRepo{DB: db}
}

func (r *SQLRepo) Beginx() (*sqlx.Tx, error) {
	return r.DB.Beginx()
}

func (r *SQLRepo) UserExistsTx(tx *sqlx.Tx, userID string) (bool, error) {
	var exists bool
	err := tx.Get(&exists, "SELECT EXISTS(SELECT 1 FROM users WHERE id=$1)", userID)
	return exists, err
}

func (r *SQLRepo) InsertPRTx(tx *sqlx.Tx, prID, title, authorID, team string, createdAt sql.NullTime) error {
	_, err := tx.Exec(
		"INSERT INTO prs(id,title,author_id,team_name,status) VALUES($1,$2,$3,$4,'OPEN')",
		prID, title, authorID, team,
	)
	return err
}


func (r *SQLRepo) GetPR(prID string) (*models.PullRequestResp, error) {
    var pr models.PullRequestResp
    err := r.DB.Get(&pr, "SELECT id, title, author_id, team_name, status FROM prs WHERE id=$1", prID)
    if err != nil {
        return nil, err
    }

    var reviewers []string
    err = r.DB.Select(&reviewers, "SELECT user_id FROM pr_reviewers WHERE pr_id=$1", prID)
    if err != nil {
        return nil, err
    }
    pr.AssignedReviewers = reviewers
    return &pr, nil
}

func (r *SQLRepo) SelectActiveReviewersTx(tx *sqlx.Tx, teamName, authorID string) ([]string, error) {
	var users []string
	err := tx.Select(&users, "SELECT user_id FROM team_members WHERE team_name=$1 AND user_id<>$2", teamName, authorID)
	return users, err
}

func (r *SQLRepo) AddPRReviewerTx(tx *sqlx.Tx, prID, userID string) error {
	_, err := tx.Exec("INSERT INTO pr_reviewers(pr_id,user_id) VALUES($1,$2)", prID, userID)
	return err
}

func (r *SQLRepo) SetUserIsActive(userID string, isActive bool) error {
	_, err := r.DB.Exec(`UPDATE users SET is_active=$1 WHERE id=$2`, isActive, userID)
	return err
}

func (r *SQLRepo) GetPRsForUser(userID string) ([]*models.PullRequestShortResp, error) {
    var prsVals []models.PullRequestShortResp

    query := `
        SELECT pr.id AS pull_request_id,
               pr.title AS pull_request_name,
               pr.author_id AS author_id,
               pr.status AS status
        FROM pr_reviewers rr
        JOIN prs pr ON rr.pr_id = pr.id
        WHERE rr.user_id = $1
    `

    if err := r.DB.Select(&prsVals, query, userID); err != nil {
        return nil, fmt.Errorf("GetPRsForUser: select query failed (user=%s): %w", userID, err)
    }

    prs := make([]*models.PullRequestShortResp, 0, len(prsVals))
    for i := range prsVals {
        p := prsVals[i]
        prs = append(prs, &p)
    }

    return prs, nil
}

