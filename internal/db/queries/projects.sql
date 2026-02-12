-- name: UpsertProject :one
INSERT INTO projects (
    github_id, full_name, description, language, license,
    topics, homepage, created_at_gh, pushed_at, metadata,
    stargazers_count, forks_count, open_issues_count, watchers_count
) VALUES (
    $1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14
)
ON CONFLICT (github_id) DO UPDATE SET
    full_name = EXCLUDED.full_name,
    description = EXCLUDED.description,
    language = EXCLUDED.language,
    license = EXCLUDED.license,
    topics = EXCLUDED.topics,
    homepage = EXCLUDED.homepage,
    pushed_at = EXCLUDED.pushed_at,
    metadata = EXCLUDED.metadata,
    stargazers_count = EXCLUDED.stargazers_count,
    forks_count = EXCLUDED.forks_count,
    open_issues_count = EXCLUDED.open_issues_count,
    watchers_count = EXCLUDED.watchers_count
RETURNING id;

-- name: GetProjectByID :one
SELECT * FROM projects WHERE id = $1;

-- name: GetProjectByGitHubID :one
SELECT * FROM projects WHERE github_id = $1;

-- name: GetProjectByFullName :one
SELECT * FROM projects WHERE full_name = $1;

-- name: ListProjectsByRank :many
SELECT p.*,
    array_agg(DISTINCT c.slug) FILTER (WHERE c.slug IS NOT NULL) AS category_slugs
FROM projects p
LEFT JOIN project_categories pc ON p.id = pc.project_id
LEFT JOIN categories c ON pc.category_id = c.id
WHERE p.rank IS NOT NULL AND p.rank <= $1
GROUP BY p.id
ORDER BY p.rank ASC
LIMIT $2 OFFSET $3;

-- name: ListProjects :many
SELECT p.*,
    array_agg(DISTINCT c.slug) FILTER (WHERE c.slug IS NOT NULL) AS category_slugs
FROM projects p
LEFT JOIN project_categories pc ON p.id = pc.project_id
LEFT JOIN categories c ON pc.category_id = c.id
WHERE p.is_archived = FALSE
GROUP BY p.id
ORDER BY p.score DESC
LIMIT $1 OFFSET $2;

-- name: ListProjectsByCategory :many
SELECT p.*,
    array_agg(DISTINCT c2.slug) FILTER (WHERE c2.slug IS NOT NULL) AS category_slugs
FROM projects p
JOIN project_categories pc ON p.id = pc.project_id
JOIN categories c ON pc.category_id = c.id AND c.slug = $1
LEFT JOIN project_categories pc2 ON p.id = pc2.project_id
LEFT JOIN categories c2 ON pc2.category_id = c2.id
WHERE p.is_archived = FALSE
GROUP BY p.id
ORDER BY p.score DESC
LIMIT $2 OFFSET $3;

-- name: CountProjects :one
SELECT COUNT(*) FROM projects WHERE is_archived = FALSE;

-- name: UpdateProjectScore :exec
UPDATE projects SET score = $2, rank = $3 WHERE id = $1;

-- name: MarkProjectArchived :exec
UPDATE projects SET is_archived = TRUE WHERE github_id = $1;
