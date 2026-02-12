CREATE TABLE daily_snapshots (
    id                BIGSERIAL PRIMARY KEY,
    project_id        BIGINT NOT NULL REFERENCES projects(id) ON DELETE CASCADE,
    snapshot_date     DATE NOT NULL,

    stargazers_count  INT NOT NULL DEFAULT 0,
    forks_count       INT NOT NULL DEFAULT 0,
    open_issues_count INT NOT NULL DEFAULT 0,
    watchers_count    INT NOT NULL DEFAULT 0,

    score             NUMERIC(5,2),
    rank              INT,

    created_at        TIMESTAMPTZ NOT NULL DEFAULT NOW(),

    UNIQUE(project_id, snapshot_date)
);

CREATE INDEX idx_snapshots_date ON daily_snapshots(snapshot_date DESC);
CREATE INDEX idx_snapshots_project_date ON daily_snapshots(project_id, snapshot_date DESC);
CREATE INDEX idx_snapshots_rank ON daily_snapshots(snapshot_date, rank ASC) WHERE rank IS NOT NULL;
