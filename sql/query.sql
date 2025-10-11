-- name: CheckDBConnection :one
SELECT 1;

-- name: CheckAccountExists :one
SELECT EXISTS (
  SELECT 1
  FROM tacokumo_admin.accounts
  WHERE email = $1
);

-- name: CreateProject :exec
INSERT INTO tacokumo_admin.projects (name, bio) VALUES ($1, $2);

-- name: ListProjectsWithPagination :many
SELECT id, name, bio, created_at, updated_at
FROM tacokumo_admin.projects
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetProjectByName :one
SELECT id, name, bio, created_at, updated_at
FROM tacokumo_admin.projects
WHERE name = $1;