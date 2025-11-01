-- name: CheckDBConnection :one
SELECT 1;

-- name: CreateProject :exec
INSERT INTO tacokumo_admin.projects (name, description, kind) VALUES ($1, $2, $3);

-- name: GetOwnedPersonalProject :one 
SELECT id, name, description, kind, created_at, updated_at
FROM tacokumo_admin.projects
WHERE kind = 'personal' AND id IN (
    SELECT project_id
    FROM tacokumo_admin.user_role_relations
    WHERE user_id = $1
);


-- name: ListProjectsWithPagination :many
SELECT id, name, description, kind, created_at, updated_at
FROM tacokumo_admin.projects
WHERE name IN (sqlc.slice('names'))
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetProjectByName :one
SELECT id, name, description, kind, created_at, updated_at
FROM tacokumo_admin.projects
WHERE name = $1;