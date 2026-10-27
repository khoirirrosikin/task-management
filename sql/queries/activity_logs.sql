-- name: CreateActivityLog :one
INSERT INTO activity_logs (
    project_id, user_id, action, entity_type, entity_id, details
) VALUES (
    $1, $2, $3, $4, $5, $6
) RETURNING *;

-- name: ListActivityLogsByProject :many
SELECT
    al.id,
    al.project_id,
    al.user_id,
    al.action,
    al.entity_type,
    al.entity_id,
    al.details,
    al.created_at,
    u.name AS user_name,
    u.email AS user_email
FROM activity_logs al
JOIN users u ON al.user_id = u.id
WHERE al.project_id = $1
ORDER BY al.created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountActivityLogsByProject :one
SELECT COUNT(*) FROM activity_logs
WHERE project_id = $1;