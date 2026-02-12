-- name: InsertSnapshot :exec
INSERT INTO daily_snapshots (
    project_id, snapshot_date,
    stargazers_count, forks_count, open_issues_count, watchers_count
) VALUES ($1, $2, $3, $4, $5, $6)
ON CONFLICT (project_id, snapshot_date) DO UPDATE SET
    stargazers_count = EXCLUDED.stargazers_count,
    forks_count = EXCLUDED.forks_count,
    open_issues_count = EXCLUDED.open_issues_count,
    watchers_count = EXCLUDED.watchers_count;

-- name: UpdateSnapshotScore :exec
UPDATE daily_snapshots
SET score = $3, rank = $4
WHERE project_id = $1 AND snapshot_date = $2;

-- name: GetProjectTrends :many
SELECT snapshot_date, stargazers_count, forks_count, open_issues_count, watchers_count, score, rank
FROM daily_snapshots
WHERE project_id = $1
  AND snapshot_date >= CURRENT_DATE - ($2 || ' days')::INTERVAL
ORDER BY snapshot_date ASC;

-- name: GetDailyStarGain :many
SELECT
    ds.project_id,
    ds.snapshot_date,
    ds.stargazers_count,
    ds.stargazers_count - LAG(ds.stargazers_count) OVER (
        PARTITION BY ds.project_id ORDER BY ds.snapshot_date
    ) AS daily_gain
FROM daily_snapshots ds
WHERE ds.project_id = $1
ORDER BY ds.snapshot_date DESC
LIMIT $2;

-- name: GetWeeklyStarGain :one
SELECT
    COALESCE(today.stargazers_count - week_ago.stargazers_count, 0) AS weekly_gain
FROM daily_snapshots today
LEFT JOIN daily_snapshots week_ago
    ON today.project_id = week_ago.project_id
    AND week_ago.snapshot_date = CURRENT_DATE - 7
WHERE today.project_id = $1
  AND today.snapshot_date = CURRENT_DATE;

-- name: GetLatestSnapshot :one
SELECT * FROM daily_snapshots
WHERE project_id = $1
ORDER BY snapshot_date DESC
LIMIT 1;

-- name: GetYesterdayRankings :many
SELECT project_id, rank
FROM daily_snapshots
WHERE snapshot_date = CURRENT_DATE - 1
  AND rank IS NOT NULL
ORDER BY rank ASC;
