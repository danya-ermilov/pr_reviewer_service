CREATE TABLE migrations (
    id TEXT PRIMARY KEY,
    applied_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT now()
);

CREATE TABLE users (
    id TEXT PRIMARY KEY,
    name TEXT NOT NULL,
    is_active BOOLEAN NOT NULL DEFAULT true
);

CREATE TABLE teams (
    name TEXT PRIMARY KEY
);

CREATE TABLE team_members (
    team_name TEXT NOT NULL REFERENCES teams(name) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (team_name, user_id)
);

DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM pg_type WHERE typname = 'pr_status') THEN
        CREATE TYPE pr_status AS ENUM ('OPEN', 'MERGED');
    END IF;
END$$;

CREATE TABLE prs (
    id TEXT PRIMARY KEY,
    title TEXT NOT NULL,
    author_id TEXT NOT NULL REFERENCES users(id) ON DELETE RESTRICT,
    team_name TEXT NOT NULL REFERENCES teams(name) ON DELETE RESTRICT,
    status pr_status NOT NULL DEFAULT 'OPEN'
);

CREATE TABLE pr_reviewers (
    pr_id TEXT NOT NULL REFERENCES prs(id) ON DELETE CASCADE,
    user_id TEXT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (pr_id, user_id)
);

CREATE INDEX idx_prs_team ON prs(team_name);
CREATE INDEX idx_pr_reviewers_user ON pr_reviewers(user_id);
