package v1alpha1

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/cockroachdb/errors"
	"github.com/samber/lo"
	adminv1alpha1 "github.com/tacokumo/admin-api/pkg/apis/v1alpha1/generated"
	"github.com/tacokumo/admin-api/pkg/apis/v1alpha1/validator"
	"github.com/tacokumo/admin-api/pkg/db/admindb"
)

type Service struct {
	logger  *slog.Logger
	queries *admindb.Queries
}

// CreateRole implements generated.Handler.
func (s *Service) CreateRole(ctx context.Context, req *adminv1alpha1.CreateRoleRequest, params adminv1alpha1.CreateRoleParams) (adminv1alpha1.CreateRoleRes, error) {
	panic("unimplemented")
}

// CreateUser implements generated.Handler.
func (s *Service) CreateUser(ctx context.Context, req *adminv1alpha1.CreateUserRequest) (adminv1alpha1.CreateUserRes, error) {
	panic("unimplemented")
}

// CreateUserGroup implements generated.Handler.
func (s *Service) CreateUserGroup(ctx context.Context, req *adminv1alpha1.CreateUserGroupRequest, params adminv1alpha1.CreateUserGroupParams) (adminv1alpha1.CreateUserGroupRes, error) {
	panic("unimplemented")
}

// GetProject implements generated.Handler.
func (s *Service) GetProject(ctx context.Context, params adminv1alpha1.GetProjectParams) (adminv1alpha1.GetProjectRes, error) {
	panic("unimplemented")
}

// GetRole implements generated.Handler.
func (s *Service) GetRole(ctx context.Context, params adminv1alpha1.GetRoleParams) (adminv1alpha1.GetRoleRes, error) {
	panic("unimplemented")
}

// GetUserGroup implements generated.Handler.
func (s *Service) GetUserGroup(ctx context.Context, params adminv1alpha1.GetUserGroupParams) (adminv1alpha1.GetUserGroupRes, error) {
	panic("unimplemented")
}

// ListRoles implements generated.Handler.
func (s *Service) ListRoles(ctx context.Context, params adminv1alpha1.ListRolesParams) (adminv1alpha1.ListRolesRes, error) {
	panic("unimplemented")
}

// ListUserGroups implements generated.Handler.
func (s *Service) ListUserGroups(ctx context.Context, params adminv1alpha1.ListUserGroupsParams) (adminv1alpha1.ListUserGroupsRes, error) {
	panic("unimplemented")
}

// ListUsers implements generated.Handler.
func (s *Service) ListUsers(ctx context.Context, params adminv1alpha1.ListUsersParams) (adminv1alpha1.ListUsersRes, error) {
	panic("unimplemented")
}

// UpdateProject implements generated.Handler.
func (s *Service) UpdateProject(ctx context.Context, req *adminv1alpha1.UpdateProjectRequest, params adminv1alpha1.UpdateProjectParams) (adminv1alpha1.UpdateProjectRes, error) {
	panic("unimplemented")
}

// UpdateRole implements generated.Handler.
func (s *Service) UpdateRole(ctx context.Context, req *adminv1alpha1.UpdateRoleRequest, params adminv1alpha1.UpdateRoleParams) (adminv1alpha1.UpdateRoleRes, error) {
	panic("unimplemented")
}

// UpdateUserGroup implements generated.Handler.
func (s *Service) UpdateUserGroup(ctx context.Context, req *adminv1alpha1.UpdateUserGroupRequest, params adminv1alpha1.UpdateUserGroupParams) (adminv1alpha1.UpdateUserGroupRes, error) {
	panic("unimplemented")
}

// CreateProject implements generated.Handler.
func (s *Service) CreateProject(ctx context.Context, req *adminv1alpha1.CreateProjectRequest) (adminv1alpha1.CreateProjectRes, error) {
	projObj := adminv1alpha1.Project{
		Name:        req.Name,
		Description: req.Description,
		Kind:        adminv1alpha1.ProjectKind(req.Kind),
	}
	if err := validator.PreValidateProjectCreate(ctx, s.logger, &projObj); err != nil {
		return nil, errors.WithStack(err)
	}
	createProjectParams := admindb.CreateProjectParams{
		Name:        req.Name,
		Description: req.Description,
		Kind:        string(req.Kind),
	}
	if err := s.queries.CreateProject(ctx, createProjectParams); err != nil {
		return nil, errors.WithStack(err)
	}

	proj, err := s.queries.GetProjectByName(ctx, req.Name)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &adminv1alpha1.Project{
		ID:          strconv.Itoa(int(proj.ID)),
		Name:        proj.Name,
		Description: proj.Description,
		CreatedAt:   proj.CreatedAt.Time,
		UpdatedAt:   proj.UpdatedAt.Time,
	}, nil
}

// ListProjects implements generated.Handler.
func (s *Service) ListProjects(ctx context.Context, params adminv1alpha1.ListProjectsParams) (adminv1alpha1.ListProjectsRes, error) {
	canReadPersonalOnly, err := validator.IsOnlyPermitedToReadPersonalProjects(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	readableProjectNames, err := validator.CollectPermittedProjectNamesToRead(ctx)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	projects := adminv1alpha1.ListProjectsOKApplicationJSON{}
	if canReadPersonalOnly {
		proj, err := s.queries.GetOwnedPersonalProject(ctx)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		projects = append(projects, adminv1alpha1.Project{
			ID:          strconv.Itoa(int(proj.ID)),
			Name:        proj.Name,
			Description: proj.Description,
			CreatedAt:   proj.CreatedAt.Time,
			UpdatedAt:   proj.UpdatedAt.Time,
		})
	} else {

		listProjectsParams := admindb.ListProjectsWithPaginationParams{
			Limit:  int32(params.Limit),
			Offset: int32(params.Offset),
			Names:  readableProjectNames,
		}
		results, err := s.queries.ListProjectsWithPagination(ctx, listProjectsParams)
		if err != nil {
			return nil, errors.WithStack(err)
		}
		projects = lo.Map(results, func(p admindb.TacokumoAdminProject, _ int) adminv1alpha1.Project {
			return adminv1alpha1.Project{
				ID:          strconv.Itoa(int(p.ID)),
				Name:        p.Name,
				Description: p.Description,
				CreatedAt:   p.CreatedAt.Time,
				UpdatedAt:   p.UpdatedAt.Time,
			}
		})
	}
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &projects, nil
}

// GetReadinessCheck implements generated.Handler.
func (s *Service) GetReadinessCheck(ctx context.Context) (*adminv1alpha1.HealthResponse, error) {
	_, err := s.queries.CheckDBConnection(ctx)
	if err != nil {
		return &adminv1alpha1.HealthResponse{
			Status: "ng",
		}, errors.WithStack(err)
	}
	return &adminv1alpha1.HealthResponse{
		Status: "ok",
	}, nil
}

// GetLivenessCheck implements generated.Handler.
func (s *Service) GetLivenessCheck(ctx context.Context) (*adminv1alpha1.HealthResponse, error) {
	return &adminv1alpha1.HealthResponse{
		Status: "ok",
	}, nil
}

func NewService(
	logger *slog.Logger,
	queries *admindb.Queries,
) *Service {
	return &Service{
		logger:  logger,
		queries: queries,
	}
}

var _ adminv1alpha1.Handler = &Service{}
