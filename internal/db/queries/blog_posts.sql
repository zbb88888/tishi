-- name: UpsertBlogPost :one
INSERT INTO blog_posts (title, slug, content, post_type, published_at)
VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (slug) DO UPDATE SET
    title = EXCLUDED.title,
    content = EXCLUDED.content,
    published_at = EXCLUDED.published_at
RETURNING id;

-- name: GetBlogPostBySlug :one
SELECT * FROM blog_posts WHERE slug = $1;

-- name: ListBlogPosts :many
SELECT id, title, slug, post_type, published_at, created_at,
    LEFT(content, 200) AS excerpt
FROM blog_posts
WHERE published_at IS NOT NULL
ORDER BY published_at DESC
LIMIT $1 OFFSET $2;

-- name: ListBlogPostsByType :many
SELECT id, title, slug, post_type, published_at, created_at,
    LEFT(content, 200) AS excerpt
FROM blog_posts
WHERE post_type = $1 AND published_at IS NOT NULL
ORDER BY published_at DESC
LIMIT $2 OFFSET $3;

-- name: CountBlogPosts :one
SELECT COUNT(*) FROM blog_posts WHERE published_at IS NOT NULL;
