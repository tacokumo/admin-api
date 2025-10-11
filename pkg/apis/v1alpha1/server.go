package v1alpha1

import (
	"context"
	"log/slog"
	"strconv"

	"github.com/cockroachdb/errors"
	"github.com/jackc/pgx/v5/pgtype"
	adminv1alpha1 "github.com/tacokumo/admin-api/pkg/apis/v1alpha1/generated"
	"github.com/tacokumo/admin-api/pkg/db/admindb"
)

type Service struct {
	logger  *slog.Logger
	queries *admindb.Queries
}

// CreateProject implements generated.Handler.
func (s *Service) CreateProject(ctx context.Context, req *adminv1alpha1.CreateProjectRequest) (adminv1alpha1.CreateProjectRes, error) {
	createProjectParams := admindb.CreateProjectParams{
		Name: pgtype.Text{String: req.Name, Valid: true},
	}
	if !req.Bio.Null {
		createProjectParams.Bio = pgtype.Text{String: req.Bio.Value, Valid: true}
	}
	if err := s.queries.CreateProject(ctx, createProjectParams); err != nil {
		return nil, errors.WithStack(err)
	}

	proj, err := s.queries.GetProjectByName(ctx, pgtype.Text{String: req.Name, Valid: true})
	if err != nil {
		return nil, errors.WithStack(err)
	}
	return &adminv1alpha1.Project{
		ID:        strconv.Itoa(int(proj.ID)),
		Name:      proj.Name.String,
		Bio:       adminv1alpha1.NewOptNilString(proj.Bio.String),
		CreatedAt: adminv1alpha1.NewOptDateTime(proj.CreatedAt.Time),
		UpdatedAt: adminv1alpha1.NewOptDateTime(proj.UpdatedAt.Time),
	}, nil
}

// ListProjects implements generated.Handler.
func (s *Service) ListProjects(ctx context.Context, params adminv1alpha1.ListProjectsParams) (adminv1alpha1.ListProjectsRes, error) {
	projects, err := s.queries.ListProjectsWithPagination(ctx, admindb.ListProjectsWithPaginationParams{
		Limit:  int32(params.Limit),
		Offset: int32(params.Offset),
	})
	if err != nil {
		return nil, errors.WithStack(err)
	}

	res := make(adminv1alpha1.ListProjectsOKApplicationJSON, 0, len(projects))
	for _, p := range projects {
		res = append(res, adminv1alpha1.Project{
			ID:        strconv.Itoa(int(p.ID)),
			Name:      p.Name.String,
			Bio:       adminv1alpha1.NewOptNilString(p.Bio.String),
			CreatedAt: adminv1alpha1.NewOptDateTime(p.CreatedAt.Time),
			UpdatedAt: adminv1alpha1.NewOptDateTime(p.UpdatedAt.Time),
		})
	}

	return &res, nil
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
