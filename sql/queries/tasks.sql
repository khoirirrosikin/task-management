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
ORDER BY created_at DESC;

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
