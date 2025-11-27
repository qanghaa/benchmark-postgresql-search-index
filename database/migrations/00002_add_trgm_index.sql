-- +goose Up
-- +goose StatementBegin
CREATE EXTENSION IF NOT EXISTS pg_trgm;

-- Create GIN index with trgm ops for efficient ILIKE search on JSONB content cast to text
CREATE INDEX idx_logs_content_trgm ON logs USING GIN ((content::text) gin_trgm_ops);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP INDEX IF EXISTS idx_logs_content_trgm;
DROP EXTENSION IF EXISTS pg_trgm;
-- +goose StatementEnd
