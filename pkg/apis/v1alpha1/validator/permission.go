package validator

import (
	"strings"

	"github.com/cockroachdb/errors"
)

type Permissions struct {
	PersonalProject PersonalProjectPermissions
	User            map[string]UserPermissions
	UserGroup       map[string]UserGroupPermissions
	Project         map[string]ProjectPermissions
	Application     map[string]ApplicationPermissions
	Role            map[string]RolePermissions
}

type PersonalProjectPermissions struct {
	CanCreate bool
	CanDelete bool
	CanRead   bool
	CanUpdate bool
}

type UserPermissions struct {
	CanCreate bool
	CanDelete bool
	CanRead   bool
	CanUpdate bool
}

type UserGroupPermissions struct {
	CanCreate bool
	CanDelete bool
	CanRead   bool
	CanUpdate bool
}

type ProjectPermissions struct {
	CanCreate bool
	CanDelete bool
	CanRead   bool
	CanUpdate bool
}

type ApplicationPermissions struct {
	CanCreate bool
	CanRead   bool
	CanUpdate bool
	CanDelete bool
}

type RolePermissions struct {
	CanCreate bool
	CanRead   bool
	CanUpdate bool
	CanDelete bool
}

func ParsePermissions(permissions []string) (Permissions, error) {
	perms := Permissions{}

	for _, p := range permissions {
		parts := strings.Split(p, ":")
		if len(parts) < 2 {
			// personal_project:create のようになっているはず
			return perms, errors.Newf("invalid permission format: %s", p)
		}

		switch parts[0] {
		case "personal_project":
			if err := parsePersonalProjectPermission(parts[1:], &perms); err != nil {
				return perms, err
			}
		case "project":
			if err := parseProjectPermission(parts[1:], &perms); err != nil {
				return perms, err
			}
		case "user":
			// Implement user permission parsing here
		case "user_group":
			// Implement user group permission parsing here
		case "application":
			// Implement application permission parsing here
		case "role":
			// Implement role permission parsing here
		default:
			return perms, errors.Newf("unknown permission type: %s", parts[0])
		}
	}

	return perms, nil
}

func parsePersonalProjectPermission(parts []string, perms *Permissions) error {
	switch parts[0] {
	case "create":
		perms.PersonalProject.CanCreate = true
		return nil
	case "delete":
		perms.PersonalProject.CanDelete = true
		return nil
	case "read":
		perms.PersonalProject.CanRead = true
		return nil
	case "update":
		perms.PersonalProject.CanUpdate = true
		return nil
	default:
		return errors.Newf("unknown personal project permission: %s", parts[0])
	}
}

func parseProjectPermission(parts []string, perms *Permissions) error {
	if len(parts) < 2 {
		return errors.Newf("invalid project permission format: %v", parts)
	}

	projectID := parts[0]
	action := parts[1]
	if perms.Project == nil {
		perms.Project = make(map[string]ProjectPermissions)
	}
	if _, exists := perms.Project[projectID]; !exists {
		perms.Project[projectID] = ProjectPermissions{}
	}

	projectPerms := perms.Project[projectID]

	switch action {
	case "create":
		projectPerms.CanCreate = true
	case "delete":
		projectPerms.CanDelete = true
	case "read":
		projectPerms.CanRead = true
	case "update":
		projectPerms.CanUpdate = true
	default:
		return errors.Newf("unknown project permission action: %s", action)
	}

	perms.Project[projectID] = projectPerms
	return nil
}
