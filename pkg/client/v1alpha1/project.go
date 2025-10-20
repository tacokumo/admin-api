package v1alpha1

import (
	"context"
	"net/http"

	"github.com/cockroachdb/errors"
	"github.com/tacokumo/admin-api/pkg/apis/v1alpha1/generated"
)

func (c *DefaultClient) CreateProject(ctx context.Context, req *generated.CreateProjectRequest) (err error) {
	resp, err := c.post(ctx, "/v1alpha1/projects", req)
	if err != nil {
		return errors.WithStack(err)
	}
	defer func() {
		err = resp.Body.Close()
	}()

	if resp.StatusCode != http.StatusCreated {
		return errors.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return nil
}
