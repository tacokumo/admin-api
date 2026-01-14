package admindb

import "errors"

// Common database errors for testing purposes
var (
	ErrProjectNotFound     = errors.New("project not found")
	ErrProjectExists       = errors.New("project already exists")
	ErrUserNotFound        = errors.New("user not found")
	ErrUserExists          = errors.New("user already exists")
	ErrRoleNotFound        = errors.New("role not found")
	ErrUserGroupNotFound   = errors.New("user group not found")
	ErrDatabaseUnavailable = errors.New("database unavailable")
)
