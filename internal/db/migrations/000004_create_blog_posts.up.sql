CREATE TABLE blog_posts (
    id              BIGSERIAL PRIMARY KEY,
    title           VARCHAR(500) NOT NULL,
    slug            VARCHAR(500) NOT NULL UNIQUE,
    content         TEXT NOT NULL,
    post_type       VARCHAR(20) NOT NULL CHECK (post_type IN ('weekly', 'monthly', 'spotlight')),
    cover_image_url VARCHAR(500),
    published_at    TIMESTAMPTZ,
    created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_posts_type ON blog_posts(post_type);
CREATE INDEX idx_posts_published ON blog_posts(published_at DESC);
CREATE INDEX idx_posts_slug ON blog_posts(slug);

CREATE TRIGGER update_blog_posts_updated_at
    BEFORE UPDATE ON blog_posts
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();
