CREATE TABLE projects (
    id              BIGSERIAL PRIMARY KEY,
    github_id       BIGINT NOT NULL UNIQUE,
    full_name       VARCHAR(255) NOT NULL UNIQUE,
    description     TEXT,
    language        VARCHAR(50),
    license         VARCHAR(50),
    topics          TEXT[] DEFAULT '{}',
    homepage        VARCHAR(500),
    created_at_gh   TIMESTAMPTZ,
    pushed_at       TIMESTAMPTZ,
    metadata        JSONB DEFAULT '{}',

    -- Current metrics (updated daily)
    stargazers_count  INT NOT NULL DEFAULT 0,
    forks_count       INT NOT NULL DEFAULT 0,
    open_issues_count INT NOT NULL DEFAULT 0,
    watchers_count    INT NOT NULL DEFAULT 0,

    -- Analysis results
    score           NUMERIC(5,2) DEFAULT 0,
    rank            INT,

    -- System fields
    first_seen_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    is_archived     BOOLEAN DEFAULT FALSE,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_projects_github_id ON projects(github_id);
CREATE INDEX idx_projects_full_name ON projects(full_name);
CREATE INDEX idx_projects_score ON projects(score DESC);
CREATE INDEX idx_projects_rank ON projects(rank ASC) WHERE rank IS NOT NULL;
CREATE INDEX idx_projects_language ON projects(language);
CREATE INDEX idx_projects_topics ON projects USING GIN(topics);
CREATE INDEX idx_projects_metadata ON projects USING GIN(metadata);
CREATE INDEX idx_projects_first_seen ON projects(first_seen_at DESC);

-- Auto-update updated_at timestamp
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = NOW();
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER update_projects_updated_at
    BEFORE UPDATE ON projects
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
