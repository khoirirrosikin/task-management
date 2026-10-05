-- name: AddProjectMember :one
INSERT INTO project_members (
    project_id, user_id, role
) VALUES (
    $1, $2, $3
) RETURNING *;

-- name: GetProjectMember :one
SELECT * FROM project_members
WHERE project_id = $1 AND user_id = $2
LIMIT 1;

-- name: ListProjectMembers :many
SELECT
    pm.id,
    pm.project_id,
    pm.user_id,
    pm.role,
    pm.joined_at,
    u.name AS user_name,
    u.email AS user_email
FROM project_members pm
JOIN users u ON pm.user_id = u.id
WHERE pm.project_id = $1
ORDER BY pm.joined_at ASC;

-- name: UpdateProjectMemberRole :one
UPDATE project_members
SET role = $3
WHERE project_id = $1 AND user_id = $2
RETURNING *;

-- name: RemoveProjectMember :exec
DELETE FROM project_members
WHERE project_id = $1 AND user_id = $2;