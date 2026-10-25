-- name: CreateComment :one
INSERT INTO comments (
    task_id, user_id, content
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetCommentByID :one
SELECT * FROM comments
WHERE id = $1
LIMIT 1;

-- name: ListCommentsByTask :many
SELECT
    c.id,
    c.task_id,
    c.user_id,
    c.content,
    c.created_at,
    u.name AS user_name,
    u.email AS user_email
FROM comments c
JOIN users u ON c.user_id = u.id
WHERE c.task_id = $1
ORDER BY c.created_at ASC;

-- name: DeleteComment :exec
DELETE FROM comments
WHERE id = $1;