package v1alpha1

import (
	"context"
	"log/slog"

	"github.com/cockroachdb/errors"
	"github.com/jackc/pgx/v5/pgtype"
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
	projectId := pgtype.UUID{}
	if err := projectId.Scan(params.ProjectId); err != nil {
		return nil, errors.WithStack(err)
	}
	proj, err := s.queries.GetProjectByDisplayID(ctx, projectId)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	err = s.queries.CreateRole(ctx, admindb.CreateRoleParams{
		ProjectID:   proj.ID,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &adminv1alpha1.Role{
		Name:        req.Name,
		Description: req.Description,
		Project: adminv1alpha1.Project{
			ID:          proj.DisplayID.String(),
			Name:        proj.Name,
			Description: proj.Description,
			Kind:        adminv1alpha1.ProjectKind(proj.Kind),
			CreatedAt:   proj.CreatedAt.Time,
			UpdatedAt:   proj.UpdatedAt.Time,
		},
	}, nil
}

// CreateUser implements generated.Handler.
func (s *Service) CreateUser(ctx context.Context, req *adminv1alpha1.CreateUserRequest) (adminv1alpha1.CreateUserRes, error) {
	if err := s.queries.CreateUser(ctx, req.Email); err != nil {
		return nil, errors.WithStack(err)
	}

	user, err := s.queries.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &adminv1alpha1.User{
		ID:        user.DisplayID.String(),
		Email:     user.Email,
		Roles:     []adminv1alpha1.Role{},
		CreatedAt: user.CreatedAt.Time,
		UpdatedAt: user.UpdatedAt.Time,
	}, nil
}

// CreateUserGroup implements generated.Handler.
func (s *Service) CreateUserGroup(ctx context.Context, req *adminv1alpha1.CreateUserGroupRequest, params adminv1alpha1.CreateUserGroupParams) (adminv1alpha1.CreateUserGroupRes, error) {
	projectId := pgtype.UUID{}
	if err := projectId.Scan(params.ProjectId); err != nil {
		return nil, errors.WithStack(err)
	}
	proj, err := s.queries.GetProjectByDisplayID(ctx, projectId)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	err = s.queries.CreateUserGroup(ctx, admindb.CreateUserGroupParams{
		ProjectID:   proj.ID,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &adminv1alpha1.UserGroup{
		Name:        req.Name,
		Description: req.Description,
		Project: adminv1alpha1.Project{
			ID:          proj.DisplayID.String(),
			Name:        proj.Name,
			Description: proj.Description,
			Kind:        adminv1alpha1.ProjectKind(proj.Kind),
			CreatedAt:   proj.CreatedAt.Time,
			UpdatedAt:   proj.UpdatedAt.Time,
		},
		Members: []adminv1alpha1.User{},
	}, nil
}

// GetProject implements generated.Handler.
func (s *Service) GetProject(ctx context.Context, params adminv1alpha1.GetProjectParams) (adminv1alpha1.GetProjectRes, error) {
	displayId := pgtype.UUID{}
	if err := displayId.Scan(params.ProjectId); err != nil {
		return nil, errors.WithStack(err)
	}

	proj, err := s.queries.GetProjectByDisplayID(ctx, displayId)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &adminv1alpha1.Project{
		ID:          proj.DisplayID.String(),
		Name:        proj.Name,
		Description: proj.Description,
		Kind:        adminv1alpha1.ProjectKind(proj.Kind),
		CreatedAt:   proj.CreatedAt.Time,
		UpdatedAt:   proj.UpdatedAt.Time,
	}, nil
}

// GetRole implements generated.Handler.
func (s *Service) GetRole(ctx context.Context, params adminv1alpha1.GetRoleParams) (adminv1alpha1.GetRoleRes, error) {
	projectId := pgtype.UUID{}
	if err := projectId.Scan(params.ProjectId); err != nil {
		return nil, errors.WithStack(err)
	}

	proj, err := s.queries.GetProjectByDisplayID(ctx, projectId)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	roleId := pgtype.UUID{}
	if err := roleId.Scan(params.RoleId); err != nil {
		return nil, errors.WithStack(err)
	}
	queryArgs := admindb.GetRoleByDisplayIDParams{
		ProjectID: proj.ID,
		DisplayID: roleId,
	}
	role, err := s.queries.GetRoleByDisplayID(ctx, queryArgs)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &adminv1alpha1.Role{
		ID:          role.DisplayID.String(),
		Name:        role.Name,
		Description: role.Description,
		Project: adminv1alpha1.Project{
			ID:          proj.DisplayID.String(),
			Name:        proj.Name,
			Description: proj.Description,
			Kind:        adminv1alpha1.ProjectKind(proj.Kind),
			CreatedAt:   proj.CreatedAt.Time,
			UpdatedAt:   proj.UpdatedAt.Time,
		},
		CreatedAt: role.CreatedAt.Time,
		UpdatedAt: role.UpdatedAt.Time,
	}, nil
}

// GetUserGroup implements generated.Handler.
func (s *Service) GetUserGroup(ctx context.Context, params adminv1alpha1.GetUserGroupParams) (adminv1alpha1.GetUserGroupRes, error) {
	projectId := pgtype.UUID{}
	if err := projectId.Scan(params.ProjectId); err != nil {
		return nil, errors.WithStack(err)
	}

	proj, err := s.queries.GetProjectByDisplayID(ctx, projectId)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	userGroupId := pgtype.UUID{}
	if err := userGroupId.Scan(params.GroupId); err != nil {
		return nil, errors.WithStack(err)
	}
	queryArgs := admindb.GetUserGroupByDisplayIDParams{
		ProjectID: proj.ID,
		DisplayID: userGroupId,
	}
	userGroup, err := s.queries.GetUserGroupByDisplayID(ctx, queryArgs)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	userRecords, err := s.queries.ListUserGroupMembers(ctx, userGroup.ID)
	if err != nil {
		return nil, errors.WithStack(err)
	}

	members := lo.Map(userRecords, func(user admindb.TacokumoAdminUser, _ int) adminv1alpha1.User {
		return adminv1alpha1.User{
			ID:        user.DisplayID.String(),
			Email:     user.Email,
			Roles:     []adminv1alpha1.Role{},
			CreatedAt: user.CreatedAt.Time,
			UpdatedAt: user.UpdatedAt.Time,
		}
	})

	return &adminv1alpha1.UserGroup{
		ID:          userGroup.DisplayID.String(),
		Name:        userGroup.Name,
		Description: userGroup.Description,
		Project: adminv1alpha1.Project{
			ID:          proj.DisplayID.String(),
			Name:        proj.Name,
			Description: proj.Description,
			Kind:        adminv1alpha1.ProjectKind(proj.Kind),
			CreatedAt:   proj.CreatedAt.Time,
			UpdatedAt:   proj.UpdatedAt.Time,
		},
		Members:   members,
		CreatedAt: userGroup.CreatedAt.Time,
		UpdatedAt: userGroup.UpdatedAt.Time,
	}, nil
}

// ListRoles implements generated.Handler.
func (s *Service) ListRoles(ctx context.Context, params adminv1alpha1.ListRolesParams) (adminv1alpha1.ListRolesRes, error) {
	roleRecords, err := s.queries.ListRolesWithPagination(ctx, admindb.ListRolesWithPaginationParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	roles := adminv1alpha1.ListRolesOKApplicationJSON(lo.Map(roleRecords, func(role admindb.TacokumoAdminRole, _ int) adminv1alpha1.Role {
		return adminv1alpha1.Role{
			ID:          role.DisplayID.String(),
			Name:        role.Name,
			Description: role.Description,
			CreatedAt:   role.CreatedAt.Time,
			UpdatedAt:   role.UpdatedAt.Time,
		}
	}))

	return &roles, nil
}

// ListUserGroups implements generated.Handler.
func (s *Service) ListUserGroups(ctx context.Context, params adminv1alpha1.ListUserGroupsParams) (adminv1alpha1.ListUserGroupsRes, error) {
	userGroupRecords, err := s.queries.ListUserGroupsWithPagination(ctx, admindb.ListUserGroupsWithPaginationParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	userGroups := adminv1alpha1.ListUserGroupsOKApplicationJSON(lo.Map(userGroupRecords, func(ug admindb.TacokumoAdminUsergroup, _ int) adminv1alpha1.UserGroup {
		return adminv1alpha1.UserGroup{
			ID:          ug.DisplayID.String(),
			Name:        ug.Name,
			Description: ug.Description,
			CreatedAt:   ug.CreatedAt.Time,
			UpdatedAt:   ug.UpdatedAt.Time,
		}
	}))

	return &userGroups, nil
}

// ListUsers implements generated.Handler.
func (s *Service) ListUsers(ctx context.Context, params adminv1alpha1.ListUsersParams) (adminv1alpha1.ListUsersRes, error) {
	userRecords, err := s.queries.ListUsersWithPagination(ctx, admindb.ListUsersWithPaginationParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	users := adminv1alpha1.ListUsersOKApplicationJSON(lo.Map(userRecords, func(u admindb.TacokumoAdminUser, _ int) adminv1alpha1.User {
		return adminv1alpha1.User{
			ID:        u.DisplayID.String(),
			Email:     u.Email,
			Roles:     []adminv1alpha1.Role{},
			CreatedAt: u.CreatedAt.Time,
			UpdatedAt: u.UpdatedAt.Time,
		}
	}))

	return &users, nil
}

// UpdateProject implements generated.Handler.
func (s *Service) UpdateProject(ctx context.Context, req *adminv1alpha1.UpdateProjectRequest, params adminv1alpha1.UpdateProjectParams) (adminv1alpha1.UpdateProjectRes, error) {
	projectId := pgtype.UUID{}
	if err := projectId.Scan(params.ProjectId); err != nil {
		return nil, errors.WithStack(err)
	}
	err := s.queries.UpdateProject(ctx, admindb.UpdateProjectParams{
		DisplayID:   projectId,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	proj, err := s.queries.GetProjectByDisplayID(ctx, projectId)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &adminv1alpha1.Project{
		ID:          proj.DisplayID.String(),
		Name:        proj.Name,
		Description: proj.Description,
		Kind:        adminv1alpha1.ProjectKind(proj.Kind),
		CreatedAt:   proj.CreatedAt.Time,
		UpdatedAt:   proj.UpdatedAt.Time,
	}, nil
}

// UpdateRole implements generated.Handler.
func (s *Service) UpdateRole(ctx context.Context, req *adminv1alpha1.UpdateRoleRequest, params adminv1alpha1.UpdateRoleParams) (adminv1alpha1.UpdateRoleRes, error) {
	projectId := pgtype.UUID{}
	if err := projectId.Scan(params.ProjectId); err != nil {
		return nil, errors.WithStack(err)
	}
	project, err := s.queries.GetProjectByDisplayID(ctx, projectId)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	roleId := pgtype.UUID{}
	if err := roleId.Scan(params.RoleId); err != nil {
		return nil, errors.WithStack(err)
	}
	err = s.queries.UpdateRole(ctx, admindb.UpdateRoleParams{
		ProjectID:   project.ID,
		DisplayID:   roleId,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	role, err := s.queries.GetRoleByDisplayID(ctx, admindb.GetRoleByDisplayIDParams{
		ProjectID: project.ID,
		DisplayID: roleId,
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &adminv1alpha1.Role{
		ID:          role.DisplayID.String(),
		Name:        role.Name,
		Description: role.Description,
		Project: adminv1alpha1.Project{
			ID:          project.DisplayID.String(),
			Name:        project.Name,
			Description: project.Description,
			Kind:        adminv1alpha1.ProjectKind(project.Kind),
			CreatedAt:   project.CreatedAt.Time,
			UpdatedAt:   project.UpdatedAt.Time,
		},
		CreatedAt: role.CreatedAt.Time,
		UpdatedAt: role.UpdatedAt.Time,
	}, nil
}

// UpdateUserGroup implements generated.Handler.
func (s *Service) UpdateUserGroup(ctx context.Context, req *adminv1alpha1.UpdateUserGroupRequest, params adminv1alpha1.UpdateUserGroupParams) (adminv1alpha1.UpdateUserGroupRes, error) {
	projectId := pgtype.UUID{}
	if err := projectId.Scan(params.ProjectId); err != nil {
		return nil, errors.WithStack(err)
	}
	project, err := s.queries.GetProjectByDisplayID(ctx, projectId)
	if err != nil {
		return nil, errors.WithStack(err)
	}
	userGroupId := pgtype.UUID{}
	if err := userGroupId.Scan(params.GroupId); err != nil {
		return nil, errors.WithStack(err)
	}
	err = s.queries.UpdateUserGroup(ctx, admindb.UpdateUserGroupParams{
		ProjectID:   project.ID,
		DisplayID:   userGroupId,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	userGroup, err := s.queries.GetUserGroupByDisplayID(ctx, admindb.GetUserGroupByDisplayIDParams{
		ProjectID: project.ID,
		DisplayID: userGroupId,
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	return &adminv1alpha1.UserGroup{
		ID:          userGroup.DisplayID.String(),
		Name:        userGroup.Name,
		Description: userGroup.Description,
		Project: adminv1alpha1.Project{
			ID:          project.DisplayID.String(),
			Name:        project.Name,
			Description: project.Description,
			Kind:        adminv1alpha1.ProjectKind(project.Kind),
			CreatedAt:   project.CreatedAt.Time,
			UpdatedAt:   project.UpdatedAt.Time,
		},
		CreatedAt: userGroup.CreatedAt.Time,
		UpdatedAt: userGroup.UpdatedAt.Time,
	}, nil
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
		ID:          proj.DisplayID.String(),
		Name:        proj.Name,
		Description: proj.Description,
		Kind:        adminv1alpha1.ProjectKind(proj.Kind),
		CreatedAt:   proj.CreatedAt.Time,
		UpdatedAt:   proj.UpdatedAt.Time,
	}, nil
}

// ListProjects implements generated.Handler.
func (s *Service) ListProjects(ctx context.Context, params adminv1alpha1.ListProjectsParams) (adminv1alpha1.ListProjectsRes, error) {
	projectRecords, err := s.queries.ListProjectsWithPagination(ctx, admindb.ListProjectsWithPaginationParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	projects := adminv1alpha1.ListProjectsOKApplicationJSON(lo.Map(projectRecords, func(p admindb.TacokumoAdminProject, _ int) adminv1alpha1.Project {
		return adminv1alpha1.Project{
			ID:          p.DisplayID.String(),
			Name:        p.Name,
			Description: p.Description,
			CreatedAt:   p.CreatedAt.Time,
			UpdatedAt:   p.UpdatedAt.Time,
		}
	}))

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
