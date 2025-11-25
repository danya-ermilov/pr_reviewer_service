package repo

import (
	"database/sql"
	"log"
	"os"
	"testing"

	"github.com/jmoiron/sqlx"
	_ "github.com/lib/pq"
	"github.com/ory/dockertest/v3"
	"github.com/ory/dockertest/v3/docker"
	"github.com/stretchr/testify/require"

	"github.com/example/prreview/internal/models"
)

var (
	testDB   *sqlx.DB
	testRepo *SQLRepo
)

func TestMain(m *testing.M) {
	pool, err := dockertest.NewPool("")
	if err != nil {
		log.Fatalf("could not connect to docker: %v", err)
	}

	runOptions := &dockertest.RunOptions{
		Repository: "postgres",
		Tag:        "16",
		Env: []string{
			"POSTGRES_USER=pruser",
			"POSTGRES_PASSWORD=prpass",
			"POSTGRES_DB=pr_review",
		},
	}

	resource, err := pool.RunWithOptions(runOptions, func(h *docker.HostConfig) {
		h.AutoRemove = true
		h.RestartPolicy = docker.RestartPolicy{Name: "no"}
	})
	if err != nil {
		log.Fatalf("could not start resource: %v", err)
	}
	if err := resource.Expire(600); err != nil {
		log.Printf("could not set resource expiration: %v", err)
	}

	var db *sqlx.DB
	if err := pool.Retry(func() error {
		dsn := "postgres://pruser:prpass@localhost:5433/pr_review?sslmode=disable"
		dbx, err := sqlx.Connect("postgres", dsn)
		if err != nil {
			return err
		}
		if err := dbx.Ping(); err != nil {
			_ = dbx.Close()
			return err
		}
		db = dbx
		return nil
	}); err != nil {
		_ = pool.Purge(resource)
		log.Fatalf("could not connect to database: %v", err)
	}

	var exists bool
	err = db.Get(&exists, "SELECT to_regclass('public.prs') IS NOT NULL")
	if err != nil {
		return
	}
	if exists {
		log.Println("Migration skipped, table prs already exists")
		return
	}

	sqlBytes, err := os.ReadFile("/mnt/c/Users/dell/Desktop/pr_reviewer_service/migrations/0001_init.sql")
	if err != nil {
		return
	}
	if _, err := db.Exec(string(sqlBytes)); err != nil {
		return
	}

	testDB = db
	testRepo = NewSQLRepo(testDB)

	code := m.Run()

	_ = testDB.Close()
	if err := pool.Purge(resource); err != nil {
		log.Printf("could not purge resource: %v", err)
	}

	os.Exit(code)
}

func (r *SQLRepo) WipeTables(t *testing.T) {
	t.Helper()
	_, err := testDB.Exec(`
		TRUNCATE TABLE pr_reviewers, prs, team_members, teams, users RESTART IDENTITY CASCADE;
	`)
	require.NoError(t, err)
}

func TestCreateAndGetTeam(t *testing.T) {
	testRepo.WipeTables(t)

	members := []models.TeamMemberResp{
		{UserID: "u1", Username: "Alice", IsActive: true},
		{UserID: "u2", Username: "Bob", IsActive: true},
	}

	team, err := testRepo.CreateTeam("team-alpha", members)
	require.NoError(t, err)
	require.Equal(t, "team-alpha", team.TeamName)
	require.Len(t, team.Members, 2)

	got, err := testRepo.GetTeamByName("team-alpha")
	require.NoError(t, err)
	require.Equal(t, "team-alpha", got.TeamName)
	require.Len(t, got.Members, 2)
}

func TestSetUserActive(t *testing.T) {
	testRepo.WipeTables(t)

	members := []models.TeamMemberResp{{UserID: "u1", Username: "Alice", IsActive: true}}
	_, err := testRepo.CreateTeam("team-beta", members)
	require.NoError(t, err)

	user, err := testRepo.SetUserActive("u1", false)
	require.NoError(t, err)
	require.Equal(t, false, user.IsActive)
	require.Equal(t, "Alice", user.Username)
	require.Equal(t, "team-beta", user.TeamName)

	user, err = testRepo.SetUserActive("u1", true)
	require.NoError(t, err)
	require.True(t, user.IsActive)
}

func TestInsertAndGetPR(t *testing.T) {
	testRepo.WipeTables(t)

	members := []models.TeamMemberResp{
		{UserID: "u1", Username: "Alice", IsActive: true},
		{UserID: "u2", Username: "Bob", IsActive: true},
	}
	_, err := testRepo.CreateTeam("team-gamma", members)
	require.NoError(t, err)

	tx, err := testRepo.Beginx()
	require.NoError(t, err)
	err = testRepo.InsertPRTx(tx, "pr1", "Add Feature", "u1", "team-gamma", sql.NullTime{})
	require.NoError(t, err)

	err = testRepo.AddPRReviewerTx(tx, "pr1", "u2")
	require.NoError(t, err)

	require.NoError(t, tx.Commit())

	pr, err := testRepo.GetPR("pr1")
	require.NoError(t, err)
	require.Equal(t, "pr1", pr.PullRequestID)
	require.Equal(t, "Add Feature", pr.PullRequestName)
	require.Equal(t, "u1", pr.AuthorID)
	require.Equal(t, []string{"u2"}, pr.AssignedReviewers)
}

func TestGetPRsAndReviewsForUser(t *testing.T) {
	testRepo.WipeTables(t)

	members := []models.TeamMemberResp{
		{UserID: "u1", Username: "Alice", IsActive: true},
		{UserID: "u2", Username: "Bob", IsActive: true},
	}
	_, err := testRepo.CreateTeam("team-delta", members)
	require.NoError(t, err)

	tx, err := testRepo.Beginx()
	require.NoError(t, err)
	require.NoError(t, testRepo.InsertPRTx(tx, "pr1", "PR 1", "u1", "team-delta", sql.NullTime{}))
	require.NoError(t, testRepo.InsertPRTx(tx, "pr2", "PR 2", "u1", "team-delta", sql.NullTime{}))

	require.NoError(t, testRepo.AddPRReviewerTx(tx, "pr1", "u2"))
	require.NoError(t, testRepo.AddPRReviewerTx(tx, "pr2", "u2"))
	require.NoError(t, tx.Commit())

	prs, err := testRepo.GetPRsForUser("u2")
	require.NoError(t, err)
	require.Len(t, prs, 2)

	reviews, err := testRepo.GetReviewsForUser("u2")
	require.NoError(t, err)
	require.Len(t, reviews, 2)
}
