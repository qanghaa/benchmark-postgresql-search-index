-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS "uuid-ossp";

CREATE TABLE logs (
    id UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    user_id UUID NOT NULL,
    domain VARCHAR(255) NOT NULL,
    action VARCHAR(255) NOT NULL,
    content JSONB NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- B-tree indexes for user_id and domain
CREATE INDEX idx_logs_user_id ON logs USING BTREE (user_id);
CREATE INDEX idx_logs_domain ON logs USING BTREE (domain);

-- GIN index for full-text search on JSONB content
CREATE INDEX idx_logs_content_gin ON logs USING GIN (content);

-- BRIN index for created_at (optimized for time-series data)
CREATE INDEX idx_logs_created_at_brin ON logs USING BRIN (created_at);

-- Composite index for common queries
CREATE INDEX idx_logs_user_domain_created ON logs USING BTREE (user_id, domain, created_at);

-- Full-text search index on content text representation
CREATE INDEX idx_logs_content_fts ON logs USING GIN (to_tsvector('english', content::text));
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE IF EXISTS logs CASCADE;
-- +goose StatementEnd
