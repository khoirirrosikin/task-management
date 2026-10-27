-- name: CreateTask :one
INSERT INTO tasks (
    project_id, title, description, status, priority, assigned_to, due_date
) VALUES (
    $1, $2, $3, $4, $5, $6, $7
) RETURNING *;

-- name: GetTaskByID :one
SELECT * FROM tasks
WHERE id = $1 LIMIT 1;

-- name: ListTaskByProject :many
SELECT * FROM tasks
WHERE project_id = $1
    AND (sqlc.narg('status')::varchar IS NULL OR status = sqlc.narg('status'))
    AND (sqlc.narg('priority')::varchar IS NULL OR priority = sqlc.narg('priority'))
    AND (sqlc.narg('assigned_to')::uuid IS NULL OR assigned_to = sqlc.narg('assigned_to'))
    AND (sqlc.narg('due_before')::timestamptz IS NULL OR due_date <= sqlc.narg('due_before'))
    AND (sqlc.narg('search')::varchar IS NULL OR title ILIKE '%' || sqlc.narg('search') || '%' OR description ILIKE '%' || sqlc.narg('search') || '%')
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountTasksByProject :one
SELECT COUNT(*) FROM tasks
WHERE project_id = $1
    AND (sqlc.narg('status')::varchar IS NULL OR status = sqlc.narg('status'))
    AND (sqlc.narg('priority')::varchar IS NULL OR priority = sqlc.narg('priority'))
    AND (sqlc.narg('assigned_to')::uuid IS NULL OR assigned_to = sqlc.narg('assigned_to'))
    AND (sqlc.narg('due_before')::timestamptz IS NULL OR due_date <= sqlc.narg('due_before'))
    AND (sqlc.narg('search')::varchar IS NULL OR title ILIKE '%' || sqlc.narg('search') || '%' OR description ILIKE '%' || sqlc.narg('search') || '%');


-- name: UpdateTask :one
UPDATE tasks
SET
    title = COALESCE($2, title),
    description = COALESCE($3, description),
    status = COALESCE($4, status),
    priority = COALESCE($5, priority),
    assigned_to = COALESCE($6, assigned_to),
    due_date = COALESCE($7, due_date),
    updated_at = NOW()
WHERE id = $1
RETURNING *;

-- name: DeleteTask :exec
DELETE FROM tasks
WHERE id = $1;
