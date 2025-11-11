-- name: CheckDBConnection :one
SELECT 1;

-- name: CreateProject :exec
INSERT INTO tacokumo_admin.projects (name, description, kind) VALUES ($1, $2, $3);

-- name: ListProjectsWithPagination :many
SELECT id, display_id, name, description, kind, created_at, updated_at
FROM tacokumo_admin.projects
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: GetProjectByDisplayID :one
SELECT id, display_id, name, description, kind, created_at, updated_at
FROM tacokumo_admin.projects
WHERE display_id = $1;

-- name: GetProjectByName :one
SELECT id, display_id, name, description, kind, created_at, updated_at
FROM tacokumo_admin.projects
WHERE name = $1;

-- name: UpdateProject :exec
UPDATE tacokumo_admin.projects
SET (name, description, updated_at) = ($2, $3, NOW())
WHERE display_id = $1;

-- name: CreateRole :exec
INSERT INTO tacokumo_admin.roles (project_id, name, description) VALUES ($1, $2, $3);

-- name: GetRoleByDisplayID :one
SELECT id, display_id, project_id, name, description, created_at, updated_at
FROM tacokumo_admin.roles
WHERE project_id = $1 AND display_id = $2;

-- name: ListRolesWithPagination :many
SELECT id, display_id, project_id, name, description, created_at, updated_at
FROM tacokumo_admin.roles
WHERE project_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: UpdateRole :exec
UPDATE tacokumo_admin.roles
SET (name, description, updated_at) = ($3, $4, NOW())
WHERE project_id = $1 AND display_id = $2;

-- name: CreateUserGroup :exec
INSERT INTO tacokumo_admin.usergroups (project_id, name, description) VALUES ($1, $2, $3);

-- name: GetUserGroupByDisplayID :one
SELECT id, display_id, project_id, name, description, created_at, updated_at
FROM tacokumo_admin.usergroups
WHERE project_id = $1 AND display_id = $2;

-- name: ListUserGroupsWithPagination :many
SELECT id, display_id, project_id, name, description, created_at, updated_at
FROM tacokumo_admin.usergroups
WHERE project_id = $1
ORDER BY created_at DESC
LIMIT $2 OFFSET $3;

-- name: ListUserGroupMembers :many
SELECT u.id,
      u.display_id,
      u.email,
      u.created_at,
      u.updated_at
  FROM tacokumo_admin.users u
  INNER JOIN tacokumo_admin.user_usergroups_relations uur ON u.id = uur.user_id
  WHERE uur.usergroup_id = $1
  ORDER BY u.created_at DESC;

-- name: UpdateUserGroup :exec
UPDATE tacokumo_admin.usergroups
SET (name, description, updated_at) = ($3, $4, NOW())
WHERE project_id = $1 AND display_id = $2;

-- name: CreateUser :exec
INSERT INTO tacokumo_admin.users (email) VALUES ($1);

-- name: GetUserByEmail :one
SELECT id, display_id, email, created_at, updated_at
FROM tacokumo_admin.users
WHERE email = $1;

-- name: ListUsersWithPagination :many
SELECT id, display_id, email, created_at, updated_at
FROM tacokumo_admin.users
ORDER BY created_at DESC
LIMIT $1 OFFSET $2;

-- name: UpdateUser :exec
UPDATE tacokumo_admin.users
SET (email, updated_at) = ($2, NOW())
WHERE display_id = $1; 