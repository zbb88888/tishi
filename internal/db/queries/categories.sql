-- name: ListCategories :many
SELECT c.*,
    COUNT(pc.project_id) AS project_count
FROM categories c
LEFT JOIN project_categories pc ON c.id = pc.category_id
GROUP BY c.id
ORDER BY c.sort_order ASC;

-- name: GetCategoryBySlug :one
SELECT * FROM categories WHERE slug = $1;

-- name: UpsertProjectCategory :exec
INSERT INTO project_categories (project_id, category_id, confidence)
VALUES ($1, $2, $3)
ON CONFLICT (project_id, category_id) DO UPDATE SET
    confidence = EXCLUDED.confidence;

-- name: DeleteProjectCategories :exec
DELETE FROM project_categories WHERE project_id = $1;
