-- Indexes for cursor pagination performance
CREATE INDEX IF NOT EXISTS idx_posts_cursor_created_at ON posts(created_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_posts_cursor_updated_at ON posts(updated_at DESC, id DESC);
CREATE INDEX IF NOT EXISTS idx_posts_cursor_title ON posts(title, id);

-- Composite index for filtering
CREATE INDEX IF NOT EXISTS idx_posts_author_created_at ON posts(author, created_at DESC);

-- GIN index for better search performance
CREATE EXTENSION IF NOT EXISTS pg_trgm;
CREATE INDEX IF NOT EXISTS idx_posts_search_gin ON posts 
USING gin((title || ' ' || content) gin_trgm_ops);