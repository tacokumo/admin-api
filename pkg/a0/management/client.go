package management

import (
	"context"

	managementapi "github.com/auth0/go-auth0/management"
	"github.com/cockroachdb/errors"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
)

type Client interface {
	UserClient
}

type DefaultClient struct {
	management *managementapi.Management
}

func NewDefaultClient(ctx context.Context, domain, clientID, clientSecret string) (*DefaultClient, error) {
	m, err := managementapi.New(domain, managementapi.WithClientCredentials(ctx, clientID, clientSecret))
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &DefaultClient{
		management: m,
	}, nil
}

// GetUser implements UserClient.
func (d *DefaultClient) GetUser(ctx context.Context, id string, opts ...managementapi.RequestOption) (*managementapi.User, error) {
	ctx, span := otel.Tracer("pkg/a0/management").Start(ctx, "DefaultClient.GetUser", trace.WithAttributes(attribute.String("id", id)))
	defer span.End()

	u, err := d.management.User.Read(ctx, id, opts...)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, errors.WithStack(err)
	}
	return u, nil
}

// ListUsers implements Client.
func (d *DefaultClient) ListUsers(ctx context.Context, opts ...managementapi.RequestOption) (*managementapi.UserList, error) {
	ctx, span := otel.Tracer("pkg/a0/management").Start(ctx, "DefaultClient.ListUsers")
	defer span.End()

	ul, err := d.management.User.List(ctx, opts...)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, errors.WithStack(err)
	}
	return ul, nil
}

// GetUserPermissions implements UserClient.
func (d *DefaultClient) GetUserPermissions(ctx context.Context, id string, opts ...managementapi.RequestOption) (*managementapi.PermissionList, error) {
	ctx, span := otel.Tracer("pkg/a0/management").Start(ctx, "DefaultClient.GetUserPermissions", trace.WithAttributes(attribute.String("id", id)))
	defer span.End()

	pl, err := d.management.User.Permissions(ctx, id, opts...)
	if err != nil {
		span.SetStatus(codes.Error, err.Error())
		return nil, errors.WithStack(err)
	}
	return pl, nil
}

var _ Client = &DefaultClient{}
