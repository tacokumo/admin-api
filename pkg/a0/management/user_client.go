package management

import (
	"context"

	managementapi "github.com/auth0/go-auth0/management"
)

type UserClient interface {
	GetUser(ctx context.Context, id string, opts ...managementapi.RequestOption) (*managementapi.User, error)
	GetUserPermissions(ctx context.Context, id string, opts ...managementapi.RequestOption) (*managementapi.PermissionList, error)
	ListUsers(ctx context.Context, opts ...managementapi.RequestOption) (*managementapi.UserList, error)
}
