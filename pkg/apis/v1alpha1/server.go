package v1alpha1

import (
	"context"
	"log/slog"
	"net/url"
	"time"

	"github.com/cockroachdb/errors"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/samber/lo"
	adminv1alpha1 "github.com/tacokumo/admin-api/pkg/apis/v1alpha1/generated"
	"github.com/tacokumo/admin-api/pkg/auth/oauth"
	"github.com/tacokumo/admin-api/pkg/auth/session"
	"github.com/tacokumo/admin-api/pkg/db/admindb"
	"github.com/tacokumo/admin-api/pkg/middleware"
)

type Service struct {
	logger       *slog.Logger
	queries      *admindb.Queries
	githubClient *oauth.GitHubClient
	sessionStore session.Store
	stateStore   session.Store
	frontendURL  string
	sessionTTL   time.Duration
}

// CreateRole implements generated.Handler.
func (s *Service) CreateRole(ctx context.Context, req *adminv1alpha1.CreateRoleRequest, params adminv1alpha1.CreateRoleParams) (adminv1alpha1.CreateRoleRes, error) {
	projectId := pgtype.UUID{}
	if err := projectId.Scan(params.ProjectId); err != nil {
		return nil, errors.Wrapf(err, "failed to scan project id")
	}
	proj, err := s.queries.GetProjectByDisplayID(ctx, projectId)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get project by display id")
	}

	err = s.queries.CreateRole(ctx, admindb.CreateRoleParams{
		ProjectID:   proj.ID,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create role")
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
		return nil, errors.Wrapf(err, "failed to create user")
	}

	user, err := s.queries.GetUserByEmail(ctx, req.Email)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get user by email")
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
		return nil, errors.Wrapf(err, "failed to scan project id")
	}
	proj, err := s.queries.GetProjectByDisplayID(ctx, projectId)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get project by display id")
	}

	err = s.queries.CreateUserGroup(ctx, admindb.CreateUserGroupParams{
		ProjectID:   proj.ID,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to create user group")
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
		return nil, errors.Wrapf(err, "failed to scan project id")
	}

	proj, err := s.queries.GetProjectByDisplayID(ctx, displayId)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get project by display id")
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
		return nil, errors.Wrapf(err, "failed to scan project id")
	}

	proj, err := s.queries.GetProjectByDisplayID(ctx, projectId)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get project by display id")
	}

	roleId := pgtype.UUID{}
	if err := roleId.Scan(params.RoleId); err != nil {
		return nil, errors.Wrapf(err, "failed to scan role id")
	}
	queryArgs := admindb.GetRoleByDisplayIDParams{
		ProjectID: proj.ID,
		DisplayID: roleId,
	}
	role, err := s.queries.GetRoleByDisplayID(ctx, queryArgs)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get role by display id")
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
		return nil, errors.Wrapf(err, "failed to scan project id")
	}

	proj, err := s.queries.GetProjectByDisplayID(ctx, projectId)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get project by display id")
	}

	userGroupId := pgtype.UUID{}
	if err := userGroupId.Scan(params.GroupId); err != nil {
		return nil, errors.Wrapf(err, "failed to scan group id")
	}
	queryArgs := admindb.GetUserGroupByDisplayIDParams{
		ProjectID: proj.ID,
		DisplayID: userGroupId,
	}
	userGroup, err := s.queries.GetUserGroupByDisplayID(ctx, queryArgs)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get user group by display id")
	}

	userRecords, err := s.queries.ListUserGroupMembers(ctx, userGroup.ID)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to list user group members")
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
		return nil, errors.Wrapf(err, "failed to list roles with pagination")
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
		return nil, errors.Wrapf(err, "failed to list user groups with pagination")
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
		return nil, errors.Wrapf(err, "failed to list users with pagination")
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
		return nil, errors.Wrapf(err, "failed to scan project id")
	}
	err := s.queries.UpdateProject(ctx, admindb.UpdateProjectParams{
		DisplayID:   projectId,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to update project")
	}

	proj, err := s.queries.GetProjectByDisplayID(ctx, projectId)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get project by display id")
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
		return nil, errors.Wrapf(err, "failed to scan project id")
	}
	project, err := s.queries.GetProjectByDisplayID(ctx, projectId)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get project by display id")
	}
	roleId := pgtype.UUID{}
	if err := roleId.Scan(params.RoleId); err != nil {
		return nil, errors.Wrapf(err, "failed to scan role id")
	}
	err = s.queries.UpdateRole(ctx, admindb.UpdateRoleParams{
		ProjectID:   project.ID,
		DisplayID:   roleId,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to update role")
	}

	role, err := s.queries.GetRoleByDisplayID(ctx, admindb.GetRoleByDisplayIDParams{
		ProjectID: project.ID,
		DisplayID: roleId,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get role by display id")
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
		return nil, errors.Wrapf(err, "failed to scan project id")
	}
	project, err := s.queries.GetProjectByDisplayID(ctx, projectId)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get project by display id")
	}
	userGroupId := pgtype.UUID{}
	if err := userGroupId.Scan(params.GroupId); err != nil {
		return nil, errors.Wrapf(err, "failed to scan group id")
	}
	err = s.queries.UpdateUserGroup(ctx, admindb.UpdateUserGroupParams{
		ProjectID:   project.ID,
		DisplayID:   userGroupId,
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to update user group")
	}

	userGroup, err := s.queries.GetUserGroupByDisplayID(ctx, admindb.GetUserGroupByDisplayIDParams{
		ProjectID: project.ID,
		DisplayID: userGroupId,
	})
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get user group by display id")
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
	createProjectParams := admindb.CreateProjectParams{
		Name:        req.Name,
		Description: req.Description,
		Kind:        string(req.Kind),
	}
	if err := s.queries.CreateProject(ctx, createProjectParams); err != nil {
		return nil, errors.Wrapf(err, "failed to create project")
	}

	proj, err := s.queries.GetProjectByName(ctx, req.Name)
	if err != nil {
		return nil, errors.Wrapf(err, "failed to get project by name")
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
		return nil, errors.Wrapf(err, "failed to list projects with pagination")
	}

	projects := adminv1alpha1.ListProjectsOKApplicationJSON(lo.Map(projectRecords, func(p admindb.TacokumoAdminProject, _ int) adminv1alpha1.Project {
		return adminv1alpha1.Project{
			ID:          p.DisplayID.String(),
			Name:        p.Name,
			Description: p.Description,
			Kind:        adminv1alpha1.ProjectKind(p.Kind),
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
		}, errors.Wrapf(err, "failed to check DB connection")
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
	githubClient *oauth.GitHubClient,
	sessionStore session.Store,
	stateStore session.Store,
	frontendURL string,
	sessionTTL time.Duration,
) *Service {
	return &Service{
		logger:       logger,
		queries:      queries,
		githubClient: githubClient,
		sessionStore: sessionStore,
		stateStore:   stateStore,
		frontendURL:  frontendURL,
		sessionTTL:   sessionTTL,
	}
}

// InitiateLogin implements generated.Handler.
// Note: This returns a 302 status but ogen doesn't support Location header in response.
// The actual redirect should be handled by Echo middleware or a custom handler.
func (s *Service) InitiateLogin(ctx context.Context, params adminv1alpha1.InitiateLoginParams) (adminv1alpha1.InitiateLoginRes, error) {
	// This endpoint is handled by Echo directly for proper redirect support
	// See pkg/server/server.go for the actual implementation
	return &adminv1alpha1.InitiateLoginFound{}, nil
}

// HandleOAuthCallback implements generated.Handler.
// Note: This returns a 302 status but ogen doesn't support Location header in response.
// The actual redirect should be handled by Echo middleware or a custom handler.
func (s *Service) HandleOAuthCallback(ctx context.Context, params adminv1alpha1.HandleOAuthCallbackParams) (adminv1alpha1.HandleOAuthCallbackRes, error) {
	// This endpoint is handled by Echo directly for proper redirect support
	// See pkg/server/server.go for the actual implementation
	return &adminv1alpha1.HandleOAuthCallbackFound{}, nil
}

// Logout implements generated.Handler.
func (s *Service) Logout(ctx context.Context) (adminv1alpha1.LogoutRes, error) {
	sess := middleware.GetCurrentSession(ctx)
	if sess == nil {
		return &adminv1alpha1.ErrorResponse{Error: "not authenticated"}, nil
	}

	if err := s.sessionStore.Delete(ctx, sess.ID); err != nil {
		s.logger.ErrorContext(ctx, "failed to delete session", slog.String("error", err.Error()))
	}

	return &adminv1alpha1.LogoutNoContent{}, nil
}

// GetCurrentUser implements generated.Handler.
func (s *Service) GetCurrentUser(ctx context.Context) (adminv1alpha1.GetCurrentUserRes, error) {
	sess := middleware.GetCurrentSession(ctx)
	if sess == nil {
		return &adminv1alpha1.ErrorResponse{Error: "not authenticated"}, nil
	}

	teamMemberships := lo.Map(sess.TeamMemberships, func(tm session.TeamMembership, _ int) adminv1alpha1.TeamMembership {
		return adminv1alpha1.TeamMembership{
			OrgName:  tm.OrgName,
			TeamName: tm.TeamName,
			Role:     adminv1alpha1.TeamMembershipRole(tm.Role),
		}
	})

	avatarURL, _ := url.Parse(sess.AvatarURL)

	return &adminv1alpha1.AuthenticatedUser{
		User: adminv1alpha1.GitHubUser{
			ID:        sess.UserID,
			GithubID:  sess.GitHubUserID,
			Username:  sess.GitHubUsername,
			Email:     sess.Email,
			Name:      sess.Name,
			AvatarURL: *avatarURL,
		},
		BearerToken:     sess.ID,
		TeamMemberships: teamMemberships,
	}, nil
}

// RefreshToken implements generated.Handler.
func (s *Service) RefreshToken(ctx context.Context) (adminv1alpha1.RefreshTokenRes, error) {
	sess := middleware.GetCurrentSession(ctx)
	if sess == nil {
		return &adminv1alpha1.ErrorResponse{Error: "not authenticated"}, nil
	}

	newExpiry := time.Now().Add(s.sessionTTL)
	if err := s.sessionStore.Refresh(ctx, sess.ID, newExpiry); err != nil {
		s.logger.ErrorContext(ctx, "failed to refresh session", slog.String("error", err.Error()))
		return &adminv1alpha1.ErrorResponse{Error: "failed to refresh session"}, nil
	}

	sess.ExpiresAt = newExpiry

	teamMemberships := lo.Map(sess.TeamMemberships, func(tm session.TeamMembership, _ int) adminv1alpha1.TeamMembership {
		return adminv1alpha1.TeamMembership{
			OrgName:  tm.OrgName,
			TeamName: tm.TeamName,
			Role:     adminv1alpha1.TeamMembershipRole(tm.Role),
		}
	})

	avatarURL, _ := url.Parse(sess.AvatarURL)

	return &adminv1alpha1.AuthenticatedUser{
		User: adminv1alpha1.GitHubUser{
			ID:        sess.UserID,
			GithubID:  sess.GitHubUserID,
			Username:  sess.GitHubUsername,
			Email:     sess.Email,
			Name:      sess.Name,
			AvatarURL: *avatarURL,
		},
		BearerToken:     sess.ID,
		TeamMemberships: teamMemberships,
	}, nil
}

// GetGitHubClient returns the GitHub OAuth client for use in Echo handlers.
func (s *Service) GetGitHubClient() *oauth.GitHubClient {
	return s.githubClient
}

// GetSessionStore returns the session store for use in Echo handlers.
func (s *Service) GetSessionStore() session.Store {
	return s.sessionStore
}

// GetStateStore returns the state store for use in Echo handlers.
func (s *Service) GetStateStore() session.Store {
	return s.stateStore
}

// GetFrontendURL returns the frontend URL for redirects.
func (s *Service) GetFrontendURL() string {
	return s.frontendURL
}

// GetSessionTTL returns the session TTL.
func (s *Service) GetSessionTTL() time.Duration {
	return s.sessionTTL
}

var _ adminv1alpha1.Handler = &Service{}
var _ adminv1alpha1.SecurityHandler = &Service{}

// HandleBearerAuth implements generated.SecurityHandler.
// Note: Actual session validation is done by the session middleware.
// This handler just verifies the session exists in context.
func (s *Service) HandleBearerAuth(ctx context.Context, operationName adminv1alpha1.OperationName, t adminv1alpha1.BearerAuth) (context.Context, error) {
	sess := middleware.GetCurrentSession(ctx)
	if sess == nil {
		return ctx, errors.New("not authenticated")
	}
	return ctx, nil
}

// HandleCookieAuth implements generated.SecurityHandler.
// Note: Actual session validation is done by the session middleware.
// This handler just verifies the session exists in context.
func (s *Service) HandleCookieAuth(ctx context.Context, operationName adminv1alpha1.OperationName, t adminv1alpha1.CookieAuth) (context.Context, error) {
	sess := middleware.GetCurrentSession(ctx)
	if sess == nil {
		return ctx, errors.New("not authenticated")
	}
	return ctx, nil
}
