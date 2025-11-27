-- name: GetLog :one
SELECT id, user_id, domain, action, content, created_at
FROM logs
WHERE id = $1;

-- name: ListLogs :many
SELECT id, user_id, domain, action, content, created_at
FROM logs
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: ListLogsByUserID :many
SELECT id, user_id, domain, action, content, created_at
FROM logs
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListLogsByDomain :many
SELECT id, user_id, domain, action, content, created_at
FROM logs
WHERE domain = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListLogsWithFilters :many
SELECT id, user_id, domain, action, content, created_at
FROM logs
WHERE 
    (sqlc.narg('user_id')::uuid IS NULL OR user_id = sqlc.narg('user_id')) AND
    (sqlc.narg('domain')::text IS NULL OR domain = sqlc.narg('domain')) AND
    (sqlc.narg('created_at_from')::timestamptz IS NULL OR created_at >= sqlc.narg('created_at_from')) AND
    (sqlc.narg('created_at_to')::timestamptz IS NULL OR created_at <= sqlc.narg('created_at_to')) AND
    (sqlc.narg('content_search')::text IS NULL OR to_tsvector('english', content::text) @@ plainto_tsquery('english', sqlc.narg('content_search')))
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: CountLogs :one
SELECT COUNT(*) FROM logs;

-- name: CountLogsWithFilters :one
SELECT COUNT(*) FROM logs
WHERE 
    (sqlc.narg('user_id')::uuid IS NULL OR user_id = sqlc.narg('user_id')) AND
    (sqlc.narg('domain')::text IS NULL OR domain = sqlc.narg('domain')) AND
    (sqlc.narg('created_at_from')::timestamptz IS NULL OR created_at >= sqlc.narg('created_at_from')) AND
    (sqlc.narg('created_at_to')::timestamptz IS NULL OR created_at <= sqlc.narg('created_at_to')) AND
    (sqlc.narg('content_search')::text IS NULL OR to_tsvector('english', content::text) @@ plainto_tsquery('english', sqlc.narg('content_search')));

-- name: CreateLog :one
INSERT INTO logs (user_id, domain, action, content, created_at)
VALUES ($1, $2, $3, $4, $5)
RETURNING id, user_id, domain, action, content, created_at;

-- name: BulkInsertLogs :copyfrom
INSERT INTO logs (user_id, domain, action, content, created_at)
VALUES ($1, $2, $3, $4, $5);

-- name: TruncateLogs :exec
TRUNCATE TABLE logs RESTART IDENTITY CASCADE;

-- name: DeleteLog :exec
DELETE FROM logs WHERE id = $1;

-- name: SearchLogsPartial :many
SELECT id, user_id, domain, action, content, created_at
FROM logs
WHERE 
    (sqlc.narg('user_id')::uuid IS NULL OR user_id = sqlc.narg('user_id')) AND
    (sqlc.narg('domain')::text IS NULL OR domain = sqlc.narg('domain')) AND
    (sqlc.narg('created_at_from')::timestamptz IS NULL OR created_at >= sqlc.narg('created_at_from')) AND
    (sqlc.narg('created_at_to')::timestamptz IS NULL OR created_at <= sqlc.narg('created_at_to')) AND
    content::text ILIKE '%' || sqlc.narg('search_term')::text || '%'
ORDER BY created_at DESC
LIMIT sqlc.narg('limit') OFFSET sqlc.narg('offset');

-- name: CountLogsPartial :one
SELECT COUNT(*) FROM logs
WHERE 
    (sqlc.narg('user_id')::uuid IS NULL OR user_id = sqlc.narg('user_id')) AND
    (sqlc.narg('domain')::text IS NULL OR domain = sqlc.narg('domain')) AND
    (sqlc.narg('created_at_from')::timestamptz IS NULL OR created_at >= sqlc.narg('created_at_from')) AND
    (sqlc.narg('created_at_to')::timestamptz IS NULL OR created_at <= sqlc.narg('created_at_to')) AND
    content::text ILIKE '%' || sqlc.narg('search_term')::text || '%';
