-- name: CreateNotification :one
INSERT INTO notifications (
    user_id, title, message
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: ListNotificationsByUser :many
SELECT * FROM notifications
WHERE user_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: CountNotificationsByUser :one
SELECT COUNT(*) FROM notifications
WHERE user_id = $1;

-- name: MarkNotificationAsRead :one
UPDATE notifications
SET is_read = TRUE
WHERE id = $1 AND user_id = $2
RETURNING *;

-- name: MarkAllNotificationsAsRead :exec
UPDATE notifications
SET is_read = TRUE
WHERE user_id = $1 AND is_read = FALSE;
