-- name: CheckDBConnection :one
SELECT 1;

-- name: CreateProject :exec
INSERT INTO tacokumo_admin.projects (name, description, kind) VALUES ($1, $2, $3);

-- name: ListProjectsWithPagination :many
SELECT id, name, description, kind, created_at, updated_at
FROM tacokumo_admin.projects
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetProjectByName :one
SELECT id, name, description, kind, created_at, updated_at
FROM tacokumo_admin.projects
WHERE name = $1;